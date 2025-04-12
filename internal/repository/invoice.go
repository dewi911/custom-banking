package repository

import (
	"custom-banking/internal/models"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type InvoiceRepo struct {
	db *sqlx.DB
}

func NewInvoiceRepo(db *sqlx.DB) *InvoiceRepo {
	return &InvoiceRepo{
		db: db,
	}
}

func (r *InvoiceRepo) CreateInvoice(invoice *models.Invoice) (int64, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.CreateInvoice: error starting transaction")
		return 0, err
	}

	if invoice.InvoiceNumber == "" {
		invoiceNumber, err := r.GenerateInvoiceNumber()
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		invoice.InvoiceNumber = invoiceNumber
	}

	now := time.Now()
	invoice.CreatedAt = now
	invoice.UpdatedAt = now
	invoice.Status = models.InvoiceStatusPending

	var id int64
	query := `
		INSERT INTO invoices (
			invoice_number, user_id, recipient_id, recipient_name, recipient_account,
			total_amount, currency, description, status, due_date, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		) RETURNING id`

	err = tx.QueryRow(
		query,
		invoice.InvoiceNumber,
		invoice.UserID,
		invoice.RecipientID,
		invoice.RecipientName,
		invoice.RecipientAccount,
		invoice.TotalAmount,
		invoice.Currency,
		invoice.Description,
		invoice.Status,
		invoice.DueDate,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	).Scan(&id)

	if err != nil {
		tx.Rollback()
		logrus.WithError(err).Error("InvoiceRepo.CreateInvoice: error creating invoice")
		return 0, err
	}

	invoice.ID = id

	if len(invoice.Items) > 0 {
		for i := range invoice.Items {
			invoice.Items[i].InvoiceID = id
		}

		err = r.createInvoiceItemsTx(tx, invoice.Items)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
	}

	err = tx.Commit()
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.CreateInvoice: error committing transaction")
		return 0, err
	}

	return id, nil
}

func (r *InvoiceRepo) GetInvoiceByID(id int64) (*models.Invoice, error) {
	query := `
		SELECT id, invoice_number, user_id, recipient_id, recipient_name, recipient_account,
			total_amount, currency, description, status, due_date, created_at, updated_at,
			paid_at, payment_method, payment_card_id, payment_account_id
		FROM invoices
		WHERE id = $1`

	var invoice models.Invoice
	var paidAt sql.NullTime
	var paymentMethod sql.NullString
	var paymentCardID, paymentAccountID sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&invoice.ID,
		&invoice.InvoiceNumber,
		&invoice.UserID,
		&invoice.RecipientID,
		&invoice.RecipientName,
		&invoice.RecipientAccount,
		&invoice.TotalAmount,
		&invoice.Currency,
		&invoice.Description,
		&invoice.Status,
		&invoice.DueDate,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
		&paidAt,
		&paymentMethod,
		&paymentCardID,
		&paymentAccountID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice with id %d not found", id)
		}
		logrus.WithError(err).Errorf("InvoiceRepo.GetInvoiceByID: error getting invoice with id %d", id)
		return nil, err
	}

	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}
	if paymentMethod.Valid {
		invoice.PaymentMethod = paymentMethod.String
	}
	if paymentCardID.Valid {
		invoice.PaymentCardID = &paymentCardID.Int64
	}
	if paymentAccountID.Valid {
		invoice.PaymentAccountID = &paymentAccountID.Int64
	}

	items, err := r.GetInvoiceItems(id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceRepo.GetInvoiceByID: error getting items for invoice %d", id)
		return &invoice, err
	}
	invoice.Items = items

	return &invoice, nil
}

func (r *InvoiceRepo) GetInvoiceByNumber(invoiceNumber string) (*models.Invoice, error) {
	query := `
		SELECT id, invoice_number, user_id, recipient_id, recipient_name, recipient_account,
			total_amount, currency, description, status, due_date, created_at, updated_at,
			paid_at, payment_method, payment_card_id, payment_account_id
		FROM invoices
		WHERE invoice_number = $1`

	var invoice models.Invoice
	var paidAt sql.NullTime
	var paymentMethod sql.NullString
	var paymentCardID, paymentAccountID sql.NullInt64

	err := r.db.QueryRow(query, invoiceNumber).Scan(
		&invoice.ID,
		&invoice.InvoiceNumber,
		&invoice.UserID,
		&invoice.RecipientID,
		&invoice.RecipientName,
		&invoice.RecipientAccount,
		&invoice.TotalAmount,
		&invoice.Currency,
		&invoice.Description,
		&invoice.Status,
		&invoice.DueDate,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
		&paidAt,
		&paymentMethod,
		&paymentCardID,
		&paymentAccountID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invoice with number %s not found", invoiceNumber)
		}
		logrus.WithError(err).Errorf("InvoiceRepo.GetInvoiceByNumber: error getting invoice with number %s", invoiceNumber)
		return nil, err
	}

	if paidAt.Valid {
		invoice.PaidAt = &paidAt.Time
	}
	if paymentMethod.Valid {
		invoice.PaymentMethod = paymentMethod.String
	}
	if paymentCardID.Valid {
		invoice.PaymentCardID = &paymentCardID.Int64
	}
	if paymentAccountID.Valid {
		invoice.PaymentAccountID = &paymentAccountID.Int64
	}

	items, err := r.GetInvoiceItems(invoice.ID)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceRepo.GetInvoiceByNumber: error getting items for invoice %s", invoiceNumber)
		return &invoice, err
	}
	invoice.Items = items

	return &invoice, nil
}

func (r *InvoiceRepo) UpdateInvoiceStatus(id int64, status string) error {
	query := `
		UPDATE invoices
		SET status = $1
		WHERE id = $2`

	_, err := r.db.Exec(query, status, id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceRepo.UpdateInvoiceStatus: error updating status for invoice %d", id)
		return err
	}

	return nil
}

func (r *InvoiceRepo) UpdateInvoicePaymentDetails(id int64, paymentMethod string, paidDate time.Time) error {
	query := `
		UPDATE invoices
		SET status = $1, payment_method = $2, paid_at = $3
		WHERE id = $4`

	_, err := r.db.Exec(query, models.InvoiceStatusPaid, paymentMethod, paidDate, id)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceRepo.UpdateInvoicePaymentDetails: error updating payment details for invoice %d", id)
		return err
	}

	return nil
}

func (r *InvoiceRepo) ListInvoices(filter models.InvoiceFilter) ([]*models.Invoice, int64, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if filter.UserID > 0 {
		whereClause += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, filter.UserID)
		argCount++
	}

	if filter.RecipientID > 0 {
		whereClause += fmt.Sprintf(" AND recipient_id = $%d", argCount)
		args = append(args, filter.RecipientID)
		argCount++
	}

	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, filter.Status)
		argCount++
	}

	if !filter.StartDate.IsZero() {
		whereClause += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, filter.StartDate)
		argCount++
	}

	if !filter.EndDate.IsZero() {
		whereClause += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, filter.EndDate)
		argCount++
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM invoices %s", whereClause)
	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.ListInvoices: error counting invoices")
		return nil, 0, err
	}

	limit := int64(10)
	offset := int64(0)
	if filter.PageSize > 0 {
		limit = filter.PageSize
	}
	if filter.Page > 0 {
		offset = (filter.Page - 1) * limit
	}

	query := fmt.Sprintf(`
		SELECT id, invoice_number, user_id, recipient_id, recipient_name, recipient_account,
			total_amount, currency, description, status, due_date, created_at, updated_at,
			paid_at, payment_method, payment_card_id, payment_account_id
		FROM invoices
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argCount, argCount+1)

	args = append(args, limit, offset)
	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.ListInvoices: error querying invoices")
		return nil, 0, err
	}
	defer rows.Close()

	var invoices []*models.Invoice
	for rows.Next() {
		var invoice models.Invoice
		var paidAt sql.NullTime
		var paymentMethod sql.NullString
		var paymentCardID, paymentAccountID sql.NullInt64

		err := rows.Scan(
			&invoice.ID,
			&invoice.InvoiceNumber,
			&invoice.UserID,
			&invoice.RecipientID,
			&invoice.RecipientName,
			&invoice.RecipientAccount,
			&invoice.TotalAmount,
			&invoice.Currency,
			&invoice.Description,
			&invoice.Status,
			&invoice.DueDate,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
			&paidAt,
			&paymentMethod,
			&paymentCardID,
			&paymentAccountID,
		)
		if err != nil {
			logrus.WithError(err).Error("InvoiceRepo.ListInvoices: error scanning row")
			return nil, 0, err
		}

		if paidAt.Valid {
			invoice.PaidAt = &paidAt.Time
		}
		if paymentMethod.Valid {
			invoice.PaymentMethod = paymentMethod.String
		}
		if paymentCardID.Valid {
			invoice.PaymentCardID = &paymentCardID.Int64
		}
		if paymentAccountID.Valid {
			invoice.PaymentAccountID = &paymentAccountID.Int64
		}

		invoices = append(invoices, &invoice)
	}

	return invoices, totalCount, nil
}

func (r *InvoiceRepo) GetUserInvoices(userID int64, page, pageSize int64) ([]*models.Invoice, int64, error) {
	filter := models.InvoiceFilter{
		UserID:   userID,
		Page:     page,
		PageSize: pageSize,
	}
	return r.ListInvoices(filter)
}

func (r *InvoiceRepo) GenerateInvoiceNumber() (string, error) {
	rand.Seed(time.Now().UnixNano())
	prefix := "INV"
	randomNum := rand.Intn(100000)

	// Format: INV-{TIMESTAMP}-{RANDOM}
	timestamp := time.Now().Format("20060102")
	invoiceNumber := fmt.Sprintf("%s-%s-%05d", prefix, timestamp, randomNum)

	query := "SELECT COUNT(*) FROM invoices WHERE invoice_number = $1"
	var count int
	err := r.db.QueryRow(query, invoiceNumber).Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.GenerateInvoiceNumber: error checking invoice number")
		return "", err
	}

	if count > 0 {
		return r.GenerateInvoiceNumber()
	}

	return invoiceNumber, nil
}

func (r *InvoiceRepo) CreateInvoiceItems(items []models.InvoiceItem) error {
	tx, err := r.db.Beginx()
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.CreateInvoiceItems: error starting transaction")
		return err
	}

	err = r.createInvoiceItemsTx(tx, items)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (r *InvoiceRepo) createInvoiceItemsTx(tx *sqlx.Tx, items []models.InvoiceItem) error {
	query := `
		INSERT INTO invoice_items (
			invoice_id, description, quantity, unit_price, amount
		) VALUES (
			$1, $2, $3, $4, $5
		)`

	for _, item := range items {
		_, err := tx.Exec(
			query,
			item.InvoiceID,
			item.Description,
			item.Quantity,
			item.UnitPrice,
			item.Amount,
		)
		if err != nil {
			logrus.WithError(err).Errorf("InvoiceRepo.createInvoiceItemsTx: error creating item for invoice %d", item.InvoiceID)
			return err
		}
	}

	return nil
}

func (r *InvoiceRepo) GetInvoiceItems(invoiceID int64) ([]models.InvoiceItem, error) {
	query := `
		SELECT id, invoice_id, description, quantity, unit_price, amount
		FROM invoice_items
		WHERE invoice_id = $1
		ORDER BY id`

	rows, err := r.db.Query(query, invoiceID)
	if err != nil {
		logrus.WithError(err).Errorf("InvoiceRepo.GetInvoiceItems: error getting items for invoice %d", invoiceID)
		return nil, err
	}
	defer rows.Close()

	var items []models.InvoiceItem
	for rows.Next() {
		var item models.InvoiceItem
		err := rows.Scan(
			&item.ID,
			&item.InvoiceID,
			&item.Description,
			&item.Quantity,
			&item.UnitPrice,
			&item.Amount,
		)
		if err != nil {
			logrus.WithError(err).Error("InvoiceRepo.GetInvoiceItems: error scanning row")
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *InvoiceRepo) ProcessInvoicePayment(invoiceID int64, paymentMethod string, accountOrCardID int64) error {
	invoice, err := r.GetInvoiceByID(invoiceID)
	if err != nil {
		return err
	}

	if invoice.Status == models.InvoiceStatusPaid {
		return fmt.Errorf("invoice is already paid")
	}

	if invoice.Status == models.InvoiceStatusCanceled {
		return fmt.Errorf("invoice is cancelled and cannot be paid")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.ProcessInvoicePayment: error starting transaction")
		return err
	}

	if paymentMethod == models.PaymentMethodAccount {
		query := "SELECT user_id FROM accounts WHERE id = $1"
		var userID int64
		err = tx.QueryRow(query, accountOrCardID).Scan(&userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error getting account details: %w", err)
		}

		if userID != invoice.UserID {
			tx.Rollback()
			return fmt.Errorf("account does not belong to the user")
		}

		query = "UPDATE accounts SET amount = amount - $1 WHERE id = $2 AND amount >= $1"
		result, err := tx.Exec(query, invoice.TotalAmount, accountOrCardID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error processing account payment: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil || affected == 0 {
			tx.Rollback()
			return fmt.Errorf("insufficient funds in account")
		}

		query = `
			UPDATE invoices
			SET status = $1, payment_method = $2, payment_account_id = $3, paid_at = $4
			WHERE id = $5`

		_, err = tx.Exec(query, models.InvoiceStatusPaid, paymentMethod, accountOrCardID, time.Now(), invoiceID)

	} else if paymentMethod == models.PaymentMethodCard {
		query := "SELECT account_id FROM cards WHERE id = $1"
		var accountID int64
		err = tx.QueryRow(query, accountOrCardID).Scan(&accountID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error getting card details: %w", err)
		}

		query = "SELECT user_id FROM accounts WHERE id = $1"
		var userID int64
		err = tx.QueryRow(query, accountID).Scan(&userID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error getting account details: %w", err)
		}

		if userID != invoice.UserID {
			tx.Rollback()
			return fmt.Errorf("card does not belong to the user")
		}

		query = "UPDATE accounts SET amount = amount - $1 WHERE id = $2 AND amount >= $1"
		result, err := tx.Exec(query, invoice.TotalAmount, accountID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error processing card payment: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil || affected == 0 {
			tx.Rollback()
			return fmt.Errorf("insufficient funds in account linked to card")
		}

		query = `
			UPDATE invoices
			SET status = $1, payment_method = $2, payment_card_id = $3, paid_at = $4
			WHERE id = $5`

		_, err = tx.Exec(query, models.InvoiceStatusPaid, paymentMethod, accountOrCardID, time.Now(), invoiceID)

	} else {
		tx.Rollback()
		return fmt.Errorf("unsupported payment method: %s", paymentMethod)
	}

	if err != nil {
		tx.Rollback()
		logrus.WithError(err).Errorf("InvoiceRepo.ProcessInvoicePayment: error updating invoice %d", invoiceID)
		return err
	}

	err = tx.Commit()
	if err != nil {
		logrus.WithError(err).Error("InvoiceRepo.ProcessInvoicePayment: error committing transaction")
		return err
	}

	return nil
}
