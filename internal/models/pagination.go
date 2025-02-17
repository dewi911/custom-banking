package models

type Paginator struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}
