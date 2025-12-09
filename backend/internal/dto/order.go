package dto

import (
	"time"
)

type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

type CreateOrderRequest struct {
	CounterpartyID string             `json:"counterparty_id" binding:"required"`
	OrderType      string             `json:"order_type" binding:"required,oneof=inbound outbound"`
	OrderDate      time.Time          `json:"order_date" binding:"required"`
	Destination    *string            `json:"destination"`
	Items          []OrderItemRequest `json:"items" binding:"required"`
}

type UpdateOrderRequest struct {
	Status *string `json:"status" binding:"required,oneof=pending processing completed canceled"`
}

type OrderItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type OrderResponse struct {
	ID             string              `json:"id"`
	CounterpartyID string              `json:"counterparty_id"`
	OrderType      string              `json:"order_type"`
	Status         string              `json:"status"`
	OrderDate      time.Time           `json:"order_date"`
	Destination    string              `json:"destination"`
	CreatedAt      time.Time           `json:"created_at"`
	UpdatedAt      time.Time           `json:"updated_at"`
	Items          []OrderItemResponse `json:"items"`
}

type PagedOrders struct {
	TotalCount int             `json:"total_count"`
	PageSize   int             `json:"page_size"`
	Page       int             `json:"page"`
	Data       []OrderResponse `json:"data"`
}
