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
	eventRepo   EventRepository
}

func NewUsers(userRepo UsersRepository, sessionRepo SessionRepository, rolerepo RolesRepository, eventRepo EventRepository) *User {
	return &User{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		roleRepo:    rolerepo,
		eventRepo:   eventRepo,
	}
}

func (s *User) SingUp(ctx context.Context, inp models.SingUpInput) error {
	role, err := s.roleRepo.GetByName(ctx, "user")
	if err != nil {
		return errors.Wrap(err, "role getting error")
	}

	user := models.User{
		Name:     inp.Name,
		Surname:  inp.Surname,
		Username: inp.Username,
		Email:    inp.Email,
		Password: inp.Password,
		RoleID:   role.ID,
		Blocked:  false,
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return errors.Wrap(err, "error creating user")
	}

	return nil
}

func (s *User) SingIn(ctx context.Context, inp models.SingInInput) (string, string, error) {
	//todo password hasher

	identifier := inp.Email
	if identifier == "" {
		identifier = inp.Username
	}

	if identifier == "" {
		return "", "", errors.New("email or username required")
	}

	user, err := s.userRepo.GetByCredentials(ctx, identifier, inp.Password)
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
	secretKey := []byte("secret")

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   fmt.Sprintf("%d:%d", user.Id, user.RoleID),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
	})

	accessToken, err := t.SignedString(secretKey)
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

func (s *User) BlockUser(ctx context.Context, blockUserID, userID int) error {
	if blockUserID == userID {
		return errors.New("user cannot block himself error")
	}

	err := s.userRepo.BlockUser(ctx, blockUserID)
	if err != nil {
		return errors.Wrap(err, "error blocking user")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.UserBlockedEvent,
		Message: "User blocked successfully",
	}

	if err = s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "user blocking event blocking error")
	}

	return nil
}

func (s *User) UnblockUser(ctx context.Context, userID int) error {
	err := s.userRepo.UnblockUser(ctx, userID)
	if err != nil {
		return errors.Wrap(err, "error unblocking user")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.UserUnblockedEvent,
		Message: "User unblocked successfully",
	}

	if err = s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "user blocking event unblocking error")
	}

	return nil
}

func (s *User) CheckBlockUser(ctx context.Context, userID int) (bool, error) {
	checkBlock, err := s.userRepo.CheckBlockUser(ctx, userID)
	if err != nil {
		return false, errors.Wrap(err, "error checking block user")
	}

	return checkBlock, nil
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
