CREATE TYPE counterparty_type AS ENUM ('client', 'supplier');

CREATE TABLE counterparties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type counterparty_type NOT NULL,
    phone_number VARCHAR(50),
    email VARCHAR(255)
);