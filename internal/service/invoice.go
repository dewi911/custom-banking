package service

import (
	"context"
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"time"
)

type InvoiceService struct {
	repo        InvoiceRepository
	accountRepo AccountRepository
	cardRepo    CardRepository
	userRepo    UsersRepository
}

func NewInvoiceService(
	repo InvoiceRepository,
	accountRepo AccountRepository,
	cardRepo CardRepository,
	userRepo UsersRepository,
) *InvoiceService {
	return &InvoiceService{
		repo:        repo,
		accountRepo: accountRepo,
		cardRepo:    cardRepo,
		userRepo:    userRepo,
	}
}

func (s *InvoiceService) CreateInvoice(ctx context.Context, userID int64, request models.InvoiceRequest) (*models.Invoice, error) {
	if request.RecipientName == "" || request.RecipientAccount == "" || request.Currency == "" {
		return nil, errors.New("missing required fields")
	}

	if len(request.Items) == 0 {
		return nil, errors.New("at least one item is required")
	}

	totalAmount := 0.0
	for i, item := range request.Items {
		if item.Description == "" || item.Quantity <= 0 || item.UnitPrice <= 0 {
			return nil, fmt.Errorf("invalid item at index %d", i)
		}

		amount := float64(item.Quantity) * item.UnitPrice
		request.Items[i].Amount = amount
		totalAmount += amount
	}

	now := time.Now()
	invoice := &models.Invoice{
		UserID:           userID,
		RecipientID:      request.RecipientID,
		RecipientName:    request.RecipientName,
		RecipientAccount: request.RecipientAccount,
		TotalAmount:      totalAmount,
		Currency:         request.Currency,
		Description:      request.Description,
		DueDate:          request.DueDate,
		CreatedAt:        now,
		UpdatedAt:        now,
		Status:           models.InvoiceStatusPending,
		Items:            request.Items,
	}

	invoiceNumber, err := s.repo.GenerateInvoiceNumber()
	if err != nil {
		logrus.WithError(err).Error("InvoiceService.CreateInvoice: error generating invoice number")
		return nil, err
	}
	invoice.InvoiceNumber = invoiceNumber

	id, err := s.repo.CreateInvoice(invoice)
	if err != nil {
		logrus.WithError(err).Error("InvoiceService.CreateInvoice: error creating invoice")
		return nil, err
	}

	invoice, err = s.repo.GetInvoiceByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.CreateInvoice: error getting created invoice %d", id)
		return nil, err
	}

	return invoice, nil
}

func (s *InvoiceService) GetInvoice(ctx context.Context, id int64, userID int64) (*models.Invoice, error) {
	invoice, err := s.repo.GetInvoiceByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.GetInvoice: error getting invoice %d", id)
		return nil, err
	}

	if invoice.UserID != userID && invoice.RecipientID != userID {
		return nil, errors.New("access denied")
	}

	return invoice, nil
}

func (s *InvoiceService) GetInvoiceByNumber(ctx context.Context, invoiceNumber string) (*models.Invoice, error) {
	invoice, err := s.repo.GetInvoiceByNumber(invoiceNumber)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.GetInvoiceByNumber: error getting invoice %s", invoiceNumber)
		return nil, err
	}

	return invoice, nil
}

func (s *InvoiceService) CancelInvoice(ctx context.Context, id int64, userID int64) error {
	invoice, err := s.repo.GetInvoiceByID(id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.CancelInvoice: error getting invoice %d", id)
		return err
	}

	if invoice.UserID != userID {
		return errors.New("access denied")
	}

	if invoice.Status != models.InvoiceStatusPending {
		return fmt.Errorf("cannot cancel invoice with status %s", invoice.Status)
	}

	err = s.repo.UpdateInvoiceStatus(id, models.InvoiceStatusCanceled)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.CancelInvoice: error updating status for invoice %d", id)
		return err
	}

	return nil
}

func (s *InvoiceService) GetUserInvoices(ctx context.Context, userID, page, pageSize int64) ([]*models.Invoice, int64, error) {
	return s.repo.GetUserInvoices(userID, page, pageSize)
}

func (s *InvoiceService) ListInvoices(ctx context.Context, filter models.InvoiceFilter) ([]*models.Invoice, int64, error) {
	return s.repo.ListInvoices(filter)
}

func (s *InvoiceService) PayInvoice(ctx context.Context, userID int64, request models.InvoicePaymentRequest) (*models.Invoice, error) {
	if request.InvoiceID <= 0 {
		return nil, errors.New("invalid invoice ID")
	}

	if request.PaymentMethod != models.PaymentMethodCard && request.PaymentMethod != models.PaymentMethodAccount {
		return nil, errors.New("invalid payment method")
	}

	if request.PaymentMethod == models.PaymentMethodCard && request.CardID <= 0 {
		return nil, errors.New("card ID is required for card payment")
	}

	if request.PaymentMethod == models.PaymentMethodAccount && request.AccountID <= 0 {
		return nil, errors.New("account ID is required for account payment")
	}

	invoice, err := s.repo.GetInvoiceByID(request.InvoiceID)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.PayInvoice: error getting invoice %d", request.InvoiceID)
		return nil, err
	}

	if invoice.Status == models.InvoiceStatusPaid {
		return nil, errors.New("invoice is already paid")
	}

	if invoice.Status == models.InvoiceStatusCanceled {
		return nil, errors.New("invoice is cancelled and cannot be paid")
	}

	if invoice.Status == models.InvoiceStatusExpired {
		return nil, errors.New("invoice is expired and cannot be paid")
	}

	accountOrCardID := request.AccountID
	if request.PaymentMethod == models.PaymentMethodCard {
		accountOrCardID = request.CardID
	}

	err = s.repo.ProcessInvoicePayment(request.InvoiceID, request.PaymentMethod, accountOrCardID)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.PayInvoice: error processing payment for invoice %d", request.InvoiceID)
		return nil, err
	}

	invoice, err = s.repo.GetInvoiceByID(request.InvoiceID)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceService.PayInvoice: error getting updated invoice %d", request.InvoiceID)
		return nil, err
	}

	return invoice, nil
}
