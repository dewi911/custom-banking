package rest

import (
	"context"
	"custom-banking/internal/models"
)

type UserService interface {
	SingUp(ctx context.Context, inp models.SingUpInput) error
	SingIn(ctx context.Context, inp models.SingInInput) (string, string, error)
}
