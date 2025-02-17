package service

import (
	"context"
	"custom-banking/internal/models"
	"github.com/pkg/errors"
)

type Card struct {
	cardRepo        CardRepository
	userRepo        UsersRepository
	accountRepo     AccountRepository
	eventRepo       EventRepository
	randomGenerator RandomGenerator
}

func NewCard(cardRepo CardRepository, userRepo UsersRepository, accountRepo AccountRepository, eventRepo EventRepository, randomGenerator RandomGenerator) *Card {
	return &Card{
		cardRepo:        cardRepo,
		userRepo:        userRepo,
		accountRepo:     accountRepo,
		eventRepo:       eventRepo,
		randomGenerator: randomGenerator,
	}
}

func (s Card) CreateCard(ctx context.Context, accountID, userID int) (models.Card, error) {
	checkUserID, err := s.accountRepo.GetUserIDByAccountID(ctx, accountID)
	if err != nil {
		return models.Card{}, errors.Wrap(err, "getting userID by accountID error")
	}

	if checkUserID != userID {
		return models.Card{}, errors.Wrap(err, "userID matching error")
	}

	cardNumber := s.randomGenerator.GenerateRandomCardNumber()
	cvvCode := s.randomGenerator.GenerateRandomCvv()

	cardholderName, err := s.userRepo.GetUserNameAndSurnameByID(ctx, userID)
	if err != nil {
		return models.Card{}, errors.Wrap(err, "getting user name and surname error")
	}

	card, err := s.cardRepo.CreateCard(ctx, accountID, cardNumber, cardholderName, cvvCode)
	if err != nil {
		return models.Card{}, errors.Wrap(err, "card creation error")
	}

	event := models.Event{
		UserID:  userID,
		Type:    models.CardCreatedEvent,
		Message: "new card successfully created",
		Metadata: map[string]any{
			"card_id":         card.Id,
			"account_id":      accountID,
			"cardholder_name": card.CardholderName,
			"expiration_date": card.ExpirationDate,
		},
	}

	if err := s.eventRepo.CreateEvent(ctx, event); err != nil {
		return models.Card{}, errors.Wrap(err, "card created event creation error")
	}

	return card, nil
}

func (s Card) GetCardListUser(ctx context.Context, userID int) ([]models.Card, error) {
	listCard, err := s.cardRepo.GetCardListUser(ctx, userID)
	if err != nil {
		return nil, errors.Wrap(err, "getting card list error")
	}

	return listCard, nil
}

func (s Card) GetCardListByAccount(ctx context.Context, userID, accountID int) ([]models.Card, error) {
	listCard, err := s.cardRepo.GetCardListByAccount(ctx, userID, accountID)
	if err != nil {
		return nil, errors.Wrap(err, "getting card list by account error")
	}

	return listCard, nil
}

func (s Card) GetCard(ctx context.Context, cardID, accountID, userID int) (models.Card, error) {
	checkUserID, err := s.accountRepo.GetUserIDByAccountID(ctx, accountID)
	if err != nil {
		return models.Card{}, errors.Wrap(err, "getting userID by accountID error")
	}

	if checkUserID != userID {
		return models.Card{}, errors.New("userID matching check error")
	}

	card, err := s.cardRepo.GetCard(ctx, cardID, accountID)
	if err != nil {
		return models.Card{}, errors.Wrap(err, "getting card error")
	}

	return card, nil
}
