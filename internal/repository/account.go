package repository

import (
	"context"
	"custom-banking/internal/models"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"strings"
)

type Account struct {
	db *sqlx.DB
}

func NewAccount(db *sqlx.DB) *Account {
	return &Account{db: db}
}

func (r *Account) GetUserIDByAccountID(ctx context.Context, accountID int) (int, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetUserIDByAccountID",
		"account_id": accountID,
	}

	var userID int

	row := r.db.QueryRowxContext(ctx, "SELECT user_id FROM accounts WHERE id = $1", accountID)
	if err := row.Err(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting userID by accountID query error")

		return 0, errors.Wrap(err, "execution getting userID by accountID query error")
	}

	if err := row.Scan(&userID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning userID error")

		return 0, errors.Wrap(err, "scanning userID error")
	}

	return userID, nil
}

func (r *Account) GetAccountIDByIban(ctx context.Context, iban string) (int, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccountIDByIban",
		"iban":       iban,
	}

	var accountID int

	row := r.db.QueryRowxContext(ctx, "SELECT id FROM accounts WHERE iban = $1", iban)
	if err := row.Scan(&accountID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning accountID error")

		return 0, errors.Wrap(err, "scanning accountID error")
	}

	return accountID, nil
}

func (r *Account) GetAccountCurrencyIDByID(ctx context.Context, accountID int) (int, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccountCurrencyIDByID",
		"account_id": accountID,
	}

	var currencyID int

	row := r.db.QueryRowxContext(ctx, "SELECT currency_id FROM accounts WHERE id = $1", accountID)
	if err := row.Scan(&currencyID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning currencyID error")

		return 0, errors.Wrap(err, "scanning currencyID error")
	}

	return currencyID, nil
}

func (r *Account) GetAccountCurrencyIDByIban(ctx context.Context, iban string) (int, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccountCurrencyIDByIban",
		"iban":       iban,
	}

	var currencyID int

	row := r.db.QueryRowxContext(ctx, "SELECT currency_id FROM accounts WHERE iban = $1", iban)
	if err := row.Scan(&currencyID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning currencyID error")

		return 0, errors.Wrap(err, "scanning currencyID error")
	}

	return currencyID, nil
}

func (r *Account) GetAccountAmount(ctx context.Context, accountID, userID int) (float64, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccountAmount",
		"user_id":    userID,
		"account_id": accountID,
	}

	var amount float64

	row := r.db.QueryRowxContext(ctx, "SELECT amount FROM  accounts WHERE id=$1 AND user_id=$2", accountID, userID)
	if err := row.Scan(&amount); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning amount error")

		return 0, errors.Wrap(err, "scanning amount error")
	}

	return amount, nil
}

func (r Account) ExistsAccount(ctx context.Context, accountID int) (bool, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "ExistsAccount",
		"account_id": accountID,
	}

	var exists bool

	query := "select exists(select amount from accounts where id = $1)"
	if err := r.db.QueryRowContext(ctx, query, accountID).Scan(&exists); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning exists error")

		return false, errors.Wrap(err, "scanning exists error")
	}

	return exists, nil
}

func (r *Account) Create(ctx context.Context, userID, currencyID int, iban string) (models.Account, error) {
	fields := logrus.Fields{
		"layer":       "repository",
		"repository":  "Account",
		"method":      "Create",
		"user_id":     userID,
		"currency_id": currencyID,
		"iban":        iban,
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("begin transaction error")
		return models.Account{}, errors.Wrap(err, "begin transaction error")
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				logrus.WithError(rbErr).
					WithFields(fields).
					Error("rollback transaction error")
			}
		}
	}()

	q := "INSERT INTO accounts (iban, currency_id, blocked, amount) VALUES ($1, $2, $3, $4) returning *"
	row := tx.QueryRowContext(ctx, q, iban, currencyID, false, 0)
	if row.Err() != nil {
		err = row.Err()
		logrus.WithError(err).
			WithFields(fields).
			Error("execution creating account query error")

		return models.Account{}, errors.Wrap(err, "execution creating account query error")
	}

	account := models.Account{}
	if err = row.Scan(
		&account.ID,
		&account.Iban,
		&account.Currency,
		&account.Blocked,
		&account.Amount,
	); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning row into struct error")

		return models.Account{}, errors.Wrap(err, "scanning row into struct error")
	}

	q = "INSERT INTO user_accounts (user_id, account_id, is_primary, access_level) VALUES ($1, $2, $3, $4)"
	_, err = tx.ExecContext(ctx, q, userID, account.ID, true, "owner")
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution creating user_accounts relationship error")

		return models.Account{}, errors.Wrap(err, "execution creating user_accounts relationship error")
	}

	if err = tx.Commit(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("commit transaction error")

		return models.Account{}, errors.Wrap(err, "commit transaction error")
	}

	account.UserID = userID

	return account, nil
}

func (r *Account) GetAccountsList(ctx context.Context, userID int, paginator models.Paginator, ordering models.Orderings) ([]models.Account, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccountsList",
		"user_id":    userID,
	}

	listAccounts := make([]models.Account, 0, paginator.PerPage)

	qb := squirrel.Select("*").
		From("accounts").
		Where("user_id = ?", userID)

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

	rows, err := r.db.QueryxContext(ctx, query, params...)
	if err != nil || rows.Err() != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting list account by id query error")

		return nil, errors.Wrap(err, "error execution getting accounts list query")
	}

	for rows.Next() {
		var account models.Account
		if err := rows.StructScan(&account); err != nil {
			return nil, errors.Wrap(err, "error scanning row into struct")
		}
		listAccounts = append(listAccounts, account)
	}

	return listAccounts, nil
}

func (r Account) GetAccount(ctx context.Context, accountID, userID int) (models.Account, error) {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "GetAccount",
		"account_id": accountID,
		"user_id":    userID,
	}

	var account models.Account

	row := r.db.QueryRowxContext(ctx, "SELECT * FROM  accounts WHERE id=$1 AND user_id=$2", accountID, userID)
	if err := row.Err(); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution getting account by id query error")

		return models.Account{}, errors.Wrap(err, "error execution getting account by id query")
	}

	if err := row.StructScan(&account); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("scanning account query result into struct error")

		return models.Account{}, errors.Wrap(err, "error scanning result into struct")
	}

	return account, nil
}

func (r Account) DeleteAccount(ctx context.Context, accountID int) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "DeleteAccount",
		"account_id": accountID,
	}

	if _, err := r.db.ExecContext(ctx, "DELETE FROM accounts WHERE id=$1", accountID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution deleting account by accountID query error")

		return errors.Wrap(err, "execution deleting account by accountID query error")
	}
	return nil
}

func (r Account) DepositAccount(ctx context.Context, accountID int, amount float64) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "DepositAccount",
		"account_id": accountID,
		"amount":     amount,
	}

	query := "UPDATE accounts SET amount = amount + $1 WHERE id = $2"
	if _, err := r.db.ExecContext(ctx, query, amount, accountID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating account by accountID query error")

		return errors.Wrap(err, "execution updating account by accountID query error")
	}
	return nil
}

func (r Account) TransferAccount(ctx context.Context, fromAccountID, userID int, amount float64, toAccountIban string) error {
	fields := logrus.Fields{
		"layer":           "repository",
		"repository":      "Account",
		"method":          "TransferAccount",
		"user_id":         userID,
		"from_account_id": fromAccountID,
		"to_account_iban": toAccountIban,
		"amount":          amount,
	}

	tx, err := r.db.Begin()
	if err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("begin transaction error")

		return errors.Wrap(err, "begin transaction error")
	}

	if _, err := tx.ExecContext(ctx, "UPDATE accounts SET amount = amount - $1 WHERE id = $2 AND user_id = $3", amount, fromAccountID, userID); err != nil {
		if err := tx.Rollback(); err != nil {
			return errors.Wrap(err, "rollback transaction error")
		}

		return errors.Wrap(err, "debit transaction error")
	}

	if _, err = tx.ExecContext(ctx, "UPDATE accounts SET amount = amount + $1 WHERE iban = $2", amount, toAccountIban); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating account by iban query error")

		if err := tx.Rollback(); err != nil {
			return errors.Wrap(err, "rollback transaction error")
		}

		return errors.Wrap(err, "money transfer transaction error")
	}

	if err = tx.Commit(); err != nil {
		return errors.Wrap(err, "commit transaction error")
	}

	return nil
}

func (r Account) BlockAccount(ctx context.Context, accountID, userID int) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "BlockAccount",
		"user_id":    userID,
		"account_id": accountID,
	}

	query := "update accounts set blocked = $1 where id = $2 and user_id = $3"

	if _, err := r.db.ExecContext(ctx, query, block, accountID, userID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating account by ID and userID query error")

		return errors.Wrap(err, "execution updating account by ID and userID query error")
	}

	return nil
}

func (r Account) UnblockAccount(ctx context.Context, accountID, userID int) error {
	fields := logrus.Fields{
		"layer":      "repository",
		"repository": "Account",
		"method":     "UnblockAccount",
		"user_id":    userID,
		"account_id": accountID,
	}

	query := "update accounts set blocked = $1 where id = $2 and user_id = $3"

	if _, err := r.db.ExecContext(ctx, query, unblock, accountID, userID); err != nil {
		logrus.WithError(err).
			WithFields(fields).
			Error("execution updating account by ID and userID query error")

		return errors.Wrap(err, "execution updating account by ID and userID query error")
	}

	return nil
}
