package rest

// Request/response structs specific to the staking handler
type interestResponse struct {
	StakingID int64   `json:"staking_id"`
	Interest  float64 `json:"interest"`
}

type projectedInterestRequest struct {
	Amount       float64 `json:"amount" binding:"required"`
	Days         int     `json:"days" binding:"required"`
	InterestRate float64 `json:"interest_rate,omitempty"`
}

type projectedInterestResponse struct {
	Amount       float64 `json:"amount"`
	Days         int     `json:"days"`
	InterestRate float64 `json:"interest_rate"`
	Interest     float64 `json:"interest"`
	TotalReturn  float64 `json:"total_return"`
}

type statusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// paginatedResponse represents a paginated list response
type paginatedResponse struct {
	Data       interface{} `json:"data"`
	TotalCount int64       `json:"total_count"`
	Page       int64       `json:"page"`
	PageSize   int64       `json:"page_size"`
}
