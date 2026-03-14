package models

import "time"

// User is the domain model for a user.
type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateUserInput is the payload for creating a user.
type CreateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateUserInput is the payload for updating a user.
type UpdateUserInput struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
