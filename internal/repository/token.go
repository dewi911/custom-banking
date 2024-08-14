package repository

import (
	"context"
	"custom-banking/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Tokens struct {
	db *sqlx.DB
}

func NewTokens(db *sqlx.DB) *Tokens {
	return &Tokens{db}
}

func (r *Tokens) Create(ctx context.Context, token models.RefreshSession) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "token",
		"method":     "Create",
		"name":       token,
	}

	_, err := r.db.ExecContext(ctx, "INSERT INTO refresh_tokens (user_id, token, expires_at) VALUES ($1, $2, $3)", token.ID, token.Token, token.ExpiresAt)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution inserting into refresh_tokens query error	")

		return errors.Wrap(err, "execution inserting into refresh_tokens query error")
	}

	return nil
}

func (r *Tokens) Get(ctx context.Context, token string) (*models.RefreshSession, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "token",
		"method":     "Create",
		"name":       token,
	}

	var refreshSession models.RefreshSession

	err := r.db.QueryRowContext(ctx, "SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1", token).Scan(&refreshSession.ID, &refreshSession.UserID, &refreshSession.Token, &refreshSession.ExpiresAt)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution get into refresh_tokens query error")

		return nil, errors.Wrap(err, "execution get into refresh_tokens query error")
	}

	_, err = r.db.ExecContext(ctx, "DELETE FROM refresh_tokens WHERE token = $1", refreshSession.UserID)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution get into refresh_tokens query error")

		return nil, errors.Wrap(err, "execution get into refresh_tokens query error")
	}

	return &refreshSession, err
}
