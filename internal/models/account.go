package models

type Account struct {
	ID       int    `json:"id"`
	Iban     string `json:"iban"`
	UserID   int    `json:"user_id"`
	Currency string `json:"currency"`
	Blocked  bool   `json:"blocked"`
	Amount   int    `json:"amount"`
}

type Orderings map[string]string
