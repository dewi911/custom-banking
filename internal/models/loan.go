package models

import (
	"database/sql"
	"time"
)

const (
	LoanStatusPending   = "pending"   // ожидает одобрения
	LoanStatusApproved  = "approved"  // одобрен
	LoanStatusActive    = "active"    // активен (используется)
	LoanStatusRejected  = "rejected"  // отклонен
	LoanStatusRepaid    = "repaid"    // полностью выплачен
	LoanStatusOverdue   = "overdue"   // просрочен
	LoanStatusCancelled = "cancelled" // отменен
)

type Loan struct {
	ID              int64     `json:"id" db:"id"`
	UserID          int64     `json:"user_id" db:"user_id"`
	Amount          float64   `json:"amount" db:"amount"`
	CurrencyID      int64     `json:"currency_id" db:"currency_id"`
	StartDate       time.Time `json:"start_date" db:"start_date"`
	EndDate         time.Time `json:"end_date" db:"end_date"`
	InterestRate    float64   `json:"interest_rate" db:"interest_rate"`
	Status          string    `json:"status" db:"status"`
	RemainingAmount float64   `json:"remaining_amount" db:"remaining_amount"`

	CurrencyCode    string       `json:"currency_code,omitempty" db:"-"`
	MonthlyPayment  float64      `json:"monthly_payment,omitempty" db:"-"`
	NextPaymentDate sql.NullTime `json:"next_payment_date,omitempty" db:"-"`
}

type LoanRequest struct {
	UserID         int64   `json:"user_id" binding:"required"`
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	CurrencyID     int64   `json:"currency_id" binding:"required"`
	MonthsDuration int     `json:"months_duration" binding:"required,min=1,max=360"` // От 1 месяца до 30 лет
	InterestRate   float64 `json:"interest_rate,omitempty"`
}

type LoanPayment struct {
	ID            int64     `json:"id" db:"id"`
	LoanID        int64     `json:"loan_id" db:"loan_id"`
	Amount        float64   `json:"amount" db:"amount"`
	Date          time.Time `json:"date" db:"date"`
	Status        string    `json:"status" db:"status"`
	PaymentMethod string    `json:"payment_method,omitempty" db:"payment_method"`
	TransactionID int64     `json:"transaction_id,omitempty" db:"transaction_id"`
}

type LoanPaymentRequest struct {
	LoanID        int64   `json:"loan_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	AccountID     int64   `json:"account_id" binding:"required"`
	PaymentMethod string  `json:"payment_method,omitempty"`
}

type LoanListParams struct {
	UserID   int64  `form:"user_id,omitempty"`
	Status   string `form:"status,omitempty"`
	Page     int64  `form:"page,default=1" binding:"min=1"`
	PageSize int64  `form:"page_size,default=20" binding:"min=1,max=100"`
}
