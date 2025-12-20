package dto

import (
	"time"
)

type OrderItemRequest struct {
	ProductID string `json:"product_id" binding:"required,uuid"`
	Quantity  int64  `json:"quantity" binding:"required,min=1"`
}

type CreateOrderRequest struct {
	CounterpartyID string             `json:"counterparty_id" binding:"required,uuid"`
	OrderType      string             `json:"order_type" binding:"required,oneof=inbound outbound"`
	OrderDate      time.Time          `json:"order_date" binding:"required"`
	Destination    *string            `json:"destination" binding:"omitempty,max=255"`
	Items          []OrderItemRequest `json:"items" binding:"required,min=1,max=100,dive"`
}

type UpdateOrderRequest struct {
	Status *string `json:"status" binding:"required,oneof=processing completed canceled"`
}

type OrderItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int64  `json:"quantity"`
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
