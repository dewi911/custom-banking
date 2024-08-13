package repository

import (
	"context"
	"custom-banking/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Role struct {
	db *sqlx.DB
}

func NewRoles(db *sqlx.DB) *Role {
	return &Role{db: db}
}

func (r *Role) GetByName(ctx context.Context, name string) (models.Role, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "role",
		"method":     "GetByName",
		"name":       name,
	}

	var role models.Role

	query := `SELECT * FROM roles WHERE name = $1`

	if err := r.db.GetContext(ctx, &role, query, name); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting name from roles query error")

		return models.Role{}, errors.Wrap(err, "execution getting name from roles query error")
	}

	return role, nil
}

func (r *Role) GetByID(ctx context.Context, id int) (models.Role, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "role",
		"method":     "GetByID",
		"name":       id,
	}

	var role models.Role

	query := `SELECT * FROM roles WHERE id = $1`

	if err := r.db.GetContext(ctx, &role, query, id); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting role by ID from roles query error")

		return models.Role{}, errors.Wrap(err, "execution getting role by ID from roles query error")
	}

	return role, nil
}
