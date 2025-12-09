package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID
	Name      string
	SKU       string
	Quantity  int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type InventoryTransaction struct {
	ID        uuid.UUID
	ProductID uuid.UUID
	Type      TransactionType
	Quantity  int
	CreatedAt time.Time
}
