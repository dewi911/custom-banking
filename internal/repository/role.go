package repository

import (
	"github.com/jmoiron/sqlx"
)

type Role struct {
	db *sqlx.DB
}

func NewRole(db *sqlx.DB) *Role {
	return &Role{}
}
