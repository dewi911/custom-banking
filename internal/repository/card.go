package repository

import (
	"context"
	"custom-banking/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const defaultCardExpirationPeriodYears = 3

type Card struct {
	db *sqlx.DB
}

func NewCard(db *sqlx.DB) *Card {
	return &Card{db: db}
}

func (r *Card) CreateCard(ctx context.Context, accountID int, cardNumber string, cardholderName string, cvvCode string) (models.Card, error) {
	expirationDate := time.Now().AddDate(defaultCardExpirationPeriodYears, 0, 0)
	cardType := "standard"

	fields := logrus.Fields{
		"layer":           "repository",
		"repository":      "Card",
		"method":          "CreateCard",
		"account_id":      accountID,
		"card_number":     cardNumber,
		"cardholder_name": cardholderName,
		"expiration_date": expirationDate,
	}

	query := "INSERT INTO cards (account_id, card_number, cardholder_name, expiration_date, cvv_code, card_type, cashback_percentage) " +
		"VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, account_id, card_number, cardholder_name, expiration_date, cvv_code, card_type, cashback_percentage"

	row := r.db.QueryRowxContext(ctx, query, accountID, cardNumber, cardholderName, expirationDate, cvvCode, cardType, 3)
	if row.Err() != nil {
		logrus.WithError(row.Err()).
			WithFields(fields).
			Error("execution inserting into cards query error")

		return models.Card{}, errors.Wrap(row.Err(), "execution inserting into cards query error")
	}

	card := models.Card{}
	if err := row.StructScan(&card); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning row into struct error")

		return models.Card{}, errors.Wrap(err, "scanning row into struct error")
	}

	return card, nil
}

func (r *Card) GetCard(ctx context.Context, id, accountID int) (models.Card, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Card",
		"method":     "GetCard",
		"id":         id,
		"account_id": accountID,
	}

	query := "SELECT * FROM  cards WHERE id=$1 AND account_id=$2"

	var card models.Card

	row := r.db.QueryRowxContext(ctx, query, id, accountID)
	if err := row.StructScan(&card); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning row into struct error")

		return models.Card{}, errors.Wrap(err, "scanning row into struct error")
	}

	return card, nil
}

func (r *Card) GetCardListUser(ctx context.Context, userID int) ([]models.Card, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Card",
		"method":     "GetCardListUser",
		"user_id":    userID,
	}

	query := "SELECT c.* FROM cards c INNER JOIN user_accounts ua on ua.account_id = c.account_id WHERE ua.user_id = $1 ORDER BY c.id DESC"

	rows, err := r.db.QueryxContext(ctx, query, userID)
	if err != nil && rows.Err() != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution select cards list query error")

		return nil, errors.Wrap(rows.Err(), "execution select list cards query error")
	}
	defer rows.Close()

	ListCards := make([]models.Card, 0)
	for rows.Next() {
		var card models.Card
		if err := rows.StructScan(&card); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning row into struct error")

			return nil, errors.Wrap(err, "scanning row into struct error")
		}
		ListCards = append(ListCards, card)
	}

	if err := rows.Err(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("error iterating over result rows")
		return nil, errors.Wrap(err, "error iterating over result rows")
	}

	return ListCards, nil
}

func (r *Card) GetCardListByAccount(ctx context.Context, userID, accountID int) ([]models.Card, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Card",
		"method":     "GetCardListByAccount",
		"user_id":    userID,
		"account_id": accountID,
	}

	query := "SELECT c.* FROM cards c INNER JOIN user_accounts ua ON ua.account_id = c.account_id WHERE ua.user_id = $1 AND c.account_id = $2 ORDER BY c.id DESC"

	rows, err := r.db.QueryxContext(ctx, query, userID, accountID)
	if err != nil && rows.Err() != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution select list cards query error")

		return nil, errors.Wrap(rows.Err(), "execution select list cards query error")
	}
	defer rows.Close()

	ListCards := make([]models.Card, 0)
	for rows.Next() {
		var card models.Card
		if err := rows.StructScan(&card); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning row into struct error")

			return nil, errors.Wrap(err, "scanning row into struct error")
		}
		ListCards = append(ListCards, card)
	}

	if err := rows.Err(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("error iterating over result rows")
		return nil, errors.Wrap(err, "error iterating over result rows")
	}

	return ListCards, nil
}
