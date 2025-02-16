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
	Get(ctx context.Context, token string) (models.RefreshSession, error)
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event models.Event) error
	GetEventsList(ctx context.Context, userID int) ([]models.Event, error)
}

type UsersRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
	GetByID(ctx context.Context, id int) (models.User, error)
	BlockUser(ctx context.Context, userID int) error
	UnblockUser(ctx context.Context, userID int) error
	CheckBlockUser(ctx context.Context, userID int) (bool, error)
}
