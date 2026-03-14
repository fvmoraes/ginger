package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ginger-framework/ginger/example/internal/models"
)

// UserRepository defines data access for users.
type UserRepository interface {
	FindAll(ctx context.Context) ([]*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Save(ctx context.Context, u *models.User) (*models.User, error)
	Update(ctx context.Context, id string, input *models.UpdateUserInput) (*models.User, error)
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new userRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindAll(ctx context.Context) ([]*models.User, error) {
	// TODO: implement with real SQL
	return []*models.User{}, nil
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	// TODO: implement with real SQL
	return &models.User{ID: id, Name: "Example", Email: "example@example.com"}, nil
}

func (r *userRepository) Save(ctx context.Context, u *models.User) (*models.User, error) {
	u.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
	return u, nil
}

func (r *userRepository) Update(ctx context.Context, id string, input *models.UpdateUserInput) (*models.User, error) {
	return &models.User{
		ID:        id,
		Name:      input.Name,
		Email:     input.Email,
		UpdatedAt: time.Now(),
	}, nil
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return nil
}
