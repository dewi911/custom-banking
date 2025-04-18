package models

import (
	"time"
)

type Card struct {
	Id                 int       `json:"id" db:"id"`
	AccountId          int       `json:"account_id" db:"account_id" json:"account_id"`
	CardNumber         int       `json:"card_number" db:"card_number"`
	CardholderName     string    `json:"cardholder_name" db:"cardholder_name"`
	ExpirationDate     time.Time `json:"expiration_date" db:"expiration_date"`
	CvvCode            string    `json:"cvv_code" db:"cvv_code"`
	CardType           *string   `json:"card_type" db:"card_type"`
	CashbackPercentage *float64  `json:"cashback_percentage" db:"cashback_percentage"`
}

type ListCards []Card

type CreateCardRequestBody struct {
	CardType           string `json:"card_type"`
	CashbackPercentage int    `json:"cashback_percentage"`
}
