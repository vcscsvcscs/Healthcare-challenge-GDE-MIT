package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// OperationType represents the type of operation performed
type OperationType string

const (
	OperationCreate OperationType = "CREATE"
	OperationUpdate OperationType = "UPDATE"
	OperationDelete OperationType = "DELETE"
	OperationRead   OperationType = "READ"
)

// ResourceType represents the type of resource being accessed
type ResourceType string

const (
	ResourceHealthCheckIn     ResourceType = "health_check_in"
	ResourceMedication        ResourceType = "medication"
	ResourceMenstruationCycle ResourceType = "menstruation_cycle"
	ResourceBloodPressure     ResourceType = "blood_pressure_reading"
	ResourceFitnessData       ResourceType = "fitness_data"
	ResourceReport            ResourceType = "report"
	ResourceSession           ResourceType = "check_in_session"
	ResourceUser              ResourceType = "user"
)

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             string
	UserID         string
	OperationType  OperationType
	ResourceType   ResourceType
	ResourceID     string
	Timestamp      time.Time
	IPAddress      string
	UserAgent      string
	AdditionalData map[string]interface{}
}

// Logger handles audit logging
type Logger struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewLogger creates a new audit logger
func NewLogger(db *pgxpool.Pool, logger *zap.Logger) *Logger {
	return &Logger{
		db:     db,
		logger: logger,
	}
}

// Log creates an audit log entry
// Validates: Requirements 10.5
func (l *Logger) Log(ctx context.Context, entry AuditLog) error {
	// Set timestamp if not provided
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Log to structured logger first
	l.logger.Info("Audit log entry",
		zap.String("user_id", entry.UserID),
		zap.String("operation", string(entry.OperationType)),
		zap.String("resource_type", string(entry.ResourceType)),
		zap.String("resource_id", entry.ResourceID),
		zap.Time("timestamp", entry.Timestamp),
		zap.String("ip_address", entry.IPAddress),
	)

	// Store in database
	query := `
		INSERT INTO audit_logs (
			user_id, operation_type, resource_type, resource_id, 
			timestamp, ip_address, user_agent, additional_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := l.db.Exec(ctx, query,
		entry.UserID,
		entry.OperationType,
		entry.ResourceType,
		entry.ResourceID,
		entry.Timestamp,
		entry.IPAddress,
		entry.UserAgent,
		entry.AdditionalData,
	)

	if err != nil {
		l.logger.Error("Failed to write audit log to database",
			zap.Error(err),
			zap.String("user_id", entry.UserID),
			zap.String("operation", string(entry.OperationType)),
			zap.String("resource_type", string(entry.ResourceType)),
		)
		return err
	}

	return nil
}

// LogCreate logs a CREATE operation
func (l *Logger) LogCreate(ctx context.Context, userID, resourceType, resourceID, ipAddress, userAgent string) error {
	return l.Log(ctx, AuditLog{
		UserID:        userID,
		OperationType: OperationCreate,
		ResourceType:  ResourceType(resourceType),
		ResourceID:    resourceID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})
}

// LogUpdate logs an UPDATE operation
func (l *Logger) LogUpdate(ctx context.Context, userID, resourceType, resourceID, ipAddress, userAgent string) error {
	return l.Log(ctx, AuditLog{
		UserID:        userID,
		OperationType: OperationUpdate,
		ResourceType:  ResourceType(resourceType),
		ResourceID:    resourceID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})
}

// LogDelete logs a DELETE operation
func (l *Logger) LogDelete(ctx context.Context, userID, resourceType, resourceID, ipAddress, userAgent string) error {
	return l.Log(ctx, AuditLog{
		UserID:        userID,
		OperationType: OperationDelete,
		ResourceType:  ResourceType(resourceType),
		ResourceID:    resourceID,
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
	})
}

// GetAuditLogs retrieves audit logs for a user
func (l *Logger) GetAuditLogs(ctx context.Context, userID string, limit int) ([]AuditLog, error) {
	query := `
		SELECT user_id, operation_type, resource_type, resource_id, 
		       timestamp, ip_address, user_agent
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`

	rows, err := l.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		err := rows.Scan(
			&log.UserID,
			&log.OperationType,
			&log.ResourceType,
			&log.ResourceID,
			&log.Timestamp,
			&log.IPAddress,
			&log.UserAgent,
		)
		if err != nil {
			l.logger.Error("Failed to scan audit log", zap.Error(err))
			continue
		}
		logs = append(logs, log)
	}

	return logs, nil
}
