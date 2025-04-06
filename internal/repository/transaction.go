package repository

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

const (
	transactionPendingStatus   = "pending"
	transactionCompletedStatus = "completed"
	transactionFailedStatus    = "failed"
	transactionCancelledStatus = "cancelled"

	ingoingTransactionType  = "ingoing"
	outgoingTransactionType = "outgoing"
)

type Transactions struct {
	db *sqlx.DB
}

func NewTransactions(db *sqlx.DB) *Transactions {
	return &Transactions{db}
}

func (r Transactions) CreateTransaction(ctx context.Context, fromAccountID, toAccountID int, amount float64) (models.Transaction, error) {
	fields := logrus.Fields{
		"layer":           "repository",
		"repository":      "Transaction",
		"method":          "CreateTransaction",
		"from_account_id": fromAccountID,
		"to_account_id":   toAccountID,
		"amount":          amount,
	}

	nullableFromAccountID := sql.NullInt64{}
	if fromAccountID != 0 {
		nullableFromAccountID.Int64 = int64(fromAccountID)
		nullableFromAccountID.Valid = true
	}

	query := `INSERT INTO transactions 
		(from_account, to_account, amount, status, date_created, transaction_type) 
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, $5) 
		RETURNING id, from_account, to_account, amount, status, date_created, date_updated, transaction_type, description, reference_number`

	var transactionType string
	if fromAccountID == 0 {
		transactionType = ingoingTransactionType
	} else {
		transactionType = outgoingTransactionType
	}

	row := r.db.QueryRowContext(ctx, query, nullableFromAccountID, toAccountID, amount, transactionPendingStatus, transactionType)
	if row.Err() != nil {
		logrus.WithError(row.Err()).
			WithFields(fields).
			Error("execution inserting into transactions query error")

		return models.Transaction{}, errors.Wrap(row.Err(), "execution inserting into transactions query error")
	}

	transaction := models.Transaction{}
	if err := row.Scan(
		&transaction.ID,
		&transaction.FromAccount,
		&transaction.ToAccount,
		&transaction.Amount,
		&transaction.Status,
		&transaction.DateCreated,
		&transaction.DateUpdated,
		&transaction.TransactionType,
		&transaction.Description,
		&transaction.ReferenceNumber,
	); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning row into struct error")

		return models.Transaction{}, errors.Wrap(err, "scanning row into struct error")
	}

	return transaction, nil
}

func (r Transactions) SetTransactionStatusToSent(ctx context.Context, transactionID int) error {
	fields := logrus.Fields{
		"layer":          "repository",
		"repository":     "Transaction",
		"method":         "SetTransactionStatusToCompleted",
		"transaction_id": transactionID,
	}

	query := "UPDATE transactions SET status=$1, date_updated=CURRENT_TIMESTAMP WHERE id = $2"

	if _, err := r.db.ExecContext(ctx, query, transactionCompletedStatus, transactionID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating status into transactions query error")

		return errors.Wrap(err, "execution updating status into transactions query error")
	}

	return nil
}

func (r Transactions) SetTransactionStatusToCancelled(ctx context.Context, transactionID int) error {
	fields := logrus.Fields{
		"layer":          "repository",
		"repository":     "Transaction",
		"method":         "SetTransactionStatusToCancelled",
		"transaction_id": transactionID,
	}

	query := "UPDATE transactions SET status=$1, date_updated=CURRENT_TIMESTAMP WHERE id = $2"

	if _, err := r.db.ExecContext(ctx, query, transactionCancelledStatus, transactionID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating status into transactions query error")

		return errors.Wrap(err, "execution updating status into transactions query error")
	}

	return nil
}

func (r Transactions) SetTransactionStatusToFailed(ctx context.Context, transactionID int) error {
	fields := logrus.Fields{
		"layer":          "repository",
		"repository":     "Transaction",
		"method":         "SetTransactionStatusToFailed",
		"transaction_id": transactionID,
	}

	query := "UPDATE transactions SET status=$1, date_updated=CURRENT_TIMESTAMP WHERE id = $2"

	if _, err := r.db.ExecContext(ctx, query, transactionFailedStatus, transactionID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating status into transactions query error")

		return errors.Wrap(err, "execution updating status into transactions query error")
	}

	return nil
}

func (r Transactions) GetTransactionList(ctx context.Context, accountID int, ordering models.Orderings, paginator models.Paginator) ([]models.Transaction, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Transaction",
		"method":     "CreateTransaction",
		"account_id": accountID,
	}

	listTransaction := make([]models.Transaction, 0, paginator.PerPage)

	//query := "SELECT * FROM transactions WHERE from_account = $1 OR to_account = $2"
	qb := squirrel.Select("*").
		From("transactions").
		Where("from_account = ? OR to_account = ?", accountID, accountID)

	if ordering != nil {
		var parts []string
		for field, direction := range ordering {
			parts = append(parts, fmt.Sprintf("%s %s", field, strings.ToUpper(direction)))
		}

		qb = qb.OrderBy(parts...)
	} else {
		qb = qb.OrderBy("id ASC")
	}

	qb = qb.Limit(uint64(paginator.PerPage)).
		Offset(uint64((paginator.Page - 1) * paginator.PerPage)).
		PlaceholderFormat(squirrel.Dollar)

	query, params, err := qb.ToSql()
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("building a ToSql query into a SQL string error")

		return nil, errors.Wrap(err, "building a ToSql query into a SQL string error")
	}

	rows, err := r.db.QueryxContext(ctx, query, params...)
	if err != nil && rows.Err() != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting transaction list from transaction query error")

		return nil, errors.Wrap(err, "execution getting transaction list from transaction query error")
	}

	for rows.Next() {
		var transaction models.Transaction
		if err := rows.StructScan(&transaction); err != nil {
			logrus.WithError(err).
				WithFields(fields).
				Error("scanning rows into struct error")

			return nil, errors.Wrap(err, "scanning rows into struct error")
		}
		listTransaction = append(listTransaction, transaction)
	}

	return listTransaction, nil
}
