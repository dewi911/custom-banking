package service

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

type User struct {
	userRepo    UsersRepository
	sessionRepo SessionRepository
	roleRepo    RolesRepository
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

func (s *User) generateTokens(ctx context.Context, user models.User) (string, string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d:%d", user.Id, user.RoleID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
	})

	accessToken, err := t.SignedString([]byte(user.Email))
	if err != nil {
		return "", "", errors.Wrap(err, "creating and returning a complete, signed JWT token error")
	}

	refreshToken, err := newRefreshToken()
	if err != nil {
		return "", "", errors.Wrap(err, "creating and returning a complete refresh token")
	}

	if err := s.sessionRepo.Create(ctx, models.RefreshSession{
		UserID:    user.Id,
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
	}); err != nil {
		return "", "", errors.Wrap(err, "creating and returning a complete refresh token")
	}

	return accessToken, refreshToken, nil
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)

	s := rand.New(rand.NewSource(time.Now().UnixNano()))
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", errors.Wrap(err, "error reading random bytes")
	}

	return fmt.Sprintf("%x", b), nil
}
