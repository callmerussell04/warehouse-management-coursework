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
