package models

import "time"

type Card struct {
	Id                 int       `json:"id" db:"id"`
	AccountId          int       `json:"account_id" db:"account_id" json:"account_id"`
	CardNumber         int       `json:"card_number" db:"card_number"`
	CardholderName     string    `json:"cardholder_name" db:"cardholder_name"`
	ExpirationDate     time.Time `json:"expiration_date" db:"expiration_date"`
	CvvCode            string    `json:"cvv_code" db:"cvv_code"`
	CardType           string    `db:"card_type" db:"card_type"`
	CashbackPercentage float64   `db:"cashback_percentage" db:"cashback_percentage"`
}

type ListCards []Card
