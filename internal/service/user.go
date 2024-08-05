package service

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"github.com/pkg/errors"
)

type User struct {
	userRepo UsersRepository
}

func NewUsers(userRepo UsersRepository) *User {
	return &User{
		userRepo: userRepo,
	}
}

func (s *User) SingUp(ctx context.Context, inp models.SingUpInput) error {
	//todo if exist
	//todo pass hash
	//todo role user

	user := models.User{
		Name:     inp.Name,
		Surname:  inp.Surname,
		Username: inp.Username,
		Email:    inp.Email,
		Password: inp.Password,
	}

	err := s.userRepo.Create(ctx, user)
	if err != nil {
		return errors.Wrap(err, "error creating user")
	}

	return nil
}

func (s *User) SingIn(ctx context.Context, inp models.SingInInput) (string, string, error) {
	//todo password hasher

	users, err := s.userRepo.GetByCredentials(ctx, inp.Email, inp.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.Wrapf(err, "GetByCredentials error getting user by email %s", inp.Email)
		}
		return "", "", errors.Wrap(err, "error getting user by credential")
	}

	//todo acces token and refresh token

	return users.Name, "done", err
}
