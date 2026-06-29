package core

import "example.com/existing-api/internal/domain"

func FindUser(id string) domain.User {
	return domain.User{ID: id, Name: "Existing user"}
}
