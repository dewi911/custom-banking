package repository

import (
	"context"
	"custom-banking/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"time"
)

type StakingRepo struct {
	db *sqlx.DB
}

func NewStaking(db *sqlx.DB) *StakingRepo {
	return &StakingRepo{db: db}
}

func (r *StakingRepo) Create(ctx context.Context, staking *models.Staking) (int64, error) {
	query := `
		INSERT INTO staking 
		(user_id, amount, currency_id, start_date, end_date, interest_rate, status) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) 
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		staking.UserID,
		staking.Amount,
		staking.CurrencyID,
		staking.StartDate,
		staking.EndDate,
		staking.InterestRate,
		staking.Status,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("StakingRepo.Create: error creating staking")
		return 0, err
	}

	return id, nil
}

func (r *StakingRepo) GetByID(ctx context.Context, id int64) (*models.Staking, error) {
	query := `
		SELECT s.*, c.code as currency_code
		FROM staking s
		JOIN currency c ON s.currency_id = c.id
		WHERE s.id = $1
	`

	staking := &models.Staking{}
	err := r.db.QueryRow(query, id).Scan(
		&staking.ID,
		&staking.UserID,
		&staking.Amount,
		&staking.CurrencyID,
		&staking.StartDate,
		&staking.EndDate,
		&staking.InterestRate,
		&staking.Status,
		&staking.CurrencyCode,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("staking with id %d not found", id)
		}
		logrus.WithError(err).Errorf("StakingRepo.GetByID: error getting staking with id %d", id)
		return nil, err
	}

	earnedInterest, err := r.CalculateEarnedInterest(ctx, id)
	if err != nil {
		logrus.WithError(err).Errorf("StakingRepo.GetByID: error calculating earned interest for staking %d", id)
	} else {
		staking.EarnedInterest = earnedInterest
		staking.TotalReturn = staking.Amount + earnedInterest
	}

	now := time.Now()
	if now.Before(staking.EndDate) {
		staking.DaysRemaining = int(staking.EndDate.Sub(now).Hours() / 24)
	} else {
		staking.DaysRemaining = 0
	}

	return staking, nil
}

func (r *StakingRepo) GetByUserID(ctx context.Context, userID int64) ([]*models.Staking, error) {
	query := `
		SELECT s.*, c.code as currency_code
		FROM staking s
		JOIN currency c ON s.currency_id = c.id
		WHERE s.user_id = $1
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		logrus.WithError(err).Errorf("StakingRepo.GetByUserID: error getting staking for user %d", userID)
		return nil, err
	}
	defer rows.Close()

	var stakings []*models.Staking
	for rows.Next() {
		staking := &models.Staking{}
		err := rows.Scan(
			&staking.ID,
			&staking.UserID,
			&staking.Amount,
			&staking.CurrencyID,
			&staking.StartDate,
			&staking.EndDate,
			&staking.InterestRate,
			&staking.Status,
			&staking.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("StakingRepo.GetByUserID: error scanning staking row")
			return nil, err
		}
		stakings = append(stakings, staking)
	}

	return stakings, nil
}

func (r *StakingRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE staking SET status = $1 WHERE id = $2`

	res, err := r.db.Exec(query, status, id)
	if err != nil {
		logrus.WithError(err).Errorf("StakingRepo.UpdateStatus: error updating status for staking %d", id)
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		logrus.WithError(err).Error("StakingRepo.UpdateStatus: error getting rows affected")
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("staking with id %d not found", id)
	}

	return nil
}

func (r *StakingRepo) List(ctx context.Context, params models.StakingListParams) ([]*models.Staking, int, error) {
	whereClause := "WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	if params.UserID != 0 {
		whereClause += fmt.Sprintf(" AND s.user_id = $%d", argCount)
		args = append(args, params.UserID)
		argCount++
	}

	if params.Status != "" {
		whereClause += fmt.Sprintf(" AND s.status = $%d", argCount)
		args = append(args, params.Status)
		argCount++
	}

	countQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM staking s
		%s
	`, whereClause)

	var totalCount int
	err := r.db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		logrus.WithError(err).Error("StakingRepo.List: error counting stakings")
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PageSize

	query := fmt.Sprintf(`
		SELECT s.*, c.code as currency_code
		FROM staking s
		JOIN currency c ON s.currency_id = c.id
		%s
		ORDER BY s.start_date DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argCount, argCount+1)

	args = append(args, params.PageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("StakingRepo.List: error getting stakings")
		return nil, 0, err
	}
	defer rows.Close()

	var stakings []*models.Staking
	for rows.Next() {
		staking := &models.Staking{}
		err := rows.Scan(
			&staking.ID,
			&staking.UserID,
			&staking.Amount,
			&staking.CurrencyID,
			&staking.StartDate,
			&staking.EndDate,
			&staking.InterestRate,
			&staking.Status,
			&staking.CurrencyCode,
		)
		if err != nil {
			logrus.WithError(err).Error("StakingRepo.List: error scanning staking row")
			return nil, 0, err
		}
		stakings = append(stakings, staking)
	}

	return stakings, totalCount, nil
}

func (r *StakingRepo) CalculateDailyInterests(ctx context.Context) error {
	checkQuery := `
        SELECT COUNT(*) 
        FROM staking
        WHERE status = 'active' 
        AND start_date <= CURRENT_DATE 
        AND (end_date IS NULL OR end_date >= CURRENT_DATE)
    `

	var count int
	err := r.db.QueryRowContext(ctx, checkQuery).Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("StakingRepo.CalculateDailyInterests: error checking active stakings")
		return err
	}

	logrus.Infof("Found %d active stakings for interest calculation", count)

	if count == 0 {
		logrus.Warn("No active stakings found for interest calculation")
		return nil
	}

	_, err = r.db.ExecContext(ctx, "SELECT * FROM staking_interests LIMIT 0")
	if err != nil {
		logrus.WithError(err).Error("Error accessing staking_interests table")
		return err
	}

	query := `
        INSERT INTO staking_interests (staking_id, amount, date_calculated, description)
        SELECT 
            id as staking_id,
            amount * (interest_rate / 100 / 365) as amount,
            CURRENT_TIMESTAMP as date,
            'Daily interest accrual' as description
        FROM staking
        WHERE status = 'active' 
        AND start_date <= CURRENT_DATE 
        AND (end_date IS NULL OR end_date >= CURRENT_DATE)
    `

	result, err := r.db.ExecContext(ctx, query)
	if err != nil {
		logrus.WithError(err).Error("StakingRepo.CalculateDailyInterests: error calculating daily interests")
		return err
	}

	affected, _ := result.RowsAffected()
	logrus.Infof("Inserted %d interest records", affected)

	return nil
}

func (r *StakingRepo) CreateInterest(ctx context.Context, interest *models.StakingInterest) (int64, error) {
	query := `
		INSERT INTO staking_interests
		(staking_id, amount, date_calculated, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	var id int64
	err := r.db.QueryRow(
		query,
		interest.StakingID,
		interest.Amount,
		interest.DateCalculated,
		interest.Description,
	).Scan(&id)

	if err != nil {
		logrus.WithError(err).Error("StakingRepo.CreateInterest: error creating staking interest")
		return 0, err
	}

	return id, nil
}

func (r *StakingRepo) GetInterestsByStakingID(ctx context.Context, stakingID int64) ([]*models.StakingInterest, error) {
	query := `
		SELECT id, staking_id, amount, date_calculated, description
		FROM staking_interests
		WHERE staking_id = $1
		ORDER BY date_calculated DESC
	`

	rows, err := r.db.Query(query, stakingID)
	if err != nil {
		logrus.WithError(err).Errorf("StakingRepo.GetInterestsByStakingID: error getting interests for staking %d", stakingID)
		return nil, err
	}
	defer rows.Close()

	var interests []*models.StakingInterest
	for rows.Next() {
		interest := &models.StakingInterest{}
		err := rows.Scan(
			&interest.ID,
			&interest.StakingID,
			&interest.Amount,
			&interest.DateCalculated,
			&interest.Description,
		)
		if err != nil {
			logrus.WithError(err).Error("StakingRepo.GetInterestsByStakingID: error scanning interest row")
			return nil, err
		}
		interests = append(interests, interest)
	}

	return interests, nil
}

func (r *StakingRepo) CalculateEarnedInterest(ctx context.Context, stakingID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM staking_interests
		WHERE staking_id = $1
	`

	var earnedInterest float64
	err := r.db.QueryRow(query, stakingID).Scan(&earnedInterest)
	if err != nil {
		logrus.WithError(err).Errorf("StakingRepo.CalculateEarnedInterest: error calculating earned interest for staking %d", stakingID)
		return 0, err
	}

	return earnedInterest, nil
}
