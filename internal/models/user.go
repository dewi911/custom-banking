package models

import (
	"github.com/pkg/errors"
	"time"
)

var ErrUserNotFound = errors.New("user with such credentials not found")

type User struct {
	Id        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Surname   string    `json:"surname" db:"surname"`
	Username  string    `json:"username" db:"username"`
	Email     string    `json:"email" db:"email"`
	Password  string    `json:"password" db:"password"`
	RoleID    int       `json:"role_id" db:"role_id"`
	Blocked   bool      `json:"blocked" db:"blocked"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
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

type ErrorResponse struct {
	Error string `json:"error"`
}

type NameSurname struct {
	Name    string `db:"name"`
	Surname string `db:"surname"`
}
