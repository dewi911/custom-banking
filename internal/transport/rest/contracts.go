package rest

import (
	"context"
	"custom-banking/internal/models"
)

type AccountService interface {
	Create(ctx context.Context, userID, currencyID int) (models.Account, error)
	GetAccountsList(ctx context.Context, userID int, paginator models.Paginator, ordering models.Orderings) ([]models.Account, error)
	GetAccount(ctx context.Context, accountID, userID int) (models.Account, error)
	DeleteAccount(ctx context.Context, accountID int) error
	DepositAccount(ctx context.Context, accountID int, amount float64) error
	TransferAccount(ctx context.Context, fromAccountID, userID int, amount float64, toAccountIban string) error
	BlockAccount(ctx context.Context, accountID, userID int) error
	UnblockAccount(ctx context.Context, accountID, userID int) error
}

type UserService interface {
	SingUp(ctx context.Context, inp models.SingUpInput) error
	SingIn(ctx context.Context, inp models.SingInInput) (string, string, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, error)
	ParseToken(ctx context.Context, token string) (int, int, error)
	BlockUser(ctx context.Context, blockUserID, userID int) error
	UnblockUser(ctx context.Context, userID int) error
	CheckBlockUser(ctx context.Context, userID int) (bool, error)
}

type EventService interface {
	GetEventList(ctx context.Context, userID int) ([]models.Event, error)
}

type RoleRepository interface {
	GetByID(ctx context.Context, id int) (models.Role, error)
}

type TransactionService interface {
	GetTransactionList(ctx context.Context, accountID, userID int, ordering models.Orderings, paginator models.Paginator) ([]models.Transaction, error)
}
