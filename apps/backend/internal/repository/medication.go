package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// MedicationRepository manages medication data
type MedicationRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewMedicationRepository creates a new MedicationRepository
func NewMedicationRepository(db *pgxpool.Pool, logger *zap.Logger) *MedicationRepository {
	return &MedicationRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new medication record
func (r *MedicationRepository) Create(ctx context.Context, med *model.Medication) error {
	query := `
		INSERT INTO medications (
			id, user_id, name, dosage, frequency,
			start_date, end_date, notes, active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
	`

	_, err := r.db.Exec(ctx, query,
		med.ID,
		med.UserID,
		med.Name,
		med.Dosage,
		med.Frequency,
		med.StartDate,
		med.EndDate,
		med.Notes,
		med.Active,
	)

	if err != nil {
		r.logger.Error("failed to create medication",
			zap.Error(err),
			zap.String("medication_id", med.ID),
			zap.String("user_id", med.UserID),
		)
		return fmt.Errorf("failed to create medication: %w", err)
	}

	return nil
}

// FindByUserID retrieves all medications for a user, sorted by start date
func (r *MedicationRepository) FindByUserID(ctx context.Context, userID string) ([]model.Medication, error) {
	query := `
		SELECT 
			id, user_id, name, dosage, frequency,
			start_date, end_date, notes, active,
			created_at, updated_at
		FROM medications
		WHERE user_id = $1
		ORDER BY start_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to find medications", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to find medications: %w", err)
	}
	defer rows.Close()

	var medications []model.Medication
	for rows.Next() {
		var med model.Medication
		err := rows.Scan(
			&med.ID,
			&med.UserID,
			&med.Name,
			&med.Dosage,
			&med.Frequency,
			&med.StartDate,
			&med.EndDate,
			&med.Notes,
			&med.Active,
			&med.CreatedAt,
			&med.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan medication", zap.Error(err))
			continue
		}
		medications = append(medications, med)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating medications", zap.Error(err))
		return nil, fmt.Errorf("error iterating medications: %w", err)
	}

	return medications, nil
}

// FindByID retrieves a medication by ID
func (r *MedicationRepository) FindByID(ctx context.Context, medicationID string) (*model.Medication, error) {
	query := `
		SELECT 
			id, user_id, name, dosage, frequency,
			start_date, end_date, notes, active,
			created_at, updated_at
		FROM medications
		WHERE id = $1
	`

	var med model.Medication
	err := r.db.QueryRow(ctx, query, medicationID).Scan(
		&med.ID,
		&med.UserID,
		&med.Name,
		&med.Dosage,
		&med.Frequency,
		&med.StartDate,
		&med.EndDate,
		&med.Notes,
		&med.Active,
		&med.CreatedAt,
		&med.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("medication not found: %s", medicationID)
		}
		r.logger.Error("failed to find medication", zap.Error(err), zap.String("medication_id", medicationID))
		return nil, fmt.Errorf("failed to find medication: %w", err)
	}

	return &med, nil
}

// Update updates an existing medication record
func (r *MedicationRepository) Update(ctx context.Context, med *model.Medication) error {
	query := `
		UPDATE medications
		SET name = $1, dosage = $2, frequency = $3,
		    start_date = $4, end_date = $5, notes = $6,
		    active = $7, updated_at = NOW()
		WHERE id = $8
	`

	result, err := r.db.Exec(ctx, query,
		med.Name,
		med.Dosage,
		med.Frequency,
		med.StartDate,
		med.EndDate,
		med.Notes,
		med.Active,
		med.ID,
	)

	if err != nil {
		r.logger.Error("failed to update medication",
			zap.Error(err),
			zap.String("medication_id", med.ID),
		)
		return fmt.Errorf("failed to update medication: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("medication not found: %s", med.ID)
	}

	return nil
}

// Delete deletes a medication record
func (r *MedicationRepository) Delete(ctx context.Context, medicationID string) error {
	query := `DELETE FROM medications WHERE id = $1`

	result, err := r.db.Exec(ctx, query, medicationID)
	if err != nil {
		r.logger.Error("failed to delete medication",
			zap.Error(err),
			zap.String("medication_id", medicationID),
		)
		return fmt.Errorf("failed to delete medication: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("medication not found: %s", medicationID)
	}

	return nil
}

// LogAdherence logs medication adherence
func (r *MedicationRepository) LogAdherence(ctx context.Context, log *model.MedicationLog) error {
	query := `
		INSERT INTO medication_logs (id, medication_id, taken_at, adherence, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`

	_, err := r.db.Exec(ctx, query,
		log.ID,
		log.MedicationID,
		log.TakenAt,
		log.Adherence,
	)

	if err != nil {
		r.logger.Error("failed to log medication adherence",
			zap.Error(err),
			zap.String("medication_id", log.MedicationID),
		)
		return fmt.Errorf("failed to log medication adherence: %w", err)
	}

	return nil
}

// GetAdherenceLogs retrieves adherence logs for a medication
func (r *MedicationRepository) GetAdherenceLogs(ctx context.Context, medicationID string) ([]model.MedicationLog, error) {
	query := `
		SELECT id, medication_id, taken_at, adherence, created_at
		FROM medication_logs
		WHERE medication_id = $1
		ORDER BY taken_at DESC
	`

	rows, err := r.db.Query(ctx, query, medicationID)
	if err != nil {
		r.logger.Error("failed to get adherence logs", zap.Error(err), zap.String("medication_id", medicationID))
		return nil, fmt.Errorf("failed to get adherence logs: %w", err)
	}
	defer rows.Close()

	var logs []model.MedicationLog
	for rows.Next() {
		var log model.MedicationLog
		err := rows.Scan(
			&log.ID,
			&log.MedicationID,
			&log.TakenAt,
			&log.Adherence,
			&log.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan adherence log", zap.Error(err))
			continue
		}
		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating adherence logs", zap.Error(err))
		return nil, fmt.Errorf("error iterating adherence logs: %w", err)
	}

	return logs, nil
}
