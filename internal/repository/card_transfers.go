package repository

import (
	"custom-banking/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type CardTransfersRepo struct {
	db *sqlx.DB
}

func NewCardTransfers(db *sqlx.DB) *CardTransfersRepo {
	return &CardTransfersRepo{db: db}
}

func (r *CardTransfersRepo) GetCardByNumber(cardNumber string) (*models.CardInfo, error) {
	query := `
		SELECT c.id, c.card_number, c.cardholder_name, c.expiration_date, c.card_type, c.account_id,
		       a.amount as available_amount, cu.code as currency_code,
		       CASE WHEN a.blocked = false AND c.expiration_date > NOW() THEN true ELSE false END as is_active
		FROM cards c
		JOIN accounts a ON c.account_id = a.id
		JOIN currency cu ON a.currency_id = cu.id
		WHERE c.card_number = $1
	`

	card := &models.CardInfo{}
	err := r.db.QueryRow(query, cardNumber).Scan(
		&card.ID,
		&card.CardNumber,
		&card.CardholderName,
		&card.ExpirationDate,
		&card.CardType,
		&card.AccountID,
		&card.AvailableAmount,
		&card.CurrencyCode,
		&card.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("card with number %s not found", cardNumber)
		}
		logrus.WithError(err).Errorf("CardTransfersRepo.GetCardByNumber: error getting card with number %s", cardNumber)
		return nil, err
	}

	if len(card.CardNumber) == 16 {
		card.MaskedNumber = fmt.Sprintf("**** **** **** %s", card.CardNumber[12:])
	}

	return card, nil
}

func (r *CardTransfersRepo) GetCardByID(id int64) (*models.CardInfo, error) {
	query := `
		SELECT c.id, c.card_number, c.cardholder_name, c.expiration_date, c.card_type, c.account_id,
		       a.amount as available_amount, cu.code as currency_code,
		       CASE WHEN a.blocked = false AND c.expiration_date > NOW() THEN true ELSE false END as is_active
		FROM cards c
		JOIN accounts a ON c.account_id = a.id
		JOIN currency cu ON a.currency_id = cu.id
		WHERE c.id = $1
	`

	card := &models.CardInfo{}
	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.CardNumber,
		&card.CardholderName,
		&card.ExpirationDate,
		&card.CardType,
		&card.AccountID,
		&card.AvailableAmount,
		&card.CurrencyCode,
		&card.IsActive,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("card with id %d not found", id)
		}
		logrus.WithError(err).Errorf("CardTransfersRepo.GetCardByID: error getting card with id %d", id)
		return nil, err
	}

	if len(card.CardNumber) == 16 {
		card.MaskedNumber = fmt.Sprintf("**** **** **** %s", card.CardNumber[12:])
	}

	return card, nil
}

func (r *CardTransfersRepo) GetAccountBalance(accountID int64) (float64, error) {
	query := `SELECT amount FROM accounts WHERE id = $1`

	var balance float64
	err := r.db.QueryRow(query, accountID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("account with id %d not found", accountID)
		}
		logrus.WithError(err).Errorf("CardTransfersRepo.GetAccountBalance: error getting balance for account %d", accountID)
		return 0, err
	}

	return balance, nil
}

func (r *CardTransfersRepo) CreateCardTransfer(fromCardID, toCardID int64, amount float64, description string) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		logrus.WithError(err).Error("CardTransfersRepo.CreateCardTransfer: error starting transaction")
		return 0, err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			logrus.WithError(err).Error("CardTransfersRepo.CreateCardTransfer: transaction rolled back")
		}
	}()

	var fromAccountID, toAccountID int64
	err = tx.QueryRow("SELECT account_id FROM cards WHERE id = $1", fromCardID).Scan(&fromAccountID)
	if err != nil {
		return 0, fmt.Errorf("error getting account for source card: %w", err)
	}

	err = tx.QueryRow("SELECT account_id FROM cards WHERE id = $1", toCardID).Scan(&toAccountID)
	if err != nil {
		return 0, fmt.Errorf("error getting account for destination card: %w", err)
	}

	referenceNumber := fmt.Sprintf("CT%d%d", time.Now().Unix(), fromCardID)

	var transactionID int64
	err = tx.QueryRow(`
		INSERT INTO transactions (from_account, to_account, amount, status, transaction_type, description, reference_number)
		VALUES ($1, $2, $3, 'pending', 'card_transfer', $4, $5)
		RETURNING id
	`, fromAccountID, toAccountID, amount, description, referenceNumber).Scan(&transactionID)

	if err != nil {
		return 0, fmt.Errorf("error creating transaction: %w", err)
	}

	_, err = tx.Exec("UPDATE accounts SET amount = amount - $1 WHERE id = $2", amount, fromAccountID)
	if err != nil {
		return 0, fmt.Errorf("error updating source account balance: %w", err)
	}

	_, err = tx.Exec("UPDATE accounts SET amount = amount + $1 WHERE id = $2", amount, toAccountID)
	if err != nil {
		return 0, fmt.Errorf("error updating destination account balance: %w", err)
	}

	_, err = tx.Exec("UPDATE transactions SET status = 'completed', date_updated = NOW() WHERE id = $1", transactionID)
	if err != nil {
		return 0, fmt.Errorf("error updating transaction status: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("error committing transaction: %w", err)
	}

	return transactionID, nil
}

func (r *CardTransfersRepo) GetTransactionDetails(transactionID int64) (*models.CardTransferResult, error) {
	query := `
		SELECT t.id, t.status, c_from.card_number as from_card_number, c_to.card_number as to_card_number,
		       t.amount, 0 as fee, t.amount as total_amount, t.date_created, t.reference_number, t.description
		FROM transactions t
		JOIN accounts a_from ON t.from_account = a_from.id
		JOIN accounts a_to ON t.to_account = a_to.id
		JOIN cards c_from ON a_from.id = c_from.account_id
		JOIN cards c_to ON a_to.id = c_to.account_id
		WHERE t.id = $1
	`

	result := &models.CardTransferResult{}
	err := r.db.QueryRow(query, transactionID).Scan(
		&result.TransactionID,
		&result.Status,
		&result.FromCardNumber,
		&result.ToCardNumber,
		&result.Amount,
		&result.Fee,
		&result.TotalAmount,
		&result.Date,
		&result.ReferenceNumber,
		&result.Description,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction with id %d not found", transactionID)
		}
		logrus.WithError(err).Errorf("CardTransfersRepo.GetTransactionDetails: error getting details for transaction %d", transactionID)
		return nil, err
	}

	return result, nil
}

func (r *CardTransfersRepo) ListCardTransactions(cardID, limit, offset int64) ([]*models.CardTransferResult, error) {
	query := `
		SELECT t.id, t.status, c_from.card_number as from_card_number, c_to.card_number as to_card_number,
		       t.amount, 0 as fee, t.amount as total_amount, t.date_created, t.reference_number, t.description
		FROM transactions t
		JOIN accounts a_from ON t.from_account = a_from.id
		JOIN accounts a_to ON t.to_account = a_to.id
		JOIN cards c_from ON a_from.id = c_from.account_id
		JOIN cards c_to ON a_to.id = c_to.account_id
		WHERE c_from.id = $1 OR c_to.id = $1
		ORDER BY t.date_created DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, cardID, limit, offset)
	if err != nil {
		logrus.WithError(err).Errorf("CardTransfersRepo.ListCardTransactions: error getting transactions for card %d", cardID)
		return nil, err
	}
	defer rows.Close()

	var transactions []*models.CardTransferResult
	for rows.Next() {
		transaction := &models.CardTransferResult{}
		err := rows.Scan(
			&transaction.TransactionID,
			&transaction.Status,
			&transaction.FromCardNumber,
			&transaction.ToCardNumber,
			&transaction.Amount,
			&transaction.Fee,
			&transaction.TotalAmount,
			&transaction.Date,
			&transaction.ReferenceNumber,
			&transaction.Description,
		)
		if err != nil {
			logrus.WithError(err).Error("CardTransfersRepo.ListCardTransactions: error scanning transaction row")
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}

func (r *CardTransfersRepo) GetCardsByUserID(userID int64, params models.CardListParams) ([]*models.CardInfo, int64, error) {
	whereClause := "WHERE ua.user_id = $1"
	args := []interface{}{userID}
	argCount := 2

	if params.AccountID != 0 {
		whereClause += fmt.Sprintf(" AND c.account_id = $%d", argCount)
		args = append(args, params.AccountID)
		argCount++
	}

	if params.CardType != "" {
		whereClause += fmt.Sprintf(" AND c.card_type = $%d", argCount)
		args = append(args, params.CardType)
		argCount++
	}

	if params.Active != nil {
		if *params.Active {
			whereClause += " AND a.blocked = false AND c.expiration_date > NOW()"
		} else {
			whereClause += " AND (a.blocked = true OR c.expiration_date <= NOW())"
		}
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM cards c
		JOIN accounts a ON c.account_id = a.id
		JOIN user_accounts ua ON a.id = ua.account_id
		%s
	`, whereClause)

	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersRepo.GetCardsByUserID: error counting cards")
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize

	query := fmt.Sprintf(`
		SELECT c.id, c.card_number, c.cardholder_name, c.expiration_date, c.card_type, c.account_id,
		       a.amount as available_amount, cu.code as currency_code,
		       CASE WHEN a.blocked = false AND c.expiration_date > NOW() THEN true ELSE false END as is_active
		FROM cards c
		JOIN accounts a ON c.account_id = a.id
		JOIN currency cu ON a.currency_id = cu.id
		JOIN user_accounts ua ON a.id = ua.account_id
		%s
		ORDER BY c.id DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("CardTransfersRepo.GetCardsByUserID: error getting cards")
		return nil, 0, err
	}
	defer rows.Close()

	var cards []*models.CardInfo
	for rows.Next() {
		card := &models.CardInfo{}
		err := rows.Scan(
			&card.ID,
			&card.CardNumber,
			&card.CardholderName,
			&card.ExpirationDate,
			&card.CardType,
			&card.AccountID,
			&card.AvailableAmount,
			&card.CurrencyCode,
			&card.IsActive,
		)
		if err != nil {
			logrus.WithError(err).Error("CardTransfersRepo.GetCardsByUserID: error scanning card row")
			return nil, 0, err
		}

		if len(card.CardNumber) == 16 {
			card.MaskedNumber = fmt.Sprintf("**** **** **** %s", card.CardNumber[12:])
		}

		cards = append(cards, card)
	}

	return cards, totalCount, nil
}
