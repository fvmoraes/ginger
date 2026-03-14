package models

import "time"

// Order is the domain model for orders.
type Order struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateOrderInput is the payload for creating a order.
type CreateOrderInput struct {
	// TODO: add fields
}

// UpdateOrderInput is the payload for updating a order.
type UpdateOrderInput struct {
	// TODO: add fields
}
