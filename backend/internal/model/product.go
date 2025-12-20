package model

import (
	"time"

	"github.com/google/uuid"
)

type Product struct {
	ID        uuid.UUID
	Name      string
	SKU       string
	Quantity  int64
	CreatedAt time.Time
	UpdatedAt time.Time
}

type TransactionType string

const (
	TransactionIncome  TransactionType = "income"
	TransactionExpense TransactionType = "expense"
)

type InventoryTransaction struct {
	ID           uuid.UUID
	ProductID    uuid.UUID
	ProductName  string
	Type         TransactionType
	Quantity     int64
	BalanceAfter int64
	CreatedAt    time.Time
}
