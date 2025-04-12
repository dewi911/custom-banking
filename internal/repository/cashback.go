package repository

import (
	"custom-banking/internal/models"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

type CashbackRepo struct {
	db *sqlx.DB
}

func NewCashbackRepo(db *sqlx.DB) *CashbackRepo {
	return &CashbackRepo{
		db: db,
	}
}

func (r *CashbackRepo) CreateCashbackSettings(settings *models.CashbackSettings) (int64, error) {
	query := `
		INSERT INTO cashback_settings (
			card_id, 
			user_id,
			rate, 
			min_amount, 
			max_amount, 
			active, 
			created_at, 
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		settings.CardID,
		settings.UserID,
		settings.Rate,
		settings.MinAmount,
		settings.MaxAmount,
		settings.Active,
		settings.CreatedAt,
		settings.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, err
	}

	settings.ID = id
	return id, nil
}

func (r *CashbackRepo) GetCashbackSettingsByCardID(cardID int64) (*models.CashbackSettings, error) {
	query := `
		SELECT 
			id, 
			card_id, 
			user_id,
			rate, 
			min_amount, 
			max_amount, 
			active, 
			created_at, 
			updated_at
		FROM cashback_settings
		WHERE card_id = $1
	`

	var settings models.CashbackSettings
	err := r.db.QueryRow(query, cardID).Scan(
		&settings.ID,
		&settings.CardID,
		&settings.UserID,
		&settings.Rate,
		&settings.MinAmount,
		&settings.MaxAmount,
		&settings.Active,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &settings, nil
}

func (r *CashbackRepo) UpdateCashbackSettings(settings *models.CashbackSettings) error {
	query := `
		UPDATE cashback_settings
		SET 
			rate = $1,
			min_amount = $2,
			max_amount = $3,
			active = $4,
			updated_at = $5
		WHERE id = $6
	`

	_, err := r.db.Exec(
		query,
		settings.Rate,
		settings.MinAmount,
		settings.MaxAmount,
		settings.Active,
		settings.UpdatedAt,
		settings.ID,
	)

	return err
}

func (r *CashbackRepo) DeactivateCashbackSettings(cardID int64) error {
	query := `
		UPDATE cashback_settings
		SET 
			active = false,
			updated_at = $1
		WHERE card_id = $2
	`

	_, err := r.db.Exec(query, time.Now(), cardID)

	return err
}

func (r *CashbackRepo) CreateCashbackTransaction(transaction *models.CashbackTransaction) (int64, error) {
	query := `
		INSERT INTO cashback_transactions (
			user_id, 
			card_id, 
			transaction_id, 
			amount, 
			status, 
			created_at, 
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		transaction.UserID,
		transaction.CardID,
		transaction.TransactionID,
		transaction.Amount,
		transaction.Status,
		transaction.CreatedAt,
		transaction.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("error creating cashback transaction: %w", err)
	}

	transaction.ID = id
	return id, nil
}

func (r *CashbackRepo) GetCashbackTransactions(userID int64, filter models.CashbackTransactionFilter) ([]*models.CashbackTransaction, int64, error) {
	var args []interface{}
	var whereClause []string

	whereClause = append(whereClause, "user_id = $1")
	args = append(args, userID)

	argNum := 2

	if filter.CardID > 0 {
		whereClause = append(whereClause, fmt.Sprintf("card_id = $%d", argNum))
		args = append(args, filter.CardID)
		argNum++
	}

	if filter.Status != "" {
		whereClause = append(whereClause, fmt.Sprintf("status = $%d", argNum))
		args = append(args, filter.Status)
		argNum++
	}

	if !filter.StartDate.IsZero() {
		whereClause = append(whereClause, fmt.Sprintf("created_at >= $%d", argNum))
		args = append(args, filter.StartDate)
		argNum++
	}

	if !filter.EndDate.IsZero() {
		whereClause = append(whereClause, fmt.Sprintf("created_at <= $%d", argNum))
		args = append(args, filter.EndDate)
		argNum++
	}

	whereSQL := ""
	if len(whereClause) > 0 {
		whereSQL = "WHERE " + strings.Join(whereClause, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cashback_transactions %s", whereSQL)

	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("error counting transactions: %w", err)
	}

	limit := filter.PageSize
	if limit <= 0 {
		limit = 10
	}

	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
		SELECT 
			id, 
			user_id, 
			card_id, 
			transaction_id, 
			amount, 
			status, 
			created_at, 
			updated_at,
			processed_at,
			payout_account_id,
			payout_reference
		FROM cashback_transactions
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argNum, argNum+1)

	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("error querying transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*models.CashbackTransaction
	for rows.Next() {
		var tx models.CashbackTransaction
		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.CardID,
			&tx.TransactionID,
			&tx.Amount,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&tx.ProcessedAt,
			&tx.PayoutAccountID,
			&tx.PayoutReference,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("error scanning transaction row: %w", err)
		}
		transactions = append(transactions, &tx)
	}

	return transactions, totalCount, nil
}

func (r *CashbackRepo) GetCashbackTransactionsByCardID(cardID int64) ([]*models.CashbackTransaction, error) {
	query := `
		SELECT 
			id, 
			user_id, 
			card_id, 
			transaction_id, 
			amount, 
			status, 
			created_at, 
			updated_at,
			processed_at,
			payout_account_id,
			payout_reference
		FROM cashback_transactions
		WHERE card_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, cardID)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions for card %d: %w", cardID, err)
	}
	defer rows.Close()

	var transactions []*models.CashbackTransaction
	for rows.Next() {
		var tx models.CashbackTransaction
		var processedAt sql.NullTime
		var payoutAccountID sql.NullInt64

		err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.CardID,
			&tx.TransactionID,
			&tx.Amount,
			&tx.Status,
			&tx.CreatedAt,
			&tx.UpdatedAt,
			&processedAt,
			&payoutAccountID,
			&tx.PayoutReference,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning transaction row: %w", err)
		}

		if processedAt.Valid {
			tx.ProcessedAt = &processedAt.Time
		}

		if payoutAccountID.Valid {
			tx.PayoutAccountID = &payoutAccountID.Int64
		}

		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (r *CashbackRepo) GetPendingCashbackAmount(userID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM cashback_transactions
		WHERE user_id = $1 AND status = $2
	`

	var amount float64
	err := r.db.QueryRow(query, userID, models.CashbackStatusPending).Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("error getting pending amount for user %d: %w", userID, err)
	}

	return amount, nil
}

func (r *CashbackRepo) GetOrCreateCashbackSummary(userID int64, accountID int64) (*models.CashbackSummary, error) {
	query := `
		SELECT id, user_id, account_id, pending_amount, last_payout_date, 
			next_payout_date, payout_frequency, updated_at
		FROM cashback_summaries
		WHERE user_id = $1`

	var summary models.CashbackSummary
	err := r.db.Get(&summary, query, userID)
	if err == nil {
		return &summary, nil
	}

	if err != sql.ErrNoRows {
		logrus.WithError(err).Errorf("CashbackRepo.GetOrCreateCashbackSummary: error getting summary for user %d", userID)
		return nil, err
	}

	now := time.Now()
	nextPayout := now.AddDate(0, 0, models.CashbackPayoutFrequency)

	createQuery := `
		INSERT INTO cashback_summaries (
			user_id, account_id, pending_amount, last_payout_date, 
			next_payout_date, payout_frequency, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id`

	var id int64
	err = r.db.QueryRow(
		createQuery,
		userID,
		accountID,
		0.0,
		now,
		nextPayout,
		models.CashbackPayoutFrequency,
		now,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Errorf("CashbackRepo.GetOrCreateCashbackSummary: error creating summary for user %d", userID)
		return nil, err
	}

	return &models.CashbackSummary{
		ID:              id,
		UserID:          userID,
		AccountID:       accountID,
		PendingAmount:   0,
		LastPayoutDate:  now,
		NextPayoutDate:  nextPayout,
		PayoutFrequency: models.CashbackPayoutFrequency,
		UpdatedAt:       now,
	}, nil
}

func (r *CashbackRepo) UpdateCashbackSummary(summary *models.CashbackSummary) error {
	query := `
		UPDATE cashback_summaries SET
			account_id = $1,
			pending_amount = $2,
			last_payout_date = $3,
			next_payout_date = $4,
			payout_frequency = $5,
			updated_at = $6
		WHERE id = $7`

	summary.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		summary.AccountID,
		summary.PendingAmount,
		summary.LastPayoutDate,
		summary.NextPayoutDate,
		summary.PayoutFrequency,
		summary.UpdatedAt,
		summary.ID,
	)

	if err != nil {
		logrus.WithError(err).Errorf("CashbackRepo.UpdateCashbackSummary: error updating summary %d", summary.ID)
		return err
	}

	return nil
}

func (r *CashbackRepo) ProcessCashbackForTransaction(transactionID int64, cardID int64) (*models.CashbackTransaction, error) {
	var tx struct {
		ID     int64
		Amount float64
		Date   time.Time
	}

	txQuery := `
		SELECT id, amount, created_at
		FROM transactions
		WHERE id = $1
	`

	err := r.db.QueryRow(txQuery, transactionID).Scan(&tx.ID, &tx.Amount, &tx.Date)
	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}

	settings, err := r.GetCashbackSettingsByCardID(cardID)
	if err != nil {
		return nil, fmt.Errorf("error getting cashback settings: %w", err)
	}

	if !settings.Active {
		return nil, fmt.Errorf("cashback is not active for this card")
	}

	var userID int64
	userQuery := `
		SELECT user_id
		FROM cards
		WHERE id = $1
	`

	err = r.db.QueryRow(userQuery, cardID).Scan(&userID)
	if err != nil {
		return nil, fmt.Errorf("error getting card owner: %w", err)
	}

	if tx.Amount < settings.MinAmount {
		return nil, fmt.Errorf("transaction amount below minimum for cashback eligibility")
	}

	cashbackAmount := tx.Amount * (settings.Rate / 100.0)

	if cashbackAmount > settings.MaxAmount {
		cashbackAmount = settings.MaxAmount
	}

	now := time.Now()
	cashbackTx := &models.CashbackTransaction{
		UserID:        userID,
		CardID:        cardID,
		TransactionID: tx.ID,
		Amount:        cashbackAmount,
		Status:        models.CashbackStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	insertQuery := `
		INSERT INTO cashback_transactions (
			user_id, 
			card_id, 
			transaction_id, 
			amount, 
			status, 
			created_at, 
			updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7
		) RETURNING id
	`

	var id int64
	err = r.db.QueryRow(
		insertQuery,
		cashbackTx.UserID,
		cashbackTx.CardID,
		cashbackTx.TransactionID,
		cashbackTx.Amount,
		cashbackTx.Status,
		cashbackTx.CreatedAt,
		cashbackTx.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("error creating cashback transaction: %w", err)
	}

	cashbackTx.ID = id

	summary, err := r.GetOrCreateCashbackSummary(userID, 0) // 0 is placeholder, would be account ID in real implementation
	if err != nil {
		return nil, fmt.Errorf("error getting cashback summary: %w", err)
	}

	summary.PendingAmount += cashbackAmount
	summary.UpdatedAt = now

	err = r.UpdateCashbackSummary(summary)
	if err != nil {
		return nil, fmt.Errorf("error updating cashback summary: %w", err)
	}

	return cashbackTx, nil
}

func (r *CashbackRepo) GetUsersPendingCashbackPayout() ([]*models.CashbackSummary, error) {
	query := `
		SELECT id, user_id, account_id, pending_amount, last_payout_date, 
			next_payout_date, payout_frequency, updated_at
		FROM cashback_summaries
		WHERE next_payout_date <= $1 AND pending_amount > 0`

	rows, err := r.db.Query(query, time.Now())
	if err != nil {
		logrus.WithError(err).Error("CashbackRepo.GetUsersPendingCashbackPayout: error querying for pending payouts")
		return nil, err
	}
	defer rows.Close()

	var summaries []*models.CashbackSummary
	for rows.Next() {
		var s models.CashbackSummary
		err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.AccountID,
			&s.PendingAmount,
			&s.LastPayoutDate,
			&s.NextPayoutDate,
			&s.PayoutFrequency,
			&s.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("CashbackRepo.GetUsersPendingCashbackPayout: error scanning row")
			return nil, err
		}

		summaries = append(summaries, &s)
	}

	return summaries, nil
}

func (r *CashbackRepo) UpdateCashbackTransactionsStatus(userID int64, status string) error {
	query := `
		UPDATE cashback_transactions
		SET 
			status = $1, 
			processed_at = $2,
			updated_at = $3
		WHERE user_id = $4 AND status = $5
	`

	now := time.Now()
	_, err := r.db.Exec(query, status, now, now, userID, models.CashbackStatusPending)
	if err != nil {
		return fmt.Errorf("error updating transaction status for user %d: %w", userID, err)
	}

	return nil
}
