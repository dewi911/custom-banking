package service

import (
	"context"
	"custom-banking/internal/models"
	"github.com/pkg/errors"
)

type Transaction struct {
	TransactionRepo TransactionRepository
	AccountRepo     AccountRepository
}

func NewTransaction(transactionRepo TransactionRepository, accountRepo AccountRepository) *Transaction {
	return &Transaction{
		TransactionRepo: transactionRepo,
		AccountRepo:     accountRepo,
	}
}

func (s Transaction) GetTransactionList(ctx context.Context, accountID, userID int, ordering models.Orderings, paginator models.Paginator) ([]models.Transaction, error) {
	checkUserID, err := s.AccountRepo.GetUserIDByAccountID(ctx, accountID)
	if err != nil {
		return nil, errors.Wrap(err, "getting userID by accountID error")
	}

	if checkUserID != userID {
		return nil, errors.New("userID matching check error")
	}

	listTransaction, err := s.TransactionRepo.GetTransactionList(ctx, accountID, ordering, paginator)
	if err != nil {
		return nil, errors.Wrap(err, "getting transaction list error")
	}

	return listTransaction, nil
}
