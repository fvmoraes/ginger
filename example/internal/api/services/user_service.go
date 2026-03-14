package services

import (
	"context"

	"github.com/fvmoraes/ginger/example/internal/api/repositories"
	"github.com/fvmoraes/ginger/example/internal/models"
	apperrors "github.com/fvmoraes/ginger/pkg/errors"
)

// UserService defines business logic for users.
type UserService interface {
	List(ctx context.Context) ([]*models.User, error)
	Get(ctx context.Context, id string) (*models.User, error)
	Create(ctx context.Context, input *models.CreateUserInput) (*models.User, error)
	Update(ctx context.Context, id string, input *models.UpdateUserInput) (*models.User, error)
	Delete(ctx context.Context, id string) error
}

type userService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new userService.
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

func (s *userService) List(ctx context.Context) ([]*models.User, error) {
	return s.repo.FindAll(ctx)
}

func (s *userService) Get(ctx context.Context, id string) (*models.User, error) {
	if id == "" {
		return nil, apperrors.BadRequest("id is required")
	}
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, apperrors.NotFound("user not found")
	}
	return u, nil
}

func (s *userService) Create(ctx context.Context, input *models.CreateUserInput) (*models.User, error) {
	if input.Name == "" || input.Email == "" {
		return nil, apperrors.BadRequest("name and email are required")
	}
	u := &models.User{Name: input.Name, Email: input.Email}
	return s.repo.Save(ctx, u)
}

func (s *userService) Update(ctx context.Context, id string, input *models.UpdateUserInput) (*models.User, error) {
	if id == "" {
		return nil, apperrors.BadRequest("id is required")
	}
	return s.repo.Update(ctx, id, input)
}

func (s *userService) Delete(ctx context.Context, id string) error {
	if id == "" {
		return apperrors.BadRequest("id is required")
	}
	return s.repo.Delete(ctx, id)
}
