package model

import (
	"github.com/google/uuid"
)

type CounterpartyType string

const (
	CounterpartyClient   CounterpartyType = "client"
	CounterpartySupplier CounterpartyType = "supplier"
)

type Counterparty struct {
	ID          uuid.UUID
	Name        string
	Type        CounterpartyType
	PhoneNumber string
	Email       string
}
