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

type CardService interface {
	CreateCard(ctx context.Context, accountID, UserID int) (models.Card, error)
	GetCardListUser(ctx context.Context, userID int) ([]models.Card, error)
	GetCardListByAccount(ctx context.Context, userID, accountID int) ([]models.Card, error)
	GetCard(ctx context.Context, cardID, accountID, userID int) (models.Card, error)
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

type LoanService interface {
	Create(request models.LoanRequest) (*models.Loan, error)
	GetByID(id int64) (*models.Loan, error)
	GetByUserID(userID int64) ([]*models.Loan, error)
	List(params models.LoanListParams) ([]*models.Loan, int, error)
	UpdateStatus(id int64, status string) error
	MakePayment(request models.LoanPaymentRequest) (*models.LoanPayment, error)
	GetPaymentsByLoanID(loanID int64) ([]*models.LoanPayment, error)
}

type StakingService interface {
	Create(request models.StakingRequest) (*models.Staking, error)
	GetByID(id int64) (*models.Staking, error)
	GetByUserID(userID int64) ([]*models.Staking, error)
	List(params models.StakingListParams) ([]*models.Staking, int64, error)
	Withdraw(request models.StakingWithdrawRequest) error
	GetEarnedInterest(stakingID int64) (float64, error)
	GetInterestsByStakingID(stakingID int64) ([]*models.StakingInterest, error)
	CalculateProjectedInterest(amount float64, days int64, interestRate float64) float64
}

type CardTransfersService interface {
	TransferBetweenCards(request models.CardTransferRequest) (*models.CardTransferResult, error)
	GetCardByNumber(cardNumber string) (*models.CardInfo, error)
	GetTransactionDetails(transactionID int64) (*models.CardTransferResult, error)
	ListCardTransactions(cardID int64, page, pageSize int64) ([]*models.CardTransferResult, error)
	GetCardsByUserID(userID int64, params models.CardListParams) ([]*models.CardInfo, int64, error)
}

type InsuranceService interface {
	Create(ctx context.Context, request models.InsuranceRequest) (*models.Insurance, error)
	GetByID(ctx context.Context, id int64) (*models.Insurance, error)
	GetByUserID(ctx context.Context, userID int64) ([]*models.Insurance, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	List(ctx context.Context, params models.InsuranceListParams) ([]*models.Insurance, int64, error)

	CreateClaim(ctx context.Context, request models.InsuranceClaimRequest) (*models.InsuranceClaim, error)
	GetClaimByID(ctx context.Context, id int64) (*models.InsuranceClaim, error)
	GetClaimsByInsuranceID(ctx context.Context, insuranceID int64) ([]*models.InsuranceClaim, error)
	UpdateClaimStatus(ctx context.Context, id int64, status string, userId int64) error
	ListClaims(ctx context.Context, insuranceID int64, page, pageSize int64) ([]*models.InsuranceClaim, int64, error)
}
