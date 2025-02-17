package service

import (
	"context"
	"custom-banking/internal/models"
)

type RolesRepository interface {
	GetByName(ctx context.Context, name string) (models.Role, error)
}

type SessionRepository interface {
	Create(ctx context.Context, token models.RefreshSession) error
	Get(ctx context.Context, token string) (models.RefreshSession, error)
}

type RandomGenerator interface {
	GenerateRandomIban() string
	GenerateRandomCardNumber() string
	GenerateRandomCvv() string
}

type AccountRepository interface {
	GetUserIDByAccountID(ctx context.Context, accountID int) (int, error)
	GetAccountIDByIban(ctx context.Context, iban string) (int, error)
	GetAccountCurrencyIDByID(ctx context.Context, accountID int) (int, error)
	GetAccountCurrencyIDByIban(ctx context.Context, iban string) (int, error)
	GetAccountAmount(ctx context.Context, accountID, userID int) (float64, error)
	ExistsAccount(ctx context.Context, accountID int) (bool, error)
	Create(ctx context.Context, userID, currencyID int, iban string) (models.Account, error)
	GetAccountsList(ctx context.Context, userID int, paginator models.Paginator, ordering models.Orderings) ([]models.Account, error)
	GetAccount(ctx context.Context, accountID, userID int) (models.Account, error)
	DeleteAccount(ctx context.Context, accountID int) error
	DepositAccount(ctx context.Context, accountID int, amount float64) error
	TransferAccount(ctx context.Context, fromAccountID, userID int, amount float64, toAccountIban string) error
	BlockAccount(ctx context.Context, accountID, userID int) error
	UnblockAccount(ctx context.Context, accountID, userID int) error
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, fromAccountID, toAccountID int, amount float64) (models.Transaction, error)
	SetTransactionStatusToSent(ctx context.Context, transactionID int) error
	GetTransactionList(ctx context.Context, accountID int, ordering models.Orderings, paginator models.Paginator) ([]models.Transaction, error)
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event models.Event) error
	GetEventsList(ctx context.Context, userID int) ([]models.Event, error)
}

type UsersRepository interface {
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
	GetByID(ctx context.Context, id int) (models.User, error)
	BlockUser(ctx context.Context, userID int) error
	UnblockUser(ctx context.Context, userID int) error
	CheckBlockUser(ctx context.Context, userID int) (bool, error)
}
