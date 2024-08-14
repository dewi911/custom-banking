package service

import (
	"context"
	"custom-banking/internal/models"
)

type RolesRepository interface {
	GetByName(ctx context.Context, name string) (models.Role, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token models.RefreshSession) error
	Get(ctx context.Context, token string) (*models.RefreshSession, error)
}

type UsersRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
}
