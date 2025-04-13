package models

import (
	"database/sql"
	"time"
)

type Transaction struct {
	ID              int            `json:"id" db:"id"`
	FromAccount     sql.NullInt64  `json:"from_account" db:"from_account"`
	ToAccount       int            `json:"to_account" db:"to_account"`
	Amount          float64        `json:"amount" db:"amount"`
	Status          string         `json:"status" db:"status"`
	DateCreated     time.Time      `json:"date_created" db:"date_created"`
	DateUpdated     time.Time      `json:"date_updated" db:"date_updated"`
	TransactionType string         `json:"transaction_type" db:"transaction_type"`
	Description     sql.NullString `json:"description" db:"description"`
	ReferenceNumber sql.NullString `json:"reference_number" db:"reference_number"`
}
