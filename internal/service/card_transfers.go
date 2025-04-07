package service

import (
	"custom-banking/internal/models"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
)

type CardTransfersService struct {
	repo CardTransfersRepository
}

func NewCardTransfersService(repo CardTransfersRepository) *CardTransfersService {
	return &CardTransfersService{
		repo: repo,
	}
}

func (s *CardTransfersService) TransferBetweenCards(request models.CardTransferRequest) (*models.CardTransferResult, error) {
	fromCard, err := s.repo.GetCardByNumber(request.FromCardNumber)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.TransferBetweenCards: error getting source card")
		return nil, fmt.Errorf("failed to find source card: %w", err)
	}

	if !fromCard.IsActive {
		return nil, errors.New("source card is not active")
	}

	toCard, err := s.repo.GetCardByNumber(request.ToCardNumber)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.TransferBetweenCards: error getting destination card")
		return nil, fmt.Errorf("failed to find destination card: %w", err)
	}

	if !toCard.IsActive {
		return nil, errors.New("destination card is not active")
	}

	if fromCard.AvailableAmount < request.Amount {
		return nil, errors.New("insufficient funds on source card")
	}

	if request.Amount <= 0 {
		return nil, errors.New("transfer amount must be positive")
	}

	transactionID, err := s.repo.CreateCardTransfer(fromCard.ID, toCard.ID, request.Amount, request.Description)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.TransferBetweenCards: error creating transfer")
		return nil, fmt.Errorf("failed to process transfer: %w", err)
	}

	result, err := s.repo.GetTransactionDetails(transactionID)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.TransferBetweenCards: error getting transaction details")
		return nil, fmt.Errorf("transfer completed but failed to retrieve details: %w", err)
	}

	return result, nil
}

func (s *CardTransfersService) GetCardByNumber(cardNumber string) (*models.CardInfo, error) {
	card, err := s.repo.GetCardByNumber(cardNumber)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.GetCardByNumber: error getting card")
		return nil, err
	}
	return card, nil
}

func (s *CardTransfersService) GetTransactionDetails(transactionID int64) (*models.CardTransferResult, error) {
	result, err := s.repo.GetTransactionDetails(transactionID)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.GetTransactionDetails: error getting transaction details")
		return nil, err
	}
	return result, nil
}

func (s *CardTransfersService) ListCardTransactions(cardID int64, page, pageSize int64) ([]*models.CardTransferResult, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	transactions, err := s.repo.ListCardTransactions(cardID, pageSize, offset)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.ListCardTransactions: error listing transactions")
		return nil, err
	}
	return transactions, nil
}

func (s *CardTransfersService) GetCardsByUserID(userID int64, params models.CardListParams) ([]*models.CardInfo, int64, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 10
	}

	cards, totalCount, err := s.repo.GetCardsByUserID(userID, params)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersService.GetCardsByUserID: error getting cards")
		return nil, 0, err
	}
	return cards, totalCount, nil
}
