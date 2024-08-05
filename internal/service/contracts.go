package service

import (
	"context"
	"custom-banking/internal/models"
)

type UsersRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
}
