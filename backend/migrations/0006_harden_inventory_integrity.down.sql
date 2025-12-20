ALTER TABLE order_items
	ALTER COLUMN quantity TYPE INTEGER;

ALTER TABLE inventory_transactions
    DROP CONSTRAINT inventory_transactions_product_id_fkey,
    DROP CONSTRAINT inventory_transactions_quantity_positive,
    DROP CONSTRAINT inventory_transactions_balance_nonnegative,
    ALTER COLUMN product_id DROP NOT NULL,
    ALTER COLUMN quantity TYPE INTEGER,
    ALTER COLUMN balance_after TYPE INTEGER,
    ALTER COLUMN created_at DROP NOT NULL,
    ADD CONSTRAINT inventory_transactions_product_id_fkey
        FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE;

ALTER TABLE products
	DROP CONSTRAINT products_quantity_nonnegative,
	ALTER COLUMN sku TYPE TEXT,
	ALTER COLUMN name TYPE TEXT;
