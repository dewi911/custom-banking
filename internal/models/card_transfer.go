package models

import (
	"time"
)

type CardTransferRequest struct {
	FromCardNumber string  `json:"from_card_number" binding:"required,len=16"`
	ToCardNumber   string  `json:"to_card_number" binding:"required,len=16"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	Description    string  `json:"description,omitempty"`
}

type CardInfo struct {
	ID              int64     `json:"id" db:"id"`
	CardNumber      string    `json:"card_number" db:"card_number"`
	MaskedNumber    string    `json:"masked_number,omitempty" db:"-"` // Например: **** **** **** 1234
	CardholderName  string    `json:"cardholder_name" db:"cardholder_name"`
	ExpirationDate  time.Time `json:"expiration_date" db:"expiration_date"`
	CardType        string    `json:"card_type" db:"card_type"`
	AccountID       int64     `json:"account_id" db:"account_id"`
	AvailableAmount float64   `json:"available_amount,omitempty" db:"-"`
	CurrencyCode    string    `json:"currency_code,omitempty" db:"-"`
	IsActive        bool      `json:"is_active,omitempty" db:"-"`
}

type CardTransferResult struct {
	TransactionID   int64     `json:"transaction_id"`
	Status          string    `json:"status"`
	FromCardNumber  string    `json:"from_card_number"`
	ToCardNumber    string    `json:"to_card_number"`
	Amount          float64   `json:"amount"`
	Fee             float64   `json:"fee,omitempty"`
	TotalAmount     float64   `json:"total_amount"`
	Date            time.Time `json:"date"`
	ReferenceNumber string    `json:"reference_number"`
	Description     string    `json:"description,omitempty"`
}

type CardListParams struct {
	UserID    int64  `form:"user_id,omitempty"`
	AccountID int64  `form:"account_id,omitempty"`
	CardType  string `form:"card_type,omitempty"`
	Active    *bool  `form:"active,omitempty"`
	Page      int64  `form:"page,default=1" binding:"min=1"`
	PageSize  int64  `form:"page_size,default=20" binding:"min=1,max=100"`
}

type CardBalanceRequest struct {
	CardNumber string `json:"card_number" binding:"required,len=16"`
}
