package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInventoryDatabaseConstraints(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	_, err := db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, 'NEGATIVE', 'Negative', -1)", uuid.New())
	assert.Error(t, err)

	productID := uuid.New()
	const largeQuantity int64 = 1 << 40
	_, err = db.Exec("INSERT INTO products (id, sku, name, quantity) VALUES ($1, 'BIGINT', 'Big integer', $2)", productID, largeQuantity)
	require.NoError(t, err)
	var storedQuantity int64
	require.NoError(t, db.QueryRow("SELECT quantity FROM products WHERE id = $1", productID).Scan(&storedQuantity))
	assert.Equal(t, largeQuantity, storedQuantity)

	_, err = db.Exec(`INSERT INTO inventory_transactions (product_id, type, quantity, balance_after)
		VALUES ($1, 'income', 1, $2)`, productID, largeQuantity)
	require.NoError(t, err)
	_, err = db.Exec("DELETE FROM products WHERE id = $1", productID)
	assert.Error(t, err, "a product with inventory history must not be deleted")

	cpID, orderID := uuid.New(), uuid.New()
	_, err = db.Exec("INSERT INTO counterparties (id, name, type) VALUES ($1, 'Supplier', 'supplier')", cpID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO orders (id, counterparty_id, status, order_date, order_type)
		VALUES ($1, $2, 'pending', NOW(), 'inbound')`, orderID, cpID)
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO order_items (order_id, product_id, quantity) VALUES ($1, $2, 0)", orderID, productID)
	assert.Error(t, err)
}
