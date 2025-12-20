ALTER TABLE products
	ALTER COLUMN quantity TYPE BIGINT,
    ALTER COLUMN sku TYPE VARCHAR(128),
    ALTER COLUMN name TYPE VARCHAR(255),
    ADD CONSTRAINT products_quantity_nonnegative CHECK (quantity >= 0);

ALTER TABLE inventory_transactions
    DROP CONSTRAINT inventory_transactions_product_id_fkey,
    ALTER COLUMN product_id SET NOT NULL,
    ALTER COLUMN quantity TYPE BIGINT,
    ALTER COLUMN balance_after TYPE BIGINT,
    ALTER COLUMN created_at SET NOT NULL,
    ADD CONSTRAINT inventory_transactions_quantity_positive CHECK (quantity > 0),
    ADD CONSTRAINT inventory_transactions_balance_nonnegative CHECK (balance_after >= 0),
    ADD CONSTRAINT inventory_transactions_product_id_fkey
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT;

ALTER TABLE order_items
	ALTER COLUMN quantity TYPE BIGINT;
