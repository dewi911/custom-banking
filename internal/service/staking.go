package service

import (
	"context"
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"math"
	"time"
)

type StakingService struct {
	stakingRepo StakingRepository
	accountRepo AccountRepository
}

func NewStakingService(stakingRepo StakingRepository, accountRepo AccountRepository) *StakingService {
	return &StakingService{
		stakingRepo: stakingRepo,
		accountRepo: accountRepo,
	}
}

func (s *StakingService) Create(ctx context.Context, request models.StakingRequest) (*models.Staking, error) {
	if request.Amount <= 0 {
		return nil, errors.New("staking amount must be positive")
	}

	if request.UserID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if request.CurrencyID <= 0 {
		return nil, errors.New("invalid currency ID")
	}

	if request.DurationDays <= 0 {
		return nil, errors.New("staking duration must be positive")
	}

	if request.AccountID <= 0 {
		return nil, errors.New("invalid account ID")
	}

	accountAmount, err := s.accountRepo.GetAccountAmount(ctx, int(request.AccountID), int(request.UserID))
	if err != nil {
		return nil, fmt.Errorf("error getting account amount: %w", err)
	}

	if accountAmount < request.Amount {
		return nil, errors.New("insufficient funds in account")
	}

	var interestRate float64
	switch {
	case request.DurationDays <= 30:
		interestRate = 3.0 // 3% for 1-30 days
	case request.DurationDays <= 90:
		interestRate = 5.0 // 5% for 31-90 days
	case request.DurationDays <= 180:
		interestRate = 7.0 // 7% for 91-180 days
	default:
		interestRate = 10.0 // 10% for 180+ days
	}

	startDate := time.Now()
	tenDaysAgo := startDate.AddDate(0, 0, -10)
	endDate := startDate.AddDate(0, 0, request.DurationDays)

	staking := &models.Staking{
		UserID:       request.UserID,
		Amount:       request.Amount,
		CurrencyID:   request.CurrencyID,
		StartDate:    tenDaysAgo,
		EndDate:      endDate,
		InterestRate: interestRate,
		Status:       models.StakingStatusActive,
	}

	err = s.accountRepo.TransferAccount(ctx, int(request.AccountID), int(request.UserID), request.Amount, "STAKING_DEPOSIT")
	if err != nil {
		return nil, fmt.Errorf("failed to transfer amount from account: %w", err)
	}

	id, err := s.stakingRepo.Create(ctx, staking)
	if err != nil {
		s.accountRepo.DepositAccount(ctx, int(request.AccountID), request.Amount)
		return nil, fmt.Errorf("failed to create staking: %w", err)
	}

	staking.ID = id
	return staking, nil
}

func (s *StakingService) GetByID(ctx context.Context, id int64) (*models.Staking, error) {
	staking, err := s.stakingRepo.GetByID(ctx, id)
	if err != nil {
		logrus.WithError(err).Error("StakingService.GetByID: error getting staking")
		return nil, err
	}

	earnedInterest, err := s.stakingRepo.CalculateEarnedInterest(ctx, id)
	if err != nil {
		logrus.WithError(err).Warn("StakingService.GetByID: error calculating earned interest")
	}
	staking.EarnedInterest = earnedInterest
	staking.TotalReturn = staking.Amount + earnedInterest

	now := time.Now()
	if now.Before(staking.EndDate) {
		daysRemaining := staking.EndDate.Sub(now).Hours() / 24
		staking.DaysRemaining = int(math.Ceil(daysRemaining))
	} else {
		staking.DaysRemaining = 0
	}

	return staking, nil
}

func (s *StakingService) GetByUserID(ctx context.Context, userID int64) ([]*models.Staking, error) {
	stakings, err := s.stakingRepo.GetByUserID(ctx, userID)
	if err != nil {
		logrus.WithError(err).Error("StakingService.GetByUserID: error getting stakings")
		return nil, err
	}

	for _, staking := range stakings {
		earnedInterest, err := s.stakingRepo.CalculateEarnedInterest(ctx, staking.ID)
		if err != nil {
			logrus.WithError(err).Warnf("StakingService.GetByUserID: error calculating earned interest for staking %d", staking.ID)
			continue
		}
		staking.EarnedInterest = earnedInterest
		staking.TotalReturn = staking.Amount + earnedInterest

		now := time.Now()
		if now.Before(staking.EndDate) {
			daysRemaining := staking.EndDate.Sub(now).Hours() / 24
			staking.DaysRemaining = int(math.Ceil(daysRemaining))
		} else {
			staking.DaysRemaining = 0
		}
	}

	return stakings, nil
}

func (s *StakingService) List(ctx context.Context, params models.StakingListParams) ([]*models.Staking, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	err := s.stakingRepo.CalculateDailyInterests(ctx)
	if err != nil {
		logrus.WithError(err).Error("StakingService.List: error listing stakings")
		return nil, 0, fmt.Errorf("CalculateDailyInterests fail")
	}

	stakings, count, err := s.stakingRepo.List(ctx, params)
	if err != nil {
		logrus.WithError(err).Error("StakingService.List: error listing stakings")
		return nil, 0, err
	}

	for _, staking := range stakings {
		earnedInterest, err := s.stakingRepo.CalculateEarnedInterest(ctx, staking.ID)
		if err != nil {
			logrus.WithError(err).Warnf("StakingService.List: error calculating earned interest for staking %d", staking.ID)
			continue
		}
		staking.EarnedInterest = earnedInterest
		staking.TotalReturn = staking.Amount + earnedInterest

		now := time.Now()
		if now.Before(staking.EndDate) {
			daysRemaining := staking.EndDate.Sub(now).Hours() / 24
			staking.DaysRemaining = int(math.Ceil(daysRemaining))
		} else {
			staking.DaysRemaining = 0
		}
	}

	return stakings, int64(count), nil
}

func (s *StakingService) Withdraw(ctx context.Context, request models.StakingWithdrawRequest) error {
	staking, err := s.stakingRepo.GetByID(ctx, request.StakingID)
	if err != nil {
		return fmt.Errorf("staking not found: %w", err)
	}

	if staking.Status != models.StakingStatusActive {
		return fmt.Errorf("cannot withdraw from a staking with status: %s", staking.Status)
	}

	earnedInterest, err := s.stakingRepo.CalculateEarnedInterest(ctx, request.StakingID)
	if err != nil {
		logrus.WithError(err).Error("StakingService.Withdraw: error calculating earned interest")
		return fmt.Errorf("failed to calculate earned interest: %w", err)
	}

	totalAmount := staking.Amount + earnedInterest
	if time.Now().Before(staking.EndDate) {
		earnedInterest = earnedInterest * 0.5
		totalAmount = staking.Amount + earnedInterest
		logrus.Infof("Early withdrawal penalty applied for staking %d", request.StakingID)
	}

	err = s.accountRepo.DepositAccount(ctx, int(request.AccountID), totalAmount)
	if err != nil {
		return fmt.Errorf("failed to deposit funds to account: %w", err)
	}

	err = s.stakingRepo.UpdateStatus(ctx, request.StakingID, models.StakingStatusWithdrawn)
	if err != nil {
		ctx := context.Background()
		s.accountRepo.TransferAccount(ctx, int(request.AccountID), int(staking.UserID), totalAmount, "STAKING_REVERSAL")
		return fmt.Errorf("failed to update staking status: %w", err)
	}

	interest := &models.StakingInterest{
		StakingID:      request.StakingID,
		Amount:         earnedInterest,
		DateCalculated: time.Now(),
		Description:    "Final interest payment on withdrawal",
	}

	_, err = s.stakingRepo.CreateInterest(ctx, interest)
	if err != nil {
		logrus.WithError(err).Error("StakingService.Withdraw: error recording final interest")
	}

	return nil
}

func (s *StakingService) GetEarnedInterest(ctx context.Context, stakingID int64) (float64, error) {
	interest, err := s.stakingRepo.CalculateEarnedInterest(ctx, stakingID)
	if err != nil {
		logrus.WithError(err).Error("StakingService.GetEarnedInterest: error calculating interest")
		return 0, err
	}
	return interest, nil
}

func (s *StakingService) GetInterestsByStakingID(ctx context.Context, stakingID int64) ([]*models.StakingInterest, error) {
	interests, err := s.stakingRepo.GetInterestsByStakingID(ctx, stakingID)
	if err != nil {
		logrus.WithError(err).Error("StakingService.GetInterestsByStakingID: error getting interests")
		return nil, err
	}
	return interests, nil
}

func (s *StakingService) CalculateProjectedInterest(ctx context.Context, amount float64, days int64, interestRate float64) float64 {
	// I = P * r * t, where:
	// P = Principal (amount)
	// r = Rate (interestRate / 100)
	// t = Time in years (days / 365)
	return amount * (interestRate / 100) * (float64(days) / 365)
}
