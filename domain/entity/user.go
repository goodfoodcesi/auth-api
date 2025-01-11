package entity

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleClient  Role = "client"
	RoleManager Role = "manager"
	RoleAdmin   Role = "admin"
	RoleDriver  Role = "driver"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
