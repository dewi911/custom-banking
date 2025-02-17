package service

import (
	"context"
	"custom-banking/internal/models"
	"github.com/pkg/errors"
)

const fromAccountIDForDeposit = 0

type Account struct {
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
	eventRepo       EventRepository
	ibanGenerator   RandomGenerator
}

func NewAccount(accountRepo AccountRepository, transactionRepo TransactionRepository, eventRepo EventRepository,
	ibanGenerator RandomGenerator) *Account {
	return &Account{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		eventRepo:       eventRepo,
		ibanGenerator:   ibanGenerator,
	}
}

func (s Account) Create(ctx context.Context, userID, currencyID int) (models.Account, error) {
	iban := s.ibanGenerator.GenerateRandomIban()

	account, err := s.accountRepo.Create(ctx, userID, currencyID, iban)
	if err != nil {
		return models.Account{}, errors.Wrap(err, "account creation error")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.AccountCreatedEvent,
		Message: "new account successfully created",
		Metadata: map[string]any{
			"account_id":  account.ID,
			"currency_id": account.Currency,
			"iban":        account.Iban,
		},
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return models.Account{}, errors.Wrap(err, "account created event creation error")
	}

	return account, nil
}

func (s Account) GetAccountsList(ctx context.Context, userID int, paginator models.Paginator, ordering models.Orderings) ([]models.Account, error) {
	list, err := s.accountRepo.GetAccountsList(ctx, userID, paginator, ordering)
	if err != nil {
		return nil, errors.Wrap(err, "getting account list error")
	}

	return list, nil
}

func (s Account) GetAccount(ctx context.Context, accountID, userID int) (models.Account, error) {
	account, err := s.accountRepo.GetAccount(ctx, accountID, userID)
	if err != nil {
		return models.Account{}, errors.Wrap(err, "getting account error")
	}

	return account, nil
}

func (s Account) DeleteAccount(ctx context.Context, accountID int) error {
	err := s.accountRepo.DeleteAccount(ctx, accountID)
	if err != nil {
		return errors.Wrap(err, "account deleting error")
	}

	event := models.Event{
		UserID:  accountID,
		Type:    models.AccountDeletedEvent,
		Message: "account successfully deleted",
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "account deleted event deletion error")
	}

	return nil
}

func (s Account) DepositAccount(ctx context.Context, accountID int, amount float64) error {
	exists, err := s.accountRepo.ExistsAccount(ctx, accountID)
	if err != nil {
		return errors.Wrap(err, "account existence check error")
	}

	if !exists {
		return errors.New("account does not exist error")
	}

	transaction, err := s.transactionRepo.CreateTransaction(ctx, fromAccountIDForDeposit, accountID, amount)
	if err != nil {
		return errors.Wrap(err, "transaction created error")
	}

	err = s.accountRepo.DepositAccount(ctx, accountID, amount)
	if err != nil {
		return errors.Wrap(err, "deposit account error")
	}

	if err := s.transactionRepo.SetTransactionStatusToSent(ctx, transaction.ID); err != nil {
		return errors.Wrap(err, "set transaction status to sent error")
	}

	event := models.Event{
		Type:    models.DepositEvent,
		Message: "deposit account successful",
		Metadata: map[string]any{
			"account_id": accountID,
			"amount":     amount,
		},
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "account deposit event deposit error")
	}

	return nil
}

func (s Account) TransferAccount(ctx context.Context, fromAccountID, userID int, amount float64, toAccountIban string) error {
	fromCurrency, err := s.accountRepo.GetAccountCurrencyIDByID(ctx, fromAccountID)
	if err != nil {
		return errors.Wrap(err, "getting account currencyID by ID error")
	}

	toCurrency, err := s.accountRepo.GetAccountCurrencyIDByIban(ctx, toAccountIban)
	if err != nil {
		return errors.Wrap(err, "getting account currencyID by iban error")
	}

	if fromCurrency != toCurrency {
		return errors.New("currency matching check error")
	}

	enough, err := s.accountRepo.GetAccountAmount(ctx, fromAccountID, userID)
	if err != nil {
		return errors.Wrap(err, "getting account amount error")
	}

	if enough < amount {
		return errors.New("checking enough money in the account for the transfer error")
	}

	toAccountID, err := s.accountRepo.GetAccountIDByIban(ctx, toAccountIban)
	if err != nil {
		return errors.Wrap(err, "getting accountID by iban error")
	}

	transaction, err := s.transactionRepo.CreateTransaction(ctx, fromAccountID, toAccountID, amount)
	if err != nil {
		return errors.Wrap(err, "transaction created error")
	}

	if err := s.accountRepo.TransferAccount(ctx, fromAccountID, userID, amount, toAccountIban); err != nil {
		return errors.Wrap(err, "transfer to account error")
	}

	if err := s.transactionRepo.SetTransactionStatusToSent(ctx, transaction.ID); err != nil {
		return errors.Wrap(err, "set transaction status to sent error")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.WithdrawalEvent,
		Message: "money transfer successful",
		Metadata: map[string]any{
			"from_account_id": fromAccountID,
			"to_account_id":   toAccountID,
			"amount":          amount,
		},
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "account transfer event withdrawal error")
	}

	return nil
}

func (s Account) BlockAccount(ctx context.Context, accountID, userID int) error {
	err := s.accountRepo.BlockAccount(ctx, accountID, userID)
	if err != nil {
		return errors.Wrap(err, "blocking account error")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.AccountBlockedEvent,
		Message: "account blocked successfully",
		Metadata: map[string]any{
			"account_id": accountID,
		},
	}

	if err = s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "account blocking event blocking error")
	}

	return nil
}

func (s Account) UnblockAccount(ctx context.Context, accountID, userID int) error {
	err := s.accountRepo.UnblockAccount(ctx, accountID, userID)
	if err != nil {
		return errors.Wrap(err, "unblocking account error")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.AccountUnblockedEvent,
		Message: "account unblocked successfully",
		Metadata: map[string]any{
			"account_id": accountID,
		},
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return errors.Wrap(err, "account unblocking event unblocking error")
	}

	return nil
}
