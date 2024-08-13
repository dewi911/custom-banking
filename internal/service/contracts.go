package service

import (
	"context"
	"custom-banking/internal/models"
)

type RolesRepository interface {
	GetByName(ctx context.Context, name string) (models.Role, error)
}

type UsersRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
}
