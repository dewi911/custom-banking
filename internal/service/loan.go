package service

import (
	"context"
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type LoanService struct {
	loanRepo    LoansRepository
	accountRepo AccountRepository
}

func NewLoanService(loanRepo LoansRepository, accountRepo AccountRepository) *LoanService {
	return &LoanService{
		loanRepo:    loanRepo,
		accountRepo: accountRepo,
	}
}

func (s *LoanService) Create(request models.LoanRequest) (*models.Loan, error) {
	if request.Amount <= 0 {
		return nil, errors.New("loan amount must be positive")
	}

	if request.UserID <= 0 {
		return nil, errors.New("invalid user ID")
	}

	if request.CurrencyID <= 0 {
		return nil, errors.New("invalid currency ID")
	}

	if request.MonthsDuration <= 0 {
		return nil, errors.New("loan duration must be positive")
	}

	if request.InterestRate <= 0 {
		request.InterestRate = 10.0
	}

	startDate := time.Now()
	endDate := startDate.AddDate(0, int(request.MonthsDuration), 0)

	loan := &models.Loan{
		UserID:          request.UserID,
		Amount:          request.Amount,
		CurrencyID:      request.CurrencyID,
		StartDate:       startDate,
		EndDate:         endDate,
		InterestRate:    request.InterestRate,
		Status:          models.LoanStatusPending,
		RemainingAmount: request.Amount,
	}

	id, err := s.loanRepo.Create(loan)
	if err != nil {
		logrus.WithError(err).Error("LoanService.Create: error creating loan")
		return nil, fmt.Errorf("failed to create loan: %w", err)
	}

	loan.ID = id

	return loan, nil
}

func (s *LoanService) GetByID(id int64) (*models.Loan, error) {
	loan, err := s.loanRepo.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("LoanService.GetByID: error getting loan")
		return nil, err
	}
	return loan, nil
}

func (s *LoanService) GetByUserID(userID int64) ([]*models.Loan, error) {
	loans, err := s.loanRepo.GetByUserID(userID)
	if err != nil {
		logrus.WithError(err).Error("LoanService.GetByUserID: error getting loans")
		return nil, err
	}
	return loans, nil
}

func (s *LoanService) List(params models.LoanListParams) ([]*models.Loan, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}

	loans, count, err := s.loanRepo.List(params)
	if err != nil {
		logrus.WithError(err).Error("LoanService.List: error listing loans")
		return nil, 0, err
	}
	return loans, count, nil
}

func (s *LoanService) UpdateStatus(id int64, status string) error {
	validStatuses := []string{
		models.LoanStatusPending,
		models.LoanStatusApproved,
		models.LoanStatusActive,
		models.LoanStatusRejected,
		models.LoanStatusRepaid,
		models.LoanStatusOverdue,
		models.LoanStatusCancelled,
	}

	valid := false
	for _, s := range validStatuses {
		if status == s {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid loan status: %s", status)
	}

	err := s.loanRepo.UpdateStatus(id, status)
	if err != nil {
		logrus.WithError(err).Error("LoanService.UpdateStatus: error updating loan status")
		return err
	}
	return nil
}

func (s *LoanService) MakePayment(request models.LoanPaymentRequest) (*models.LoanPayment, error) {
	if request.LoanID <= 0 {
		return nil, errors.New("invalid loan ID")
	}

	if request.Amount <= 0 {
		return nil, errors.New("payment amount must be positive")
	}

	if request.AccountID <= 0 {
		return nil, errors.New("invalid account ID")
	}

	loan, err := s.loanRepo.GetByID(request.LoanID)
	if err != nil {
		return nil, fmt.Errorf("loan not found: %w", err)
	}

	if loan.Status != models.LoanStatusApproved && loan.Status != models.LoanStatusActive {
		return nil, fmt.Errorf("cannot make payment on a loan with status: %s", loan.Status)
	}

	if request.Amount > loan.RemainingAmount {
		return nil, fmt.Errorf("payment amount (%f) exceeds remaining loan amount (%f)", request.Amount, loan.RemainingAmount)
	}

	ctx := context.Background()
	accountAmount, err := s.accountRepo.GetAccountAmount(ctx, int(request.AccountID), int(loan.UserID))
	if err != nil {
		return nil, fmt.Errorf("error getting account amount: %w", err)
	}

	if accountAmount < request.Amount {
		return nil, errors.New("insufficient funds in account")
	}

	paymentMethod := request.PaymentMethod
	if paymentMethod == "" {
		paymentMethod = "account_transfer"
	}

	payment := &models.LoanPayment{
		LoanID:        request.LoanID,
		Amount:        request.Amount,
		Date:          time.Now(),
		Status:        "completed",
		PaymentMethod: paymentMethod,
	}

	ctx = context.Background()
	err = s.accountRepo.TransferAccount(ctx, int(request.AccountID), int(loan.UserID), request.Amount, "LOAN_PAYMENT")
	if err != nil {
		return nil, fmt.Errorf("failed to transfer funds from account: %w", err)
	}

	newRemainingAmount := loan.RemainingAmount - request.Amount
	err = s.loanRepo.UpdateRemainingAmount(request.LoanID, newRemainingAmount)
	if err != nil {
		ctx = context.Background()
		s.accountRepo.DepositAccount(ctx, int(request.AccountID), request.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to deposit funds from account: %w", err)
		}
		return nil, fmt.Errorf("failed to update loan remaining amount: %w", err)
	}

	if newRemainingAmount <= 0 {
		err = s.loanRepo.UpdateStatus(request.LoanID, models.LoanStatusRepaid)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to update loan status to paid: %d", request.LoanID)
		}
	}

	paymentID, err := s.loanRepo.CreatePayment(payment)
	if err != nil {
		s.loanRepo.UpdateRemainingAmount(request.LoanID, loan.RemainingAmount)
		if err != nil {
			logrus.WithError(err).Warnf("Failed to create payment")
		}
		ctx = context.Background()
		s.accountRepo.DepositAccount(ctx, int(request.AccountID), request.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to deposit funds from account: %w", err)
		}
		return nil, fmt.Errorf("failed to record payment: %w", err)
	}

	payment.ID = paymentID
	return payment, nil
}

func (s *LoanService) GetPaymentsByLoanID(loanID int64) ([]*models.LoanPayment, error) {
	payments, err := s.loanRepo.GetPaymentsByLoanID(loanID)
	if err != nil {
		logrus.WithError(err).Error("LoanService.GetPaymentsByLoanID: error getting payments")
		return nil, err
	}
	return payments, nil
}
