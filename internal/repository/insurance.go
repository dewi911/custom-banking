package repository

import (
	"custom-banking/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"math/rand"
	"time"
)

type InsuranceRepo struct {
	db *sqlx.DB
}

func NewInsurance(db *sqlx.DB) *InsuranceRepo {
	return &InsuranceRepo{db: db}
}

func (r *InsuranceRepo) ensureInsuranceTablesExist() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS insurance_policies (
		  id serial PRIMARY KEY,
		  user_id integer,
		  type varchar(50),
		  policy_type varchar(50),
		  insured_item varchar(255),
		  coverage_amount numeric,
		  premium_amount numeric,
		  premium numeric,
		  start_date timestamp,
		  end_date timestamp,
		  status varchar(20),
		  policy_number varchar(50),
		  description text,
		  currency_id integer,
		  payment_account_id integer
		);
		
		CREATE INDEX IF NOT EXISTS idx_insurance_user_id ON insurance_policies (user_id);
	`)

	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.ensureInsuranceTablesExist: error creating insurance_policies table")
		return err
	}

	_, err = r.db.Exec(`
		CREATE TABLE IF NOT EXISTS insurance_claims (
		  id serial PRIMARY KEY,
		  insurance_id integer REFERENCES insurance_policies(id),
		  claim_date timestamp,
		  description text,
		  status varchar(20),
		  amount numeric,
		  filing_date timestamp DEFAULT CURRENT_TIMESTAMP,
		  resolution_date timestamp,
		  document_links text[]
		);
		
		CREATE INDEX IF NOT EXISTS idx_insurance_claims_insurance_id ON insurance_claims (insurance_id);
	`)

	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.ensureInsuranceTablesExist: error creating insurance_claims table")
		return err
	}

	return nil
}

func (r *InsuranceRepo) Create(insurance *models.Insurance) (int64, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return 0, err
	}

	query := `
		INSERT INTO insurance_policies 
		(user_id, type, insured_item, coverage_amount, premium, start_date, end_date, 
		 status, policy_number, description, currency_id, payment_account_id) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) 
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		insurance.UserID,
		insurance.Type,
		insurance.InsuredItem,
		insurance.CoverageAmount,
		insurance.Premium,
		insurance.StartDate,
		insurance.EndDate,
		insurance.Status,
		insurance.PolicyNumber,
		insurance.Description,
		insurance.CurrencyID,
		insurance.PaymentAccountID,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.Create: error creating insurance")
		return 0, err
	}

	return id, nil
}

func (r *InsuranceRepo) GetByID(id int64) (*models.Insurance, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT 
			i.id, i.user_id, i.type, i.insured_item, i.coverage_amount, 
			i.premium, i.start_date, i.end_date, i.status, i.policy_number, 
			i.description, i.currency_id, i.payment_account_id, c.code as currency_code
		FROM insurance_policies i
		JOIN currency c ON i.currency_id = c.id
		WHERE i.id = $1
	`

	insurance := &models.Insurance{}
	err := r.db.QueryRow(query, id).Scan(
		&insurance.ID,
		&insurance.UserID,
		&insurance.Type,
		&insurance.InsuredItem,
		&insurance.CoverageAmount,
		&insurance.Premium,
		&insurance.StartDate,
		&insurance.EndDate,
		&insurance.Status,
		&insurance.PolicyNumber,
		&insurance.Description,
		&insurance.CurrencyID,
		&insurance.PaymentAccountID,
		&insurance.CurrencyCode,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("insurance with id %d not found", id)
		}
		logrus.WithError(err).Errorf("InsuranceRepo.GetByID: error getting insurance with id %d", id)
		return nil, err
	}

	now := time.Now()
	if now.Before(insurance.EndDate) {
		insurance.DaysRemaining = int64(insurance.EndDate.Sub(now).Hours() / 24)
	} else {
		insurance.DaysRemaining = 0
	}

	return insurance, nil
}

func (r *InsuranceRepo) GetByUserID(userID int64) ([]*models.Insurance, error) {
	query := `
		SELECT i.*, c.code as currency_code
		FROM insurance_policies i
		JOIN currency c ON i.currency_id = c.id
		WHERE i.user_id = $1
		ORDER BY i.start_date DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceRepo.GetByUserID: error getting insurance for user %d", userID)
		return nil, err
	}
	defer rows.Close()

	var insurances []*models.Insurance
	for rows.Next() {
		insurance := &models.Insurance{}
		err := rows.Scan(
			&insurance.ID,
			&insurance.UserID,
			&insurance.Type,
			&insurance.InsuredItem,
			&insurance.CoverageAmount,
			&insurance.Premium,
			&insurance.StartDate,
			&insurance.EndDate,
			&insurance.Status,
			&insurance.PolicyNumber,
			&insurance.Description,
			&insurance.CurrencyID,
			&insurance.PaymentAccountID,
			&insurance.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("InsuranceRepo.GetByUserID: error scanning insurance row")
			return nil, err
		}

		now := time.Now()
		if now.Before(insurance.EndDate) {
			insurance.DaysRemaining = int64(insurance.EndDate.Sub(now).Hours() / 24)
		} else {
			insurance.DaysRemaining = 0
		}

		insurances = append(insurances, insurance)
	}

	return insurances, nil
}

func (r *InsuranceRepo) UpdateStatus(id int64, status string) error {
	query := `UPDATE insurance_policies SET status = $1 WHERE id = $2`

	res, err := r.db.Exec(query, status, id)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceRepo.UpdateStatus: error updating status for insurance %d", id)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.UpdateStatus: error getting rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("insurance with id %d not found", id)
	}

	return nil
}

func (r *InsuranceRepo) List(params models.InsuranceListParams) ([]*models.Insurance, int64, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if params.UserID != 0 {
		whereClause += fmt.Sprintf(" AND i.user_id = $%d", argCount)
		args = append(args, params.UserID)
		argCount++
	}

	if params.Type != "" {
		whereClause += fmt.Sprintf(" AND i.type = $%d", argCount)
		args = append(args, params.Type)
		argCount++
	}

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND i.status = $%d", argCount)
		args = append(args, params.Status)
		argCount++
	}

	if params.InsuredItem != "" {
		whereClause += fmt.Sprintf(" AND i.insured_item ILIKE $%d", argCount)
		args = append(args, "%"+params.InsuredItem+"%") // Add wildcards for partial matching
		argCount++
	}

	countQuery := fmt.Sprintf(`
       SELECT COUNT(*)
       FROM insurance_policies i
       %s
    `, whereClause)

	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.List: error counting insurances")
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize

	query := fmt.Sprintf(`
       SELECT 
          i.id,
          i.user_id,
          i.type,
          i.insured_item,
          i.coverage_amount,
          i.premium,
          i.start_date,
          i.end_date,
          i.status,
          i.policy_number,
          i.description,
          i.currency_id,
          i.payment_account_id,
          c.code as currency_code
       FROM insurance_policies i
       JOIN currency c ON i.currency_id = c.id
       %s
       ORDER BY i.start_date DESC
       LIMIT $%d OFFSET $%d
    `, whereClause, argCount, argCount+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.List: error getting insurances")
		return nil, 0, err
	}
	defer rows.Close()

	var insurances []*models.Insurance
	for rows.Next() {
		insurance := &models.Insurance{}
		err := rows.Scan(
			&insurance.ID,
			&insurance.UserID,
			&insurance.Type,
			&insurance.InsuredItem,
			&insurance.CoverageAmount,
			&insurance.Premium,
			&insurance.StartDate,
			&insurance.EndDate,
			&insurance.Status,
			&insurance.PolicyNumber,
			&insurance.Description,
			&insurance.CurrencyID,
			&insurance.PaymentAccountID,
			&insurance.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("InsuranceRepo.List: error scanning insurance row")
			return nil, 0, err
		}

		now := time.Now()
		if now.Before(insurance.EndDate) {
			insurance.DaysRemaining = int64(insurance.EndDate.Sub(now).Hours() / 24)
		} else {
			insurance.DaysRemaining = 0
		}

		insurances = append(insurances, insurance)
	}

	return insurances, totalCount, nil
}

func (r *InsuranceRepo) GeneratePolicyNumber() (string, error) {
	year := time.Now().Year()

	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	randomNum := random.Intn(90000000) + 10000000 // Ensures it's an 8-digit number

	policyNumber := fmt.Sprintf("INS-%d-%08d", year, randomNum)

	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM insurance_policies WHERE policy_number = $1)`
	err := r.db.QueryRow(query, policyNumber).Scan(&exists)
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.GeneratePolicyNumber: error checking policy number existence")
		return "", err
	}

	if exists {
		return r.GeneratePolicyNumber()
	}

	return policyNumber, nil
}

func (r *InsuranceRepo) CreateClaim(claim *models.InsuranceClaim) (int64, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return 0, err
	}

	query := `
		INSERT INTO insurance_claims 
		(insurance_id, claim_date, description, status, amount, filing_date, document_links) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		claim.InsuranceID,
		claim.ClaimDate,
		claim.Description,
		claim.Status,
		claim.Amount,
		claim.FilingDate,
		pq.Array(claim.DocumentLinks),
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.CreateClaim: error creating claim")
		return 0, err
	}

	return id, nil
}

func (r *InsuranceRepo) GetClaimByID(id int64) (*models.InsuranceClaim, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return nil, err
	}

	query := `
		SELECT id, insurance_id, claim_date, description, status, amount, 
		       filing_date, resolution_date, document_links
		FROM insurance_claims
		WHERE id = $1
	`

	claim := &models.InsuranceClaim{}
	var resolutionDate sql.NullTime

	err := r.db.QueryRow(query, id).Scan(
		&claim.ID,
		&claim.InsuranceID,
		&claim.ClaimDate,
		&claim.Description,
		&claim.Status,
		&claim.Amount,
		&claim.FilingDate,
		&resolutionDate,
		pq.Array(&claim.DocumentLinks),
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("claim with id %d not found", id)
		}
		logrus.WithError(err).Errorf("InsuranceRepo.GetClaimByID: error getting claim with id %d", id)
		return nil, err
	}

	if resolutionDate.Valid {
		claim.ResolutionDate = resolutionDate.Time
	}

	return claim, nil
}

func (r *InsuranceRepo) GetClaimsByInsuranceID(insuranceID int64) ([]*models.InsuranceClaim, error) {
	query := `
		SELECT id, insurance_id, claim_date, description, status, amount, 
		       filing_date, resolution_date, document_links
		FROM insurance_claims
		WHERE insurance_id = $1
		ORDER BY filing_date DESC
	`

	rows, err := r.db.Query(query, insuranceID)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceRepo.GetClaimsByInsuranceID: error getting claims for insurance %d", insuranceID)
		return nil, err
	}
	defer rows.Close()

	var claims []*models.InsuranceClaim
	for rows.Next() {
		claim := &models.InsuranceClaim{}
		var resolutionDate sql.NullTime

		err := rows.Scan(
			&claim.ID,
			&claim.InsuranceID,
			&claim.ClaimDate,
			&claim.Description,
			&claim.Status,
			&claim.Amount,
			&claim.FilingDate,
			&resolutionDate,
			pq.Array(&claim.DocumentLinks),
		)
		if err != nil {
			logrus.WithError(err).Error("InsuranceRepo.GetClaimsByInsuranceID: error scanning claim row")
			return nil, err
		}

		if resolutionDate.Valid {
			claim.ResolutionDate = resolutionDate.Time
		}

		claims = append(claims, claim)
	}

	return claims, nil
}

func (r *InsuranceRepo) UpdateClaimStatus(id int64, status string, resolutionDate time.Time) error {
	var query string
	var args []interface{}

	if resolutionDate.IsZero() {
		query = `UPDATE insurance_claims SET status = $1 WHERE id = $2`
		args = []interface{}{status, id}
	} else {
		query = `UPDATE insurance_claims SET status = $1, resolution_date = $2 WHERE id = $3`
		args = []interface{}{status, resolutionDate, id}
	}

	res, err := r.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Errorf("InsuranceRepo.UpdateClaimStatus: error updating status for claim %d", id)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.UpdateClaimStatus: error getting rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("claim with id %d not found", id)
	}

	return nil
}

func (r *InsuranceRepo) ListClaims(insuranceID int64, page, pageSize int64) ([]*models.InsuranceClaim, int64, error) {
	if err := r.ensureInsuranceTablesExist(); err != nil {
		return nil, 0, err
	}

	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if insuranceID != 0 {
		whereClause += fmt.Sprintf(" AND insurance_id = $%d", argCount)
		args = append(args, insuranceID)
		argCount++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM insurance_claims
		%s
	`, whereClause)

	var totalCount int64
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.ListClaims: error counting claims")
		return nil, 0, err
	}

	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
		SELECT id, insurance_id, claim_date, description, status, amount, 
		       filing_date, resolution_date, document_links
		FROM insurance_claims
		%s
		ORDER BY filing_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("InsuranceRepo.ListClaims: error getting claims")
		return nil, 0, err
	}
	defer rows.Close()

	var claims []*models.InsuranceClaim
	for rows.Next() {
		claim := &models.InsuranceClaim{}
		var resolutionDate sql.NullTime

		err := rows.Scan(
			&claim.ID,
			&claim.InsuranceID,
			&claim.ClaimDate,
			&claim.Description,
			&claim.Status,
			&claim.Amount,
			&claim.FilingDate,
			&resolutionDate,
			pq.Array(&claim.DocumentLinks),
		)
		if err != nil {
			logrus.WithError(err).Error("InsuranceRepo.ListClaims: error scanning claim row")
			return nil, 0, err
		}

		if resolutionDate.Valid {
			claim.ResolutionDate = resolutionDate.Time
		}

		claims = append(claims, claim)
	}

	return claims, totalCount, nil
}
