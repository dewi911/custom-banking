package repository

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type LoansRepo struct {
	db *sqlx.DB
}

func NewLoans(db *sqlx.DB) *LoansRepo {
	return &LoansRepo{db: db}
}

func (r *LoansRepo) Create(ctx context.Context, loan *models.Loan) (int64, error) {
	query := `
		INSERT INTO loans 
		(user_id, amount, currency_id, start_date, end_date, interest_rate, status, remaining_amount) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) 
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		loan.UserID,
		loan.Amount,
		loan.CurrencyID,
		loan.StartDate,
		loan.EndDate,
		loan.InterestRate,
		loan.Status,
		loan.RemainingAmount,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("LoansRepo.Create: error creating loan")
		return 0, err
	}

	return id, nil
}

func (r *LoansRepo) GetByID(ctx context.Context, id int64) (*models.Loan, error) {
	query := `
		SELECT l.*, c.code as currency_code
		FROM loans l
		JOIN currency c ON l.currency_id = c.id
		WHERE l.id = $1
	`

	loan := &models.Loan{}
	err := r.db.QueryRow(query, id).Scan(
		&loan.ID,
		&loan.UserID,
		&loan.Amount,
		&loan.CurrencyID,
		&loan.StartDate,
		&loan.EndDate,
		&loan.InterestRate,
		&loan.Status,
		&loan.RemainingAmount,
		&loan.CurrencyCode,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("loan with id %d not found", id)
		}
		logrus.WithError(err).Errorf("LoansRepo.GetByID: error getting loan with id %d", id)
		return nil, err
	}

	return loan, nil
}

func (r *LoansRepo) GetByUserID(ctx context.Context, userID int64) ([]*models.Loan, error) {
	query := `
		SELECT l.*, c.code as currency_code
		FROM loans l
		JOIN currency c ON l.currency_id = c.id
		WHERE l.user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		logrus.WithError(err).Errorf("LoansRepo.GetByUserID: error getting loans for user %d", userID)
		return nil, err
	}
	defer rows.Close()

	var loans []*models.Loan
	for rows.Next() {
		loan := &models.Loan{}
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Amount,
			&loan.CurrencyID,
			&loan.StartDate,
			&loan.EndDate,
			&loan.InterestRate,
			&loan.Status,
			&loan.RemainingAmount,
			&loan.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("LoansRepo.GetByUserID: error scanning loan row")
			return nil, err
		}
		loans = append(loans, loan)
	}

	return loans, nil
}

func (r *LoansRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE loans SET status = $1 WHERE id = $2`

	res, err := r.db.Exec(query, status, id)
	if err != nil {
		logrus.WithError(err).Errorf("LoansRepo.UpdateStatus: error updating status for loan %d", id)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("LoansRepo.UpdateStatus: error getting rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("loan with id %d not found", id)
	}

	return nil
}

func (r *LoansRepo) UpdateRemainingAmount(ctx context.Context, id int64, amount float64) error {
	query := `UPDATE loans SET remaining_amount = $1 WHERE id = $2`

	res, err := r.db.Exec(query, amount, id)
	if err != nil {
		logrus.WithError(err).Errorf("LoansRepo.UpdateRemainingAmount: error updating amount for loan %d", id)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("LoansRepo.UpdateRemainingAmount: error getting rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("loan with id %d not found", id)
	}

	return nil
}

func (r *LoansRepo) List(ctx context.Context, params models.LoanListParams) ([]*models.Loan, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if params.UserID != 0 {
		whereClause += fmt.Sprintf(" AND l.user_id = $%d", argCount)
		args = append(args, params.UserID)
		argCount++
	}

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND l.status = $%d", argCount)
		args = append(args, params.Status)
		argCount++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM loans l
		%s
	`, whereClause)

	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("LoansRepo.List: error counting loans")
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize

	query := fmt.Sprintf(`
		SELECT l.*, c.code as currency_code
		FROM loans l
		JOIN currency c ON l.currency_id = c.id
		%s
		ORDER BY l.start_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("LoansRepo.List: error getting loans")
		return nil, 0, err
	}
	defer rows.Close()

	var loans []*models.Loan
	for rows.Next() {
		loan := &models.Loan{}
		err := rows.Scan(
			&loan.ID,
			&loan.UserID,
			&loan.Amount,
			&loan.CurrencyID,
			&loan.StartDate,
			&loan.EndDate,
			&loan.InterestRate,
			&loan.Status,
			&loan.RemainingAmount,
			&loan.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("LoansRepo.List: error scanning loan row")
			return nil, 0, err
		}
		loans = append(loans, loan)
	}

	return loans, totalCount, nil
}

func (r *LoansRepo) CreatePayment(ctx context.Context, payment *models.LoanPayment) (int64, error) {
	//_, err := r.db.ExecContext(ctx, `
	//	CREATE TABLE IF NOT EXISTS loan_payments (
	//		id serial PRIMARY KEY,
	//		loan_id integer REFERENCES loans(id),
	//		amount numeric,
	//		date timestamp,
	//		status varchar(20),
	//		payment_method varchar(50),
	//		transaction_id integer
	//	);
	//
	//	CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_id ON loan_payments (loan_id);
	//`)
	//
	//if err != nil {
	//	logrus.WithError(err).Error("LoansRepo.CreatePayment: error creating loan_payments table")
	//	return 0, err
	//}

	query := `
		INSERT INTO loan_payments
		(loan_id, amount, date, status, payment_method, transaction_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(
		ctx,
		query,
		payment.LoanID,
		payment.Amount,
		payment.Date,
		payment.Status,
		payment.PaymentMethod,
		payment.TransactionID,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("LoansRepo.CreatePayment: error creating loan payment")
		return 0, err
	}

	return id, nil
}

func (r *LoansRepo) GetPaymentsByLoanID(ctx context.Context, loanID int64) ([]*models.LoanPayment, error) {
	//_, err := r.db.ExecContext(ctx, `
	//	CREATE TABLE IF NOT EXISTS loan_payments (
	//		id serial PRIMARY KEY,
	//		loan_id integer REFERENCES loans(id),
	//		amount numeric,
	//		date timestamp,
	//		status varchar(20),
	//		payment_method varchar(50),
	//		transaction_id integer
	//	);
	//
	//	CREATE INDEX IF NOT EXISTS idx_loan_payments_loan_id ON loan_payments (loan_id);
	//`)
	//
	//if err != nil {
	//	logrus.WithError(err).Error("LoansRepo.GetPaymentsByLoanID: error creating loan_payments table")
	//	return nil, err
	//}

	query := `
		SELECT id, loan_id, amount, date, status, payment_method, transaction_id
		FROM loan_payments
		WHERE loan_id = $1
		ORDER BY date DESC
	`

	rows, err := r.db.QueryContext(ctx, query, loanID)
	if err != nil {
		logrus.WithError(err).Errorf("LoansRepo.GetPaymentsByLoanID: error getting payments for loan %d", loanID)
		return nil, err
	}
	defer rows.Close()

	var payments []*models.LoanPayment
	for rows.Next() {
		payment := &models.LoanPayment{}
		err := rows.Scan(
			&payment.ID,
			&payment.LoanID,
			&payment.Amount,
			&payment.Date,
			&payment.Status,
			&payment.PaymentMethod,
			&payment.TransactionID,
		)
		if err != nil {
			logrus.WithError(err).Error("LoansRepo.GetPaymentsByLoanID: error scanning payment row")
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, nil
}
