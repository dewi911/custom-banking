package models

import (
	"time"
)

const (
	InsuranceStatusActive    = "active"
	InsuranceStatusExpired   = "expired"
	InsuranceStatusCancelled = "cancelled"
	InsuranceStatusPending   = "pending"

	InsuranceTypeHealth   = "health"
	InsuranceTypeProperty = "property"
	InsuranceTypeVehicle  = "vehicle"
	InsuranceTypeTravel   = "travel"
	InsuranceTypeLife     = "life"

	ClaimStatusPending   = "pending"
	ClaimStatusApproved  = "approved"
	ClaimStatusRejected  = "rejected"
	ClaimStatusCancelled = "cancelled"
)

type Insurance struct {
	ID               int64     `json:"id" db:"id"`
	UserID           int64     `json:"user_id" db:"user_id"`
	Type             string    `json:"type" db:"type"`
	InsuredItem      string    `json:"insured_item" db:"insured_item"`
	CoverageAmount   float64   `json:"coverage_amount" db:"coverage_amount"`
	Premium          float64   `json:"premium" db:"premium"` // Monthly premium
	StartDate        time.Time `json:"start_date" db:"start_date"`
	EndDate          time.Time `json:"end_date" db:"end_date"`
	Status           string    `json:"status" db:"status"`
	PolicyNumber     string    `json:"policy_number" db:"policy_number"`
	Description      string    `json:"description" db:"description"`
	CurrencyID       int64     `json:"currency_id" db:"currency_id"`
	CurrencyCode     string    `json:"currency_code,omitempty" db:"-"`
	DaysRemaining    int64     `json:"days_remaining,omitempty" db:"-"`
	PaymentAccountID int64     `json:"payment_account_id" db:"payment_account_id"`
}

type InsuranceRequest struct {
	UserID           int64     `json:"user_id" binding:"required"`
	Type             string    `json:"type" binding:"required"`
	InsuredItem      string    `json:"insured_item" binding:"required"`
	CoverageAmount   float64   `json:"coverage_amount" binding:"required,gt=0"`
	Premium          float64   `json:"premium" binding:"required,gt=0"`
	DurationMonths   int       `json:"duration_months,omitempty"`
	DurationDays     int64     `json:"duration_days,omitempty"`
	StartDate        time.Time `json:"start_date,omitempty"`
	CurrencyID       int64     `json:"currency_id" binding:"required"`
	Description      string    `json:"description,omitempty"`
	PaymentAccountID int64     `json:"payment_account_id" binding:"required"`
}

type InsuranceListParams struct {
	UserID      int64  `form:"user_id,omitempty"`
	Type        string `form:"type,omitempty"`
	Status      string `form:"status,omitempty"`
	Page        int64  `form:"page,default=1" binding:"min=1"`
	PageSize    int64  `form:"page_size,default=20" binding:"min=1,max=100"`
	InsuredItem string `form:"insured_item,omitempty"`
}

type InsuranceClaim struct {
	ID             int64     `json:"id" db:"id"`
	InsuranceID    int64     `json:"insurance_id" db:"insurance_id"`
	ClaimDate      time.Time `json:"claim_date" db:"claim_date"`
	Description    string    `json:"description" db:"description"`
	Status         string    `json:"status" db:"status"`
	Amount         float64   `json:"amount" db:"amount"`
	FilingDate     time.Time `json:"filing_date" db:"filing_date"`
	ResolutionDate time.Time `json:"resolution_date,omitempty" db:"resolution_date"`
	DocumentLinks  string    `json:"document_links,omitempty" db:"document_links"`
}

type InsuranceClaimRequest struct {
	InsuranceID   int64     `json:"insurance_id" binding:"required"`
	Description   string    `json:"description" binding:"required"`
	Amount        float64   `json:"amount" binding:"required,gt=0"`
	ClaimDate     time.Time `json:"claim_date" binding:"required"`
	DocumentLinks string    `json:"document_links,omitempty"`
}
