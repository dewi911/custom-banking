package models

import (
	"time"
)

const (
	InvoiceStatusPending  = "PENDING"
	InvoiceStatusPaid     = "PAID"
	InvoiceStatusCanceled = "CANCELED"
	InvoiceStatusExpired  = "EXPIRED"
)

const (
	PaymentMethodCard    = "CARD"
	PaymentMethodAccount = "ACCOUNT"
)

type Invoice struct {
	ID               int64         `json:"id" db:"id"`
	InvoiceNumber    string        `json:"invoice_number" db:"invoice_number"`
	UserID           int64         `json:"user_id" db:"user_id"` // Creator/Sender ID
	RecipientID      int64         `json:"recipient_id" db:"recipient_id"`
	RecipientName    string        `json:"recipient_name" db:"recipient_name"`
	RecipientAccount string        `json:"recipient_account" db:"recipient_account"`
	TotalAmount      float64       `json:"total_amount" db:"total_amount"`
	Currency         string        `json:"currency" db:"currency"`
	Description      string        `json:"description" db:"description"`
	Status           string        `json:"status" db:"status"`
	CreatedAt        time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at" db:"updated_at"`
	DueDate          time.Time     `json:"due_date" db:"due_date"`
	PaidAt           *time.Time    `json:"paid_at,omitempty" db:"paid_at"`
	PaymentMethod    string        `json:"payment_method,omitempty" db:"payment_method"`
	PaymentCardID    *int64        `json:"payment_card_id,omitempty" db:"payment_card_id"`
	PaymentAccountID *int64        `json:"payment_account_id,omitempty" db:"payment_account_id"`
	Items            []InvoiceItem `json:"items,omitempty" db:"-"`
}

type InvoiceItem struct {
	ID          int64   `json:"id" db:"id"`
	InvoiceID   int64   `json:"invoice_id" db:"invoice_id"`
	Description string  `json:"description" db:"description"`
	Quantity    int     `json:"quantity" db:"quantity"`
	UnitPrice   float64 `json:"unit_price" db:"unit_price"`
	Amount      float64 `json:"amount" db:"amount"`
}

type InvoiceRequest struct {
	RecipientID      int64         `json:"recipient_id,omitempty"`
	RecipientName    string        `json:"recipient_name" binding:"required"`
	RecipientAccount string        `json:"recipient_account" binding:"required"`
	Currency         string        `json:"currency" binding:"required"`
	Description      string        `json:"description" binding:"required"`
	DueDate          time.Time     `json:"due_date" binding:"required"`
	Items            []InvoiceItem `json:"items" binding:"required,min=1"`
}

type InvoicePaymentRequest struct {
	InvoiceID     int64  `json:"invoice_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"`
	CardID        int64  `json:"card_id,omitempty"`
	AccountID     int64  `json:"account_id,omitempty"`
}

type InvoiceFilter struct {
	UserID      int64     `json:"user_id,omitempty"`
	RecipientID int64     `json:"recipient_id,omitempty"`
	Status      string    `json:"status,omitempty"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	Page        int64     `json:"page,omitempty"`
	PageSize    int64     `json:"page_size,omitempty"`
}
