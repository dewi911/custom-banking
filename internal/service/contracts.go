package service

import (
	"context"
	"custom-banking/internal/models"
	"time"
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

type CardRepository interface {
	CreateCard(ctx context.Context, accountID int, cardNumber string, cardholderName string, cvvCode string) (models.Card, error)
	GetCardListUser(ctx context.Context, userID int) ([]models.Card, error)
	GetCardListByAccount(ctx context.Context, userID, accountID int) ([]models.Card, error)
	GetCard(ctx context.Context, id, accountID int) (models.Card, error)
}

type EventRepository interface {
	CreateEvent(ctx context.Context, event models.Event) error
	GetEventsList(ctx context.Context, userID int) ([]models.Event, error)
}

type UsersRepository interface {
	GetUserNameAndSurnameByID(ctx context.Context, userID int) (string, error)
	Create(ctx context.Context, user models.User) error
	GetByCredentials(ctx context.Context, email, password string) (models.User, error)
	GetByID(ctx context.Context, id int) (models.User, error)
	BlockUser(ctx context.Context, userID int) error
	UnblockUser(ctx context.Context, userID int) error
	CheckBlockUser(ctx context.Context, userID int) (bool, error)
}

type RoleRepository interface {
	GetByID(ctx context.Context, id int) (models.Role, error)
	GetByName(ctx context.Context, name string) (models.Role, error)
}

type LoansRepository interface {
	Create(loan *models.Loan) (int64, error)
	GetByID(id int64) (*models.Loan, error)
	GetByUserID(userID int64) ([]*models.Loan, error)
	UpdateStatus(id int64, status string) error
	UpdateRemainingAmount(id int64, amount float64) error
	List(params models.LoanListParams) ([]*models.Loan, int, error)
	CreatePayment(payment *models.LoanPayment) (int64, error)
	GetPaymentsByLoanID(loanID int64) ([]*models.LoanPayment, error)
}

type StakingRepository interface {
	Create(staking *models.Staking) (int64, error)
	GetByID(id int64) (*models.Staking, error)
	GetByUserID(userID int64) ([]*models.Staking, error)
	UpdateStatus(id int64, status string) error
	List(params models.StakingListParams) ([]*models.Staking, int, error)
	CreateInterest(interest *models.StakingInterest) (int64, error)
	GetInterestsByStakingID(stakingID int64) ([]*models.StakingInterest, error)
	CalculateEarnedInterest(stakingID int64) (float64, error)
}

type CardTransfersRepository interface {
	GetCardByNumber(cardNumber string) (*models.CardInfo, error)
	GetCardByID(id int64) (*models.CardInfo, error)
	GetAccountBalance(accountID int64) (float64, error)
	CreateCardTransfer(fromCardID, toCardID int64, amount float64, description string) (int64, error)
	GetTransactionDetails(transactionID int64) (*models.CardTransferResult, error)
	ListCardTransactions(cardID int64, limit, offset int64) ([]*models.CardTransferResult, error)
	GetCardsByUserID(userID int64, params models.CardListParams) ([]*models.CardInfo, int64, error)
}

type InsuranceRepository interface {
	Create(insurance *models.Insurance) (int64, error)
	GetByID(id int64) (*models.Insurance, error)
	GetByUserID(userID int64) ([]*models.Insurance, error)
	UpdateStatus(id int64, status string) error
	List(params models.InsuranceListParams) ([]*models.Insurance, int64, error)
	GeneratePolicyNumber() (string, error)

	CreateClaim(claim *models.InsuranceClaim) (int64, error)
	GetClaimByID(id int64) (*models.InsuranceClaim, error)
	GetClaimsByInsuranceID(insuranceID int64) ([]*models.InsuranceClaim, error)
	UpdateClaimStatus(id int64, status string, resolutionDate time.Time) error
	ListClaims(insuranceID int64, page, pageSize int64) ([]*models.InsuranceClaim, int64, error)
}

type CashbackRepository interface {
	CreateCashbackSettings(settings *models.CashbackSettings) (int64, error)
	GetCashbackSettingsByCardID(cardID int64) (*models.CashbackSettings, error)
	UpdateCashbackSettings(settings *models.CashbackSettings) error
	DeactivateCashbackSettings(cardID int64) error

	CreateCashbackTransaction(transaction *models.CashbackTransaction) (int64, error)
	GetCashbackTransactions(userID int64, filter models.CashbackTransactionFilter) ([]*models.CashbackTransaction, int64, error)
	GetCashbackTransactionsByCardID(cardID int64) ([]*models.CashbackTransaction, error)
	GetPendingCashbackAmount(userID int64) (float64, error)

	GetOrCreateCashbackSummary(userID int64, accountID int64) (*models.CashbackSummary, error)
	UpdateCashbackSummary(summary *models.CashbackSummary) error

	ProcessCashbackForTransaction(transactionID int64, cardID int64) (*models.CashbackTransaction, error)
	GetUsersPendingCashbackPayout() ([]*models.CashbackSummary, error)
	UpdateCashbackTransactionsStatus(userID int64, status string) error
}

type InvoiceRepository interface {
	CreateInvoice(invoice *models.Invoice) (int64, error)
	GetInvoiceByID(id int64) (*models.Invoice, error)
	GetInvoiceByNumber(invoiceNumber string) (*models.Invoice, error)
	UpdateInvoiceStatus(id int64, status string) error
	UpdateInvoicePaymentDetails(id int64, paymentMethod string, paidDate time.Time) error
	ListInvoices(filter models.InvoiceFilter) ([]*models.Invoice, int64, error)
	GetUserInvoices(userID int64, page, pageSize int64) ([]*models.Invoice, int64, error)
	GenerateInvoiceNumber() (string, error)

	CreateInvoiceItems(items []models.InvoiceItem) error
	GetInvoiceItems(invoiceID int64) ([]models.InvoiceItem, error)

	ProcessInvoicePayment(invoiceID int64, paymentMethod string, accountOrCardID int64) error
}
