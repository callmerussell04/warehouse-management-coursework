package model

import "github.com/google/uuid"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleWorker Role = "worker"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Email        string
	PasswordHash string
	FullName     string
	Role         Role
}
