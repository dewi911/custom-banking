package models

import "time"

type Card struct {
	Id              int       `json:"id"`
	AccountId       int       `json:"account_id"`
	CardNumber      int       `json:"card_number"`
	CardholderName  string    `json:"cardholder_name"`
	ExpirationMonth time.Time `json:"expiration_month"`
	CvvCode         string    `json:"cvv_code"`
}

type ListCards []Card
