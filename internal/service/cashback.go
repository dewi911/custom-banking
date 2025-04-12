package service

import (
	"context"
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type CashbackService struct {
	repo            CashbackRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
}

func NewCashbackService(
	repo CashbackRepository,
	accountRepo AccountRepository,
	transactionRepo TransactionRepository,
) *CashbackService {
	return &CashbackService{
		repo:            repo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *CashbackService) ConfigureCashback(ctx context.Context, userID int64, cardID int64, rate float64, minAmount float64, maxAmount float64) error {
	accountID, err := s.getAccountIDForCard(ctx, cardID)
	if err != nil {
		return err
	}

	accOwnerID, err := s.accountRepo.GetUserIDByAccountID(ctx, int(accountID))
	if err != nil {
		logrus.WithError(err).Error("CashbackService.ConfigureCashback: error getting user ID for account")
		return fmt.Errorf("error getting account owner: %w", err)
	}

	if int64(accOwnerID) != userID {
		return errors.New("card does not belong to the user")
	}

	settings, err := s.repo.GetCashbackSettingsByCardID(cardID)
	if err != nil {
		settings = &models.CashbackSettings{
			CardID:    cardID,
			UserID:    userID,
			Rate:      rate,
			MinAmount: minAmount,
			MaxAmount: maxAmount,
			Active:    true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = s.repo.CreateCashbackSettings(settings)
		if err != nil {
			logrus.WithError(err).Error("CashbackService.ConfigureCashback: error creating settings")
			return err
		}

		return nil
	}

	settings.Rate = rate
	settings.MinAmount = minAmount
	settings.MaxAmount = maxAmount
	settings.Active = true
	settings.UpdatedAt = time.Now()

	err = s.repo.UpdateCashbackSettings(settings)
	if err != nil {
		logrus.WithError(err).Error("CashbackService.ConfigureCashback: error updating settings")
		return err
	}

	return nil
}

func (s *CashbackService) GetCashbackSettings(ctx context.Context, cardID int64) (*models.CashbackSettings, error) {
	settings, err := s.repo.GetCashbackSettingsByCardID(cardID)
	if err != nil {
		logrus.WithError(err).Errorf("CashbackService.GetCashbackSettings: error getting settings for card %d", cardID)
		return nil, err
	}

	return settings, nil
}

func (s *CashbackService) DeactivateCashback(ctx context.Context, cardID int64) error {
	err := s.repo.DeactivateCashbackSettings(cardID)
	if err != nil {
		logrus.WithError(err).Errorf("CashbackService.DeactivateCashback: error deactivating cashback for card %d", cardID)
		return err
	}

	return nil
}

func (s *CashbackService) GetCashbackTransactions(ctx context.Context, userID int64, filter models.CashbackTransactionFilter) ([]models.CashbackTransaction, int64, error) {
	transactions, totalCount, err := s.repo.GetCashbackTransactions(userID, filter)
	if err != nil {
		logrus.WithError(err).Errorf("CashbackService.GetCashbackTransactions: error getting transactions for user %d", userID)
		return nil, 0, err
	}

	result := make([]models.CashbackTransaction, len(transactions))
	for i, tx := range transactions {
		result[i] = *tx
	}

	return result, totalCount, nil
}

func (s *CashbackService) GetPendingCashbackAmount(ctx context.Context, userID int64) (float64, error) {
	amount, err := s.repo.GetPendingCashbackAmount(userID)
	if err != nil {
		logrus.WithError(err).Errorf("CashbackService.GetPendingCashbackAmount: error getting pending amount for user %d", userID)
		return 0, err
	}

	return amount, nil
}

func (s *CashbackService) ProcessPayouts(ctx context.Context) error {
	summaries, err := s.repo.GetUsersPendingCashbackPayout()
	if err != nil {
		logrus.WithError(err).Error("CashbackService.ProcessPayouts: error getting summaries for payout")
		return err
	}

	for _, summary := range summaries {
		if summary.PendingAmount <= 0 {
			continue
		}

		err = s.accountRepo.DepositAccount(ctx, int(summary.AccountID), summary.PendingAmount)
		if err != nil {
			logrus.WithError(err).Errorf("CashbackService.ProcessPayouts: error depositing to account %d", summary.AccountID)
			continue
		}

		err = s.repo.UpdateCashbackTransactionsStatus(summary.UserID, models.CashbackStatusPaid)
		if err != nil {
			logrus.WithError(err).Errorf("CashbackService.ProcessPayouts: error updating transactions for user %d", summary.UserID)
			continue
		}

		now := time.Now()
		summary.PendingAmount = 0
		summary.LastPayoutDate = now
		summary.NextPayoutDate = now.AddDate(0, 0, summary.PayoutFrequency)

		err = s.repo.UpdateCashbackSummary(summary)
		if err != nil {
			logrus.WithError(err).Errorf("CashbackService.ProcessPayouts: error updating summary %d", summary.ID)
			continue
		}

		_, err = s.transactionRepo.CreateTransaction(ctx, -1, int(summary.AccountID), summary.PendingAmount)
		if err != nil {
			logrus.WithError(err).Errorf("CashbackService.ProcessPayouts: error creating transaction for user %d", summary.UserID)
		}
	}

	return nil
}

func (s *CashbackService) getAccountIDForCard(ctx context.Context, cardID int64) (int64, error) {

	return 0, fmt.Errorf("not implemented")
}
