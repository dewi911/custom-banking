package models

import (
	"time"
)

const (
	DefaultCashbackRate = 0.5 // 0.5% cashback on all transactions

	GroceryCashbackRate = 1.0 // 1% for grocery stores
	TravelCashbackRate  = 1.5 // 1.5% for travel expenses

	CashbackPayoutFrequency = 7 // Weekly payout

	TransactionTypeCashback = "CASHBACK"
)

const (
	CashbackStatusPending = "PENDING"
	CashbackStatusPaid    = "PAID"
	CashbackStatusFailed  = "FAILED"
)

type CashbackSettings struct {
	ID        int64     `json:"id" db:"id"`
	CardID    int64     `json:"card_id" db:"card_id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Rate      float64   `json:"rate" db:"rate"`
	MinAmount float64   `json:"min_amount" db:"min_amount"`
	MaxAmount float64   `json:"max_amount" db:"max_amount"`
	Active    bool      `json:"active" db:"active"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type CashbackTransaction struct {
	ID              int64      `json:"id" db:"id"`
	UserID          int64      `json:"user_id" db:"user_id"`
	CardID          int64      `json:"card_id" db:"card_id"`
	TransactionID   int64      `json:"transaction_id" db:"transaction_id"`
	Amount          float64    `json:"amount" db:"amount"`
	Status          string     `json:"status" db:"status"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
	ProcessedAt     *time.Time `json:"processed_at,omitempty" db:"processed_at"`
	PayoutAccountID *int64     `json:"payout_account_id,omitempty" db:"payout_account_id"`
	PayoutReference string     `json:"payout_reference,omitempty" db:"payout_reference"`
}

type CashbackSummary struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	AccountID       int64     `json:"account_id" db:"account_id"`
	PendingAmount   float64   `json:"pending_amount" db:"pending_amount"`
	LastPayoutDate  time.Time `json:"last_payout_date" db:"last_payout_date"`
	NextPayoutDate  time.Time `json:"next_payout_date" db:"next_payout_date"`
	PayoutFrequency int       `json:"payout_frequency" db:"payout_frequency"` // in days
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

type CashbackSettingsRequest struct {
	CardID                    int64   `json:"card_id" binding:"required"`
	CashbackRate              float64 `json:"cashback_rate,omitempty"`
	IsActive                  bool    `json:"is_active,omitempty"`
	MinTransactionAmount      float64 `json:"min_transaction_amount,omitempty"`
	MaxCashbackPerTransaction float64 `json:"max_cashback_per_transaction,omitempty"`
	MaxCashbackPerPeriod      float64 `json:"max_cashback_per_period,omitempty"`
}

type CashbackTransactionFilter struct {
	CardID    int64     `json:"card_id,omitempty"`
	Status    string    `json:"status,omitempty"`
	StartDate time.Time `json:"start_date,omitempty"`
	EndDate   time.Time `json:"end_date,omitempty"`
	Page      int64     `json:"page,omitempty"`
	PageSize  int64     `json:"page_size,omitempty"`
}
