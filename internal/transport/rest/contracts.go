package rest

import (
	"context"
	"custom-banking/internal/models"
)

type UserService interface {
	SingUp(ctx context.Context, inp models.SingUpInput) error
	SingIn(ctx context.Context, inp models.SingInInput) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	ParseToken(ctx context.Context, token string) (int, int, error)
}
