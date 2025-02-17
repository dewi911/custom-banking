package repository

import (
	"context"
	"custom-banking/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const block = true
const unblock = false

type Users struct {
	db *sqlx.DB
}

func NewUsers(db *sqlx.DB) *Users {
	return &Users{db}
}

func (r Users) GetUserNameAndSurnameByID(ctx context.Context, userID int) (string, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Users",
		"method":     "GetUserNameAndSurnameByID",
		"user_id":    userID,
	}

	query := "select name, surname from users where id = $1"

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil && rows.Err() != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting name and surname from query users error")

		return "", errors.Wrap(err, "execution getting name and surname from query users error")
	}

	var nameSurname models.NameSurname
	for rows.Next() {
		if err := rows.StructScan(&nameSurname); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning row into struct error")

			return "", errors.Wrap(err, "scanning row into struct error")
		}
	}

	fullNameStr := nameSurname.Name + " " + nameSurname.Surname

	return fullNameStr, nil
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
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "GetByID",
		"id":         id,
	}

	var user models.User

	query := "select id, name, surname, username, email, password, registered_at from users where id = $1"

	err := r.db.QueryRowxContext(ctx, query, id).
		Scan(&user)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution get user query error")

		return models.User{}, errors.Wrap(err, "execution get user query error")
	}

	return user, nil
}

func (r *Users) BlockUser(ctx context.Context, userID int) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "BlockUser",
		"user":       userID,
	}

	query := "update users set blocked = $1 where id = $2"

	if _, err := r.db.ExecContext(ctx, query, block, userID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution block user query by id error")

		return errors.Wrap(err, "execution block user query by id error")
	}

	return nil
}

func (r *Users) UnblockUser(ctx context.Context, userID int) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "users",
		"method":     "UnblockUser",
		"user":       userID,
	}

	query := "update users set blocked = $1 where id = $2"

	if _, err := r.db.ExecContext(ctx, query, block, userID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution unblock user query by id error")

		return errors.Wrap(err, "execution unblock user query by id error")
	}

	return nil
}

func (r *Users) CheckBlockUser(ctx context.Context, userID int) (bool, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Users",
		"method":     "CheckBlockUser",
		"user_id":    userID,
	}

	var checkBlock bool

	query := "SELECT blocked FROM users WHERE id = $1"

	if err := r.db.QueryRowContext(ctx, query, userID).Scan(&checkBlock); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting blocked from users query error")

		return false, errors.Wrap(err, "execution getting blocked from users query error")
	}

	if checkBlock {
		return true, nil
	} else {
		return false, nil
	}
}
