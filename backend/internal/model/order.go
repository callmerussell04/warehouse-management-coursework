package model

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending    OrderStatus = "pending"
	StatusProcessing OrderStatus = "processing"
	StatusCompleted  OrderStatus = "completed"
	StatusCanceled   OrderStatus = "canceled"
)

type OrderType string

const (
	OrderInbound  OrderType = "inbound"
	OrderOutbound OrderType = "outbound"
)

type Order struct {
	ID             uuid.UUID
	CounterpartyID uuid.UUID
	OrderType      OrderType
	Status         OrderStatus
	OrderDate      time.Time
	Destination    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Items          []OrderItem
}

type OrderItem struct {
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Quantity  int
}
