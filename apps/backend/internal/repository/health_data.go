package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// HealthDataRepository manages health metrics data
type HealthDataRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewHealthDataRepository creates a new HealthDataRepository
func NewHealthDataRepository(db *pgxpool.Pool, logger *zap.Logger) *HealthDataRepository {
	return &HealthDataRepository{
		db:     db,
		logger: logger,
	}
}

// SaveMenstruation saves a menstruation cycle record
func (r *HealthDataRepository) SaveMenstruation(ctx context.Context, data *model.MenstruationCycle) error {
	query := `
		INSERT INTO menstruation_cycles (
			id, user_id, start_date, end_date,
			flow_intensity, symptoms,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.db.Exec(ctx, query,
		data.ID,
		data.UserID,
		data.StartDate,
		data.EndDate,
		data.FlowIntensity,
		data.Symptoms,
	)

	if err != nil {
		r.logger.Error("failed to save menstruation data",
			zap.Error(err),
			zap.String("user_id", data.UserID),
		)
		return fmt.Errorf("failed to save menstruation data: %w", err)
	}

	return nil
}

// GetMenstruationByUserID retrieves menstruation cycles for a user, sorted by start date descending
func (r *HealthDataRepository) GetMenstruationByUserID(ctx context.Context, userID string) ([]model.MenstruationCycle, error) {
	query := `
		SELECT 
			id, user_id, start_date, end_date,
			flow_intensity, symptoms,
			created_at, updated_at
		FROM menstruation_cycles
		WHERE user_id = $1
		ORDER BY start_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get menstruation data", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get menstruation data: %w", err)
	}
	defer rows.Close()

	var cycles []model.MenstruationCycle
	for rows.Next() {
		var cycle model.MenstruationCycle
		err := rows.Scan(
			&cycle.ID,
			&cycle.UserID,
			&cycle.StartDate,
			&cycle.EndDate,
			&cycle.FlowIntensity,
			&cycle.Symptoms,
			&cycle.CreatedAt,
			&cycle.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan menstruation cycle", zap.Error(err))
			continue
		}
		cycles = append(cycles, cycle)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating menstruation cycles", zap.Error(err))
		return nil, fmt.Errorf("error iterating menstruation cycles: %w", err)
	}

	return cycles, nil
}

// UpdateMenstruation updates a menstruation cycle record
func (r *HealthDataRepository) UpdateMenstruation(ctx context.Context, data *model.MenstruationCycle) error {
	query := `
		UPDATE menstruation_cycles
		SET end_date = $1, flow_intensity = $2, symptoms = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query,
		data.EndDate,
		data.FlowIntensity,
		data.Symptoms,
		data.ID,
	)

	if err != nil {
		r.logger.Error("failed to update menstruation data",
			zap.Error(err),
			zap.String("cycle_id", data.ID),
		)
		return fmt.Errorf("failed to update menstruation data: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("menstruation cycle not found: %s", data.ID)
	}

	return nil
}

// SaveBloodPressure saves a blood pressure reading
func (r *HealthDataRepository) SaveBloodPressure(ctx context.Context, reading *model.BloodPressureReading) error {
	query := `
		INSERT INTO blood_pressure_readings (
			id, user_id, systolic, diastolic, pulse,
			measured_at, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err := r.db.Exec(ctx, query,
		reading.ID,
		reading.UserID,
		reading.Systolic,
		reading.Diastolic,
		reading.Pulse,
		reading.MeasuredAt,
	)

	if err != nil {
		r.logger.Error("failed to save blood pressure reading",
			zap.Error(err),
			zap.String("user_id", reading.UserID),
		)
		return fmt.Errorf("failed to save blood pressure reading: %w", err)
	}

	return nil
}

// GetBloodPressureByUserID retrieves blood pressure readings for a user, sorted by measured_at descending
func (r *HealthDataRepository) GetBloodPressureByUserID(ctx context.Context, userID string) ([]model.BloodPressureReading, error) {
	query := `
		SELECT 
			id, user_id, systolic, diastolic, pulse,
			measured_at, created_at
		FROM blood_pressure_readings
		WHERE user_id = $1
		ORDER BY measured_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get blood pressure readings", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get blood pressure readings: %w", err)
	}
	defer rows.Close()

	var readings []model.BloodPressureReading
	for rows.Next() {
		var reading model.BloodPressureReading
		err := rows.Scan(
			&reading.ID,
			&reading.UserID,
			&reading.Systolic,
			&reading.Diastolic,
			&reading.Pulse,
			&reading.MeasuredAt,
			&reading.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan blood pressure reading", zap.Error(err))
			continue
		}
		readings = append(readings, reading)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating blood pressure readings", zap.Error(err))
		return nil, fmt.Errorf("error iterating blood pressure readings: %w", err)
	}

	return readings, nil
}

// SaveFitnessData saves a fitness data point
func (r *HealthDataRepository) SaveFitnessData(ctx context.Context, data *model.FitnessDataPoint) error {
	query := `
		INSERT INTO fitness_data (
			id, user_id, date, data_type, value,
			unit, source, source_data_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
	`

	_, err := r.db.Exec(ctx, query,
		data.ID,
		data.UserID,
		data.Date,
		data.DataType,
		data.Value,
		data.Unit,
		data.Source,
		data.SourceDataID,
	)

	if err != nil {
		r.logger.Error("failed to save fitness data",
			zap.Error(err),
			zap.String("user_id", data.UserID),
			zap.String("data_type", data.DataType),
		)
		return fmt.Errorf("failed to save fitness data: %w", err)
	}

	return nil
}

// FitnessDataExists checks if a fitness data point already exists by source_data_id
func (r *HealthDataRepository) FitnessDataExists(ctx context.Context, sourceDataID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM fitness_data WHERE source_data_id = $1)`

	var exists bool
	err := r.db.QueryRow(ctx, query, sourceDataID).Scan(&exists)
	if err != nil {
		r.logger.Error("failed to check fitness data existence",
			zap.Error(err),
			zap.String("source_data_id", sourceDataID),
		)
		return false, fmt.Errorf("failed to check fitness data existence: %w", err)
	}

	return exists, nil
}

// GetFitnessDataByUserID retrieves fitness data for a user within a date range
func (r *HealthDataRepository) GetFitnessDataByUserID(ctx context.Context, userID string, startDate, endDate time.Time) ([]model.FitnessDataPoint, error) {
	query := `
		SELECT 
			id, user_id, date, data_type, value,
			unit, source, source_data_id, created_at
		FROM fitness_data
		WHERE user_id = $1 AND date >= $2 AND date <= $3
		ORDER BY date DESC, data_type ASC
	`

	rows, err := r.db.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		r.logger.Error("failed to get fitness data",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get fitness data: %w", err)
	}
	defer rows.Close()

	var dataPoints []model.FitnessDataPoint
	for rows.Next() {
		var data model.FitnessDataPoint
		err := rows.Scan(
			&data.ID,
			&data.UserID,
			&data.Date,
			&data.DataType,
			&data.Value,
			&data.Unit,
			&data.Source,
			&data.SourceDataID,
			&data.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan fitness data", zap.Error(err))
			continue
		}
		dataPoints = append(dataPoints, data)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating fitness data", zap.Error(err))
		return nil, fmt.Errorf("error iterating fitness data: %w", err)
	}

	return dataPoints, nil
}

// SaveAudioRecording saves an audio recording record
func (r *HealthDataRepository) SaveAudioRecording(ctx context.Context, recording *model.AudioRecording) error {
	query := `
		INSERT INTO audio_recordings (
			id, session_id, message_id, file_path,
			duration_seconds, transcription, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	_, err := r.db.Exec(ctx, query,
		recording.ID,
		recording.SessionID,
		recording.MessageID,
		recording.FilePath,
		recording.DurationSeconds,
		recording.Transcription,
	)

	if err != nil {
		r.logger.Error("failed to save audio recording",
			zap.Error(err),
			zap.String("session_id", recording.SessionID),
		)
		return fmt.Errorf("failed to save audio recording: %w", err)
	}

	return nil
}

// GetAudioRecordingsBySessionID retrieves audio recordings for a session
func (r *HealthDataRepository) GetAudioRecordingsBySessionID(ctx context.Context, sessionID string) ([]model.AudioRecording, error) {
	query := `
		SELECT 
			id, session_id, message_id, file_path,
			duration_seconds, transcription, created_at
		FROM audio_recordings
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		r.logger.Error("failed to get audio recordings", zap.Error(err), zap.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to get audio recordings: %w", err)
	}
	defer rows.Close()

	var recordings []model.AudioRecording
	for rows.Next() {
		var recording model.AudioRecording
		err := rows.Scan(
			&recording.ID,
			&recording.SessionID,
			&recording.MessageID,
			&recording.FilePath,
			&recording.DurationSeconds,
			&recording.Transcription,
			&recording.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan audio recording", zap.Error(err))
			continue
		}
		recordings = append(recordings, recording)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating audio recordings", zap.Error(err))
		return nil, fmt.Errorf("error iterating audio recordings: %w", err)
	}

	return recordings, nil
}
