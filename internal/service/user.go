package service

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"math/rand"
	"strconv"
	"strings"
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

	user, err := s.userRepo.GetByCredentials(ctx, inp.Email, inp.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", "", errors.Wrapf(err, "GetByCredentials error getting user by email %s", inp.Email)
		}
		return "", "", errors.Wrap(err, "error getting user by credential")
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return "", "", errors.Wrap(err, "error generating tokens")
	}

	//todo acces token and refresh token

	return accessToken, refreshToken, nil
}

func (s *User) ParseToken(_ context.Context, token string) (int, int, error) {
	t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte("secret"), nil
	})

	if err != nil {
		return 0, 0, errors.Wrap(err, "error parsing jwt token")
	}

	if !t.Valid {
		return 0, 0, errors.Wrapf(err, "jwt token is invalid")
	}

	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, errors.Wrapf(err, "claims is invalid")
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return 0, 0, errors.Wrapf(err, "subject is invalid")
	}

	subjectParts := strings.Split(subject, ":")
	if len(subjectParts) != 2 {
		return 0, 0, errors.Wrapf(err, "token subject content error")
	}

	userID, err := strconv.Atoi(subjectParts[0])
	if err != nil {
		return 0, 0, errors.Wrapf(err, "invalid user id %s", subjectParts[0])
	}

	roleID, err := strconv.Atoi(subjectParts[1])
	if err != nil {
		return 0, 0, errors.Wrapf(err, "invalid role id %s", subjectParts[1])
	}

	return userID, roleID, nil
}

func (s *User) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	session, err := s.sessionRepo.Get(ctx, refreshToken)
	if err != nil {
		return "", "", errors.Wrap(err, "error getting refresh session")
	}

	if session.ExpiresAt.Unix() < time.Now().Unix() {
		return "", "", errors.Wrapf(err, "session is expired")
	}

	user, err := s.userRepo.GetByID(ctx, session.UserID)
	if err != nil {
		return "", "", errors.Wrap(err, "error getting user by id")
	}

	accessToken, refreshToken, err := s.generateTokens(ctx, user)
	if err != nil {
		return "", "", errors.Wrap(err, "error generating tokens")
	}

	return accessToken, refreshToken, nil
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
