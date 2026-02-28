package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/audit"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// GDPRService handles GDPR compliance operations
type GDPRService struct {
	db          *pgxpool.Pool
	auditLogger *audit.Logger
	logger      *zap.Logger
}

// NewGDPRService creates a new GDPR service
func NewGDPRService(db *pgxpool.Pool, auditLogger *audit.Logger, logger *zap.Logger) *GDPRService {
	return &GDPRService{
		db:          db,
		auditLogger: auditLogger,
		logger:      logger,
	}
}

// UserDataExport represents all user data for export
type UserDataExport struct {
	User                  *model.User                  `json:"user"`
	HealthCheckIns        []model.HealthCheckIn        `json:"health_check_ins"`
	Medications           []model.Medication           `json:"medications"`
	MenstruationCycles    []model.MenstruationCycle    `json:"menstruation_cycles"`
	BloodPressureReadings []model.BloodPressureReading `json:"blood_pressure_readings"`
	FitnessData           []model.FitnessDataPoint     `json:"fitness_data"`
	Reports               []model.Report               `json:"reports"`
	ExportedAt            time.Time                    `json:"exported_at"`
}

// DeleteUserData deletes all user data (GDPR right to be forgotten)
// Validates: Requirements 10.3
func (s *GDPRService) DeleteUserData(ctx context.Context, userID, ipAddress, userAgent string) error {
	s.logger.Info("Starting user data deletion (GDPR)",
		zap.String("user_id", userID),
	)

	// Start transaction
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete health check-ins
	_, err = tx.Exec(ctx, "DELETE FROM health_check_ins WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete health check-ins: %w", err)
	}

	// Delete medications
	_, err = tx.Exec(ctx, "DELETE FROM medications WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete medications: %w", err)
	}

	// Delete menstruation cycles
	_, err = tx.Exec(ctx, "DELETE FROM menstruation_cycles WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete menstruation cycles: %w", err)
	}

	// Delete blood pressure readings
	_, err = tx.Exec(ctx, "DELETE FROM blood_pressure_readings WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete blood pressure readings: %w", err)
	}

	// Delete fitness data
	_, err = tx.Exec(ctx, "DELETE FROM fitness_data WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete fitness data: %w", err)
	}

	// Delete reports
	_, err = tx.Exec(ctx, "DELETE FROM reports WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete reports: %w", err)
	}

	// Delete check-in sessions
	_, err = tx.Exec(ctx, "DELETE FROM check_in_sessions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete check-in sessions: %w", err)
	}

	// Mark user as deleted (soft delete to maintain referential integrity in audit logs)
	_, err = tx.Exec(ctx, "UPDATE users SET deleted_at = $1 WHERE id = $2", time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to mark user as deleted: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Log audit entry
	if err := s.auditLogger.LogDelete(ctx, userID, "user", userID, ipAddress, userAgent); err != nil {
		s.logger.Error("Failed to log audit entry for user deletion", zap.Error(err))
	}

	s.logger.Info("User data deletion completed (GDPR)",
		zap.String("user_id", userID),
	)

	return nil
}

// ExportUserData exports all user data to JSON (GDPR right to data portability)
// Validates: Requirements 10.4
func (s *GDPRService) ExportUserData(ctx context.Context, userID string) ([]byte, error) {
	s.logger.Info("Starting user data export (GDPR)",
		zap.String("user_id", userID),
	)

	export := UserDataExport{
		ExportedAt: time.Now(),
	}

	// Get user
	var user model.User
	err := s.db.QueryRow(ctx, `
		SELECT id, name, email, created_at, updated_at, deleted_at
		FROM users WHERE id = $1
	`, userID).Scan(&user.ID, &user.Name, &user.Email, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	export.User = &user

	// Get health check-ins
	checkInRows, err := s.db.Query(ctx, `
		SELECT id, user_id, session_id, check_in_date, symptoms, mood, pain_level,
		       energy_level, sleep_quality, medication_taken, physical_activity,
		       breakfast, lunch, dinner, general_feeling, additional_notes,
		       raw_transcript, created_at, updated_at
		FROM health_check_ins WHERE user_id = $1
		ORDER BY check_in_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health check-ins: %w", err)
	}
	defer checkInRows.Close()

	for checkInRows.Next() {
		var checkIn model.HealthCheckIn
		err := checkInRows.Scan(
			&checkIn.ID, &checkIn.UserID, &checkIn.SessionID, &checkIn.CheckInDate,
			&checkIn.Symptoms, &checkIn.Mood, &checkIn.PainLevel, &checkIn.EnergyLevel,
			&checkIn.SleepQuality, &checkIn.MedicationTaken, &checkIn.PhysicalActivity,
			&checkIn.Breakfast, &checkIn.Lunch, &checkIn.Dinner, &checkIn.GeneralFeeling,
			&checkIn.AdditionalNotes, &checkIn.RawTranscript, &checkIn.CreatedAt, &checkIn.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan health check-in", zap.Error(err))
			continue
		}
		export.HealthCheckIns = append(export.HealthCheckIns, checkIn)
	}

	// Get medications
	medRows, err := s.db.Query(ctx, `
		SELECT id, user_id, name, dosage, frequency, start_date, end_date,
		       notes, active, created_at, updated_at
		FROM medications WHERE user_id = $1
		ORDER BY start_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get medications: %w", err)
	}
	defer medRows.Close()

	for medRows.Next() {
		var med model.Medication
		err := medRows.Scan(
			&med.ID, &med.UserID, &med.Name, &med.Dosage, &med.Frequency,
			&med.StartDate, &med.EndDate, &med.Notes, &med.Active,
			&med.CreatedAt, &med.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan medication", zap.Error(err))
			continue
		}
		export.Medications = append(export.Medications, med)
	}

	// Get menstruation cycles
	cycleRows, err := s.db.Query(ctx, `
		SELECT id, user_id, start_date, end_date, flow_intensity, symptoms,
		       created_at, updated_at
		FROM menstruation_cycles WHERE user_id = $1
		ORDER BY start_date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get menstruation cycles: %w", err)
	}
	defer cycleRows.Close()

	for cycleRows.Next() {
		var cycle model.MenstruationCycle
		err := cycleRows.Scan(
			&cycle.ID, &cycle.UserID, &cycle.StartDate, &cycle.EndDate,
			&cycle.FlowIntensity, &cycle.Symptoms, &cycle.CreatedAt, &cycle.UpdatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan menstruation cycle", zap.Error(err))
			continue
		}
		export.MenstruationCycles = append(export.MenstruationCycles, cycle)
	}

	// Get blood pressure readings
	bpRows, err := s.db.Query(ctx, `
		SELECT id, user_id, systolic, diastolic, pulse, measured_at, created_at
		FROM blood_pressure_readings WHERE user_id = $1
		ORDER BY measured_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get blood pressure readings: %w", err)
	}
	defer bpRows.Close()

	for bpRows.Next() {
		var bp model.BloodPressureReading
		err := bpRows.Scan(
			&bp.ID, &bp.UserID, &bp.Systolic, &bp.Diastolic,
			&bp.Pulse, &bp.MeasuredAt, &bp.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan blood pressure reading", zap.Error(err))
			continue
		}
		export.BloodPressureReadings = append(export.BloodPressureReadings, bp)
	}

	// Get fitness data
	fitnessRows, err := s.db.Query(ctx, `
		SELECT id, user_id, date, data_type, value, unit, source, source_data_id, created_at
		FROM fitness_data WHERE user_id = $1
		ORDER BY date DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fitness data: %w", err)
	}
	defer fitnessRows.Close()

	for fitnessRows.Next() {
		var fitness model.FitnessDataPoint
		err := fitnessRows.Scan(
			&fitness.ID, &fitness.UserID, &fitness.Date, &fitness.DataType,
			&fitness.Value, &fitness.Unit, &fitness.Source, &fitness.SourceDataID,
			&fitness.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan fitness data", zap.Error(err))
			continue
		}
		export.FitnessData = append(export.FitnessData, fitness)
	}

	// Get reports
	reportRows, err := s.db.Query(ctx, `
		SELECT id, user_id, date_range_start, date_range_end, file_path,
		       generated_at, created_at
		FROM reports WHERE user_id = $1
		ORDER BY generated_at DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}
	defer reportRows.Close()

	for reportRows.Next() {
		var report model.Report
		err := reportRows.Scan(
			&report.ID, &report.UserID, &report.DateRangeStart, &report.DateRangeEnd,
			&report.FilePath, &report.GeneratedAt, &report.CreatedAt,
		)
		if err != nil {
			s.logger.Error("Failed to scan report", zap.Error(err))
			continue
		}
		export.Reports = append(export.Reports, report)
	}

	// Convert to JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal export data: %w", err)
	}

	s.logger.Info("User data export completed (GDPR)",
		zap.String("user_id", userID),
		zap.Int("health_check_ins", len(export.HealthCheckIns)),
		zap.Int("medications", len(export.Medications)),
		zap.Int("menstruation_cycles", len(export.MenstruationCycles)),
		zap.Int("blood_pressure_readings", len(export.BloodPressureReadings)),
		zap.Int("fitness_data", len(export.FitnessData)),
		zap.Int("reports", len(export.Reports)),
	)

	return jsonData, nil
}
