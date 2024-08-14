package models

import (
	"github.com/pkg/errors"
	"time"
)

var ErrUserNotFound = errors.New("user with such credentials not found")

type User struct {
	Id           int       `json:"id"`
	Name         string    `json:"name"`
	Surname      string    `json:"surname"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	Password     string    `json:"password"`
	RoleID       int       `json:"role_id"`
	RegisteredAt time.Time `json:"registered_at" db:"registered_at"`
}

type SingUpInput struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SingInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
