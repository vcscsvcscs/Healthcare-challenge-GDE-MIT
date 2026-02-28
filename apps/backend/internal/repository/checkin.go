package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// CheckInRepository manages check-in session data
type CheckInRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewCheckInRepository creates a new CheckInRepository
func NewCheckInRepository(db *pgxpool.Pool, logger *zap.Logger) *CheckInRepository {
	return &CheckInRepository{
		db:     db,
		logger: logger,
	}
}

// CreateSession creates a new check-in session
func (r *CheckInRepository) CreateSession(ctx context.Context, session *model.Session) error {
	query := `
		INSERT INTO check_in_sessions (id, user_id, started_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
	`

	_, err := r.db.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.StartedAt,
		session.Status,
	)

	if err != nil {
		r.logger.Error("failed to create session", zap.Error(err), zap.String("session_id", session.ID))
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSession retrieves a session by ID
func (r *CheckInRepository) GetSession(ctx context.Context, sessionID string) (*model.Session, error) {
	query := `
		SELECT id, user_id, started_at, completed_at, expired_at, status, created_at, updated_at
		FROM check_in_sessions
		WHERE id = $1
	`

	var session model.Session
	var createdAt, updatedAt time.Time
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.StartedAt,
		&session.CompletedAt,
		&session.ExpiredAt,
		&session.Status,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("session not found: %s", sessionID)
		}
		r.logger.Error("failed to get session", zap.Error(err), zap.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (r *CheckInRepository) UpdateSession(ctx context.Context, session *model.Session) error {
	query := `
		UPDATE check_in_sessions
		SET completed_at = $1, expired_at = $2, status = $3, updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query,
		session.CompletedAt,
		session.ExpiredAt,
		session.Status,
		session.ID,
	)

	if err != nil {
		r.logger.Error("failed to update session", zap.Error(err), zap.String("session_id", session.ID))
		return fmt.Errorf("failed to update session: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("session not found: %s", session.ID)
	}

	return nil
}

// SaveConversationMessage saves a conversation message
func (r *CheckInRepository) SaveConversationMessage(ctx context.Context, msg *model.Message) error {
	query := `
		INSERT INTO conversation_messages (id, session_id, role, content, audio_file_path, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := r.db.Exec(ctx, query,
		msg.ID,
		msg.SessionID,
		msg.Role,
		msg.Content,
		msg.AudioFilePath,
		msg.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to save conversation message",
			zap.Error(err),
			zap.String("session_id", msg.SessionID),
			zap.String("message_id", msg.ID),
		)
		return fmt.Errorf("failed to save conversation message: %w", err)
	}

	return nil
}

// GetConversationMessages retrieves all messages for a session
func (r *CheckInRepository) GetConversationMessages(ctx context.Context, sessionID string) ([]model.Message, error) {
	query := `
		SELECT id, session_id, role, content, audio_file_path, created_at
		FROM conversation_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(ctx, query, sessionID)
	if err != nil {
		r.logger.Error("failed to get conversation messages", zap.Error(err), zap.String("session_id", sessionID))
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	defer rows.Close()

	var messages []model.Message
	for rows.Next() {
		var msg model.Message
		err := rows.Scan(
			&msg.ID,
			&msg.SessionID,
			&msg.Role,
			&msg.Content,
			&msg.AudioFilePath,
			&msg.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan message", zap.Error(err))
			continue
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating messages", zap.Error(err))
		return nil, fmt.Errorf("error iterating messages: %w", err)
	}

	return messages, nil
}

// SaveHealthCheckIn saves a completed health check-in
func (r *CheckInRepository) SaveHealthCheckIn(ctx context.Context, checkIn *model.HealthCheckIn) error {
	query := `
		INSERT INTO health_check_ins (
			id, user_id, session_id, check_in_date,
			symptoms, mood, pain_level, energy_level, sleep_quality,
			medication_taken, physical_activity,
			breakfast, lunch, dinner,
			general_feeling, additional_notes, raw_transcript,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4,
			$5, $6, $7, $8, $9,
			$10, $11,
			$12, $13, $14,
			$15, $16, $17,
			NOW(), NOW()
		)
	`

	_, err := r.db.Exec(ctx, query,
		checkIn.ID,
		checkIn.UserID,
		checkIn.SessionID,
		checkIn.CheckInDate,
		checkIn.Symptoms,
		checkIn.Mood,
		checkIn.PainLevel,
		checkIn.EnergyLevel,
		checkIn.SleepQuality,
		checkIn.MedicationTaken,
		checkIn.PhysicalActivity,
		checkIn.Breakfast,
		checkIn.Lunch,
		checkIn.Dinner,
		checkIn.GeneralFeeling,
		checkIn.AdditionalNotes,
		checkIn.RawTranscript,
	)

	if err != nil {
		r.logger.Error("failed to save health check-in",
			zap.Error(err),
			zap.String("check_in_id", checkIn.ID),
			zap.String("user_id", checkIn.UserID),
		)
		return fmt.Errorf("failed to save health check-in: %w", err)
	}

	return nil
}

// GetHealthCheckInsByUserID retrieves health check-ins for a user
func (r *CheckInRepository) GetHealthCheckInsByUserID(ctx context.Context, userID string) ([]model.HealthCheckIn, error) {
	query := `
		SELECT 
			id, user_id, session_id, check_in_date,
			symptoms, mood, pain_level, energy_level, sleep_quality,
			medication_taken, physical_activity,
			breakfast, lunch, dinner,
			general_feeling, additional_notes, raw_transcript,
			created_at, updated_at
		FROM health_check_ins
		WHERE user_id = $1
		ORDER BY check_in_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get health check-ins", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get health check-ins: %w", err)
	}
	defer rows.Close()

	var checkIns []model.HealthCheckIn
	for rows.Next() {
		var checkIn model.HealthCheckIn
		err := rows.Scan(
			&checkIn.ID,
			&checkIn.UserID,
			&checkIn.SessionID,
			&checkIn.CheckInDate,
			&checkIn.Symptoms,
			&checkIn.Mood,
			&checkIn.PainLevel,
			&checkIn.EnergyLevel,
			&checkIn.SleepQuality,
			&checkIn.MedicationTaken,
			&checkIn.PhysicalActivity,
			&checkIn.Breakfast,
			&checkIn.Lunch,
			&checkIn.Dinner,
			&checkIn.GeneralFeeling,
			&checkIn.AdditionalNotes,
			&checkIn.RawTranscript,
			&checkIn.CreatedAt,
			&checkIn.UpdatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan health check-in", zap.Error(err))
			continue
		}
		checkIns = append(checkIns, checkIn)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating health check-ins", zap.Error(err))
		return nil, fmt.Errorf("error iterating health check-ins: %w", err)
	}

	return checkIns, nil
}
