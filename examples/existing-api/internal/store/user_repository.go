package store

import "example.com/existing-api/internal/domain"

func SeedUser() domain.User {
	return domain.User{ID: "1", Name: "Stored user"}
}
