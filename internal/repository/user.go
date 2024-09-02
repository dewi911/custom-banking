package repository

import (
	"context"
	"custom-banking/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type Users struct {
	db *sqlx.DB
}

func NewUsers(db *sqlx.DB) *Users {
	return &Users{db}
}

func (r *Users) Create(ctx context.Context, user models.User) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "create",
		"user":       user,
	}

	query := "insert into users (name, surname, username, email, password, registered_at) values ($1, $2, $3,$4, $5, now())"

	_, err := r.db.ExecContext(ctx, query, user.Name, user.Surname, user.Username, user.Email, user.Password)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution creating user query error")

		return errors.Wrap(err, "execution creating user query error")
	}

	return nil
}

func (r *Users) GetByCredentials(ctx context.Context, email, password string) (models.User, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "get",
		"email":      email,
		"password":   password,
	}

	var user models.User

	query := "select * from users where email = $1 and password = $2"

	err := r.db.QueryRowxContext(ctx, query, email, password).StructScan(&user)
	if err != nil {
		logrus.WithFields(fields).
			Error("execution get user query error")

		return models.User{}, errors.Wrap(err, "execution get user query error")
	}

	return user, nil
}

func (r *Users) GetByID(ctx context.Context, id int) (models.User, error) {
	return models.User{}, nil
}
