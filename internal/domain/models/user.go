package models

import (
	"time"

	"github.com/google/uuid"
)

type UserRole string

const (
	RoleEmployee  UserRole = "employee"
	RoleModerator UserRole = "moderator"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"createdAt"`
}

// AuthRequest представляет данные для аутентификации
type AuthRequest struct {
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"role,omitempty"`
}

// TokenResponse представляет ответ с токеном
type TokenResponse struct {
	Token string `json:"token"`
}
