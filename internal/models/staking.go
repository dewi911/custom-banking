package models

import (
	"time"
)

const (
	StakingStatusPending   = "pending"
	StakingStatusActive    = "active"
	StakingStatusMatured   = "matured"
	StakingStatusWithdrawn = "withdrawn"
	StakingStatusCancelled = "cancelled"
)

type Staking struct {
	ID           int64     `json:"id" db:"id"`
	UserID       int64     `json:"user_id" db:"user_id"`
	Amount       float64   `json:"amount" db:"amount"`
	CurrencyID   int64     `json:"currency_id" db:"currency_id"`
	StartDate    time.Time `json:"start_date" db:"start_date"`
	EndDate      time.Time `json:"end_date" db:"end_date"`
	InterestRate float64   `json:"interest_rate" db:"interest_rate"`
	Status       string    `json:"status" db:"status"`

	CurrencyCode   string  `json:"currency_code,omitempty" db:"-"`
	EarnedInterest float64 `json:"earned_interest,omitempty" db:"-"`
	TotalReturn    float64 `json:"total_return,omitempty" db:"-"`
	DaysRemaining  int     `json:"days_remaining,omitempty" db:"-"`
}

type StakingRequest struct {
	UserID       int64   `json:"user_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	CurrencyID   int64   `json:"currency_id" binding:"required"`
	DurationDays int     `json:"duration_days" binding:"required,min=1"` // days
	AccountID    int64   `json:"account_id" binding:"required"`
}

type StakingWithdrawRequest struct {
	StakingID int64 `json:"staking_id" binding:"required"`
	AccountID int64 `json:"account_id" binding:"required"`
}

type StakingListParams struct {
	UserID   int64  `form:"user_id,omitempty"`
	Status   string `form:"status,omitempty"`
	Page     int64  `form:"page,default=1" binding:"min=1"`
	PageSize int64  `form:"page_size,default=20" binding:"min=1,max=100"`
}

type StakingInterest struct {
	ID          int64     `json:"id" db:"id"`
	StakingID   int64     `json:"staking_id" db:"staking_id"`
	Amount      float64   `json:"amount" db:"amount"`
	Date        time.Time `json:"date" db:"date"`
	Description string    `json:"description,omitempty" db:"description"`
}
