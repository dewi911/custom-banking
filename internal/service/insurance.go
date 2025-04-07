package service

import (
	"context"
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type InsuranceService struct {
	repo            InsuranceRepository
	accountRepo     AccountRepository
	transactionRepo TransactionRepository
}

func NewInsuranceService(repo InsuranceRepository, accountRepo AccountRepository, transactionRepo TransactionRepository) *InsuranceService {
	return &InsuranceService{
		repo:            repo,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
	}
}

func (s *InsuranceService) Create(ctx context.Context, request models.InsuranceRequest) (*models.Insurance, error) {
	exists, err := s.accountRepo.ExistsAccount(ctx, int(request.PaymentAccountID))
	if err != nil || !exists {
		logrus.WithError(err).Error("InsuranceService.Create: error checking account existence")
		return nil, fmt.Errorf("invalid payment account: %w", err)
	}

	userID, err := s.accountRepo.GetUserIDByAccountID(ctx, int(request.PaymentAccountID))
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error getting account user ID")
		return nil, fmt.Errorf("invalid payment account: %w", err)
	}

	if int64(userID) != request.UserID {
		return nil, errors.New("payment account does not belong to user")
	}

	amount, err := s.accountRepo.GetAccountAmount(ctx, int(request.PaymentAccountID), int(request.UserID))
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error getting account amount")
		return nil, fmt.Errorf("invalid payment account: %w", err)
	}

	if amount < request.Premium {
		return nil, errors.New("insufficient funds in payment account")
	}

	policyNumber, err := s.repo.GeneratePolicyNumber()
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error generating policy number")
		return nil, err
	}

	startDate := time.Now()
	if !request.StartDate.IsZero() {
		startDate = request.StartDate
	}

	var durationDays int64
	if request.DurationDays > 0 {
		durationDays = request.DurationDays
	} else {
		durationDays = int64(request.DurationMonths * 30)
	}

	endDate := startDate.AddDate(0, 0, int(durationDays))

	insurance := &models.Insurance{
		UserID:           request.UserID,
		Type:             request.Type,
		InsuredItem:      request.InsuredItem,
		CoverageAmount:   request.CoverageAmount,
		Premium:          request.Premium,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           models.InsuranceStatusPending,
		PolicyNumber:     policyNumber,
		Description:      request.Description,
		CurrencyID:       request.CurrencyID,
		PaymentAccountID: request.PaymentAccountID,
		DaysRemaining:    durationDays,
	}

	id, err := s.repo.Create(insurance)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error creating insurance")
		return nil, err
	}
	insurance.ID = id

	err = s.accountRepo.DepositAccount(ctx, int(request.PaymentAccountID), -request.Premium)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error updating account balance")
		return nil, fmt.Errorf("error processing payment: %w", err)
	}

	_, err = s.transactionRepo.CreateTransaction(ctx, int(request.PaymentAccountID), -1, -request.Premium)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.Create: error creating transaction")
	}

	if startDate.Before(time.Now().AddDate(0, 0, 1)) {
		err = s.repo.UpdateStatus(id, models.InsuranceStatusActive)
		if err != nil {
			logrus.WithError(err).Error("InsuranceService.Create: error updating insurance status")
		}
		insurance.Status = models.InsuranceStatusActive
	}

	return insurance, nil
}

func (s *InsuranceService) GetByID(ctx context.Context, id int64) (*models.Insurance, error) {
	insurance, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.GetByID: error getting insurance with id %d", id)
		return nil, err
	}

	return insurance, nil
}

func (s *InsuranceService) GetByUserID(ctx context.Context, userID int64) ([]*models.Insurance, error) {
	insurances, err := s.repo.GetByUserID(userID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.GetByUserID: error getting insurances for user %d", userID)
		return nil, err
	}

	return insurances, nil
}

func (s *InsuranceService) UpdateStatus(ctx context.Context, id int64, status string) error {
	insurance, err := s.repo.GetByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.UpdateStatus: error getting insurance with id %d", id)
		return err
	}

	if !isValidStatusTransition(insurance.Status, status) {
		return fmt.Errorf("invalid status transition from %s to %s", insurance.Status, status)
	}

	err = s.repo.UpdateStatus(id, status)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.UpdateStatus: error updating status for insurance %d", id)
		return err
	}

	return nil
}

func (s *InsuranceService) List(ctx context.Context, params models.InsuranceListParams) ([]*models.Insurance, int64, error) {
	insurances, totalCount, err := s.repo.List(params)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.List: error listing insurances")
		return nil, 0, err
	}

	return insurances, totalCount, nil
}

func (s *InsuranceService) CreateClaim(ctx context.Context, request models.InsuranceClaimRequest) (*models.InsuranceClaim, error) {
	insurance, err := s.repo.GetByID(request.InsuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.CreateClaim: error getting insurance with id %d", request.InsuranceID)
		return nil, err
	}

	if insurance.Status != models.InsuranceStatusActive {
		return nil, fmt.Errorf("cannot file claim for insurance with status %s", insurance.Status)
	}

	if request.ClaimDate.Before(insurance.StartDate) || request.ClaimDate.After(insurance.EndDate) {
		return nil, errors.New("claim date must be within insurance period")
	}

	claim := &models.InsuranceClaim{
		InsuranceID:   request.InsuranceID,
		ClaimDate:     request.ClaimDate,
		Description:   request.Description,
		Status:        models.ClaimStatusPending,
		Amount:        request.Amount,
		FilingDate:    time.Now(),
		DocumentLinks: request.DocumentLinks,
	}

	id, err := s.repo.CreateClaim(claim)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.CreateClaim: error creating claim")
		return nil, err
	}
	claim.ID = id

	return claim, nil
}

func (s *InsuranceService) GetClaimByID(ctx context.Context, id int64) (*models.InsuranceClaim, error) {
	claim, err := s.repo.GetClaimByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.GetClaimByID: error getting claim with id %d", id)
		return nil, err
	}

	return claim, nil
}

func (s *InsuranceService) GetClaimsByInsuranceID(ctx context.Context, insuranceID int64) ([]*models.InsuranceClaim, error) {
	claims, err := s.repo.GetClaimsByInsuranceID(insuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.GetClaimsByInsuranceID: error getting claims for insurance %d", insuranceID)
		return nil, err
	}

	return claims, nil
}

func (s *InsuranceService) UpdateClaimStatus(ctx context.Context, id int64, status string, userId int64) error {
	claim, err := s.repo.GetClaimByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.UpdateClaimStatus: error getting claim with id %d", id)
		return err
	}

	insurance, err := s.repo.GetByID(claim.InsuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.UpdateClaimStatus: error getting insurance with id %d", claim.InsuranceID)
		return err
	}

	if status != models.ClaimStatusCancelled && insurance.UserID != userId {
		return errors.New("only administrators can approve or reject claims")
	}

	if status == models.ClaimStatusCancelled && insurance.UserID != userId {
		return errors.New("only claim owner can cancel claims")
	}

	var resolutionDate time.Time
	if status == models.ClaimStatusApproved || status == models.ClaimStatusRejected {
		resolutionDate = time.Now()

		if status == models.ClaimStatusApproved {
			err = s.accountRepo.DepositAccount(ctx, int(insurance.PaymentAccountID), claim.Amount)
			if err != nil {
				logrus.WithError(err).Error("InsuranceService.UpdateClaimStatus: error updating account balance")
				return fmt.Errorf("error processing claim payment: %w", err)
			}

			_, err = s.transactionRepo.CreateTransaction(ctx, -1, int(insurance.PaymentAccountID), claim.Amount)
			if err != nil {
				logrus.WithError(err).Error("InsuranceService.UpdateClaimStatus: error creating transaction")
			}
		}
	}

	err = s.repo.UpdateClaimStatus(id, status, resolutionDate)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceService.UpdateClaimStatus: error updating status for claim %d", id)
		return err
	}

	return nil
}

func (s *InsuranceService) ListClaims(ctx context.Context, insuranceID int64, page, pageSize int64) ([]*models.InsuranceClaim, int64, error) {
	claims, totalCount, err := s.repo.ListClaims(insuranceID, page, pageSize)
	if err != nil {
		logrus.WithError(err).Error("InsuranceService.ListClaims: error listing claims")
		return nil, 0, err
	}

	return claims, totalCount, nil
}

func isValidStatusTransition(currentStatus, newStatus string) bool {
	transitions := map[string][]string{
		models.InsuranceStatusPending:   {models.InsuranceStatusActive, models.InsuranceStatusCancelled},
		models.InsuranceStatusActive:    {models.InsuranceStatusExpired, models.InsuranceStatusCancelled},
		models.InsuranceStatusExpired:   {},
		models.InsuranceStatusCancelled: {},
	}

	for _, validStatus := range transitions[currentStatus] {
		if validStatus == newStatus {
			return true
		}
	}

	return false
}
