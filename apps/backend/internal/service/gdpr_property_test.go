package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/audit"
	"go.uber.org/zap"
)

// setupTestDB creates a PostgreSQL testcontainer and returns the connection pool
func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	postgresContainer, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("eva_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(t, err)

	// Get connection string
	connString, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Run migrations
	runMigrations(t, pool)

	cleanup := func() {
		pool.Close()
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	}

	return pool, cleanup
}

// runMigrations runs the database migrations for GDPR tests
func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS check_in_sessions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			started_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP,
			expired_at TIMESTAMP,
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS health_check_ins (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			session_id UUID REFERENCES check_in_sessions(id) ON DELETE SET NULL,
			check_in_date DATE NOT NULL,
			symptoms TEXT[],
			mood VARCHAR(50),
			pain_level INTEGER,
			energy_level VARCHAR(50),
			sleep_quality VARCHAR(50),
			medication_taken VARCHAR(50),
			physical_activity TEXT[],
			breakfast TEXT,
			lunch TEXT,
			dinner TEXT,
			general_feeling TEXT,
			additional_notes TEXT,
			raw_transcript TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS medications (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			name VARCHAR(255) NOT NULL,
			dosage VARCHAR(255) NOT NULL,
			frequency VARCHAR(255) NOT NULL,
			start_date DATE NOT NULL,
			end_date DATE,
			notes TEXT,
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS menstruation_cycles (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			start_date DATE NOT NULL,
			end_date DATE,
			flow_intensity VARCHAR(50),
			symptoms TEXT[],
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS blood_pressure_readings (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			systolic INTEGER NOT NULL,
			diastolic INTEGER NOT NULL,
			pulse INTEGER NOT NULL,
			measured_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS fitness_data (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			date DATE NOT NULL,
			data_type VARCHAR(50) NOT NULL,
			value FLOAT NOT NULL,
			unit VARCHAR(50) NOT NULL,
			source VARCHAR(50) NOT NULL,
			source_data_id VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			date_range_start DATE NOT NULL,
			date_range_end DATE NOT NULL,
			file_path VARCHAR(500) NOT NULL,
			generated_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL,
			operation_type VARCHAR(50) NOT NULL,
			resource_type VARCHAR(50) NOT NULL,
			resource_id UUID NOT NULL,
			timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
			ip_address VARCHAR(50),
			user_agent TEXT,
			additional_data JSONB
		)`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		require.NoError(t, err)
	}
}

// Property 21: Data Deletion Completeness
// When a user requests data deletion, ALL user data across ALL tables must be deleted
// Validates: Requirements 10.3
func TestProperty_DataDeletionCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("data deletion removes all user data from all tables", prop.ForAll(
		func(userID string) bool {
			ctx := context.Background()
			db, cleanup := setupTestDB(t)
			defer cleanup()

			auditLogger := audit.NewLogger(db, zap.NewNop())
			service := NewGDPRService(db, auditLogger, zap.NewNop())

			// Create test data across all tables
			createTestUserData(t, db, userID)

			// Verify data exists before deletion
			if !verifyUserDataExists(t, db, userID) {
				t.Logf("Failed to create test data for user %s", userID)
				return false
			}

			// Delete user data
			err := service.DeleteUserData(ctx, userID, "127.0.0.1", "test-agent")
			if err != nil {
				t.Logf("DeleteUserData failed: %v", err)
				return false
			}

			// Verify all data is deleted
			return verifyUserDataDeleted(t, db, userID)
		},
		gen.Const(uuid.New().String()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 22: Data Export Completeness
// When a user requests data export, ALL user data from ALL tables must be included in the export
// Validates: Requirements 10.4
func TestProperty_DataExportCompleteness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("data export includes all user data from all tables", prop.ForAll(
		func(userID string) bool {
			ctx := context.Background()
			db, cleanup := setupTestDB(t)
			defer cleanup()

			auditLogger := audit.NewLogger(db, zap.NewNop())
			service := NewGDPRService(db, auditLogger, zap.NewNop())

			// Create test data across all tables
			counts := createTestUserDataWithCounts(t, db, userID)

			// Export user data
			jsonData, err := service.ExportUserData(ctx, userID)
			if err != nil {
				t.Logf("ExportUserData failed: %v", err)
				return false
			}

			// Parse exported data
			var export UserDataExport
			if err := json.Unmarshal(jsonData, &export); err != nil {
				t.Logf("Failed to parse export JSON: %v", err)
				return false
			}

			// Verify all data is included
			if export.User == nil {
				t.Logf("User data missing from export")
				return false
			}

			if len(export.HealthCheckIns) != counts.HealthCheckIns {
				t.Logf("Health check-ins count mismatch: expected %d, got %d", counts.HealthCheckIns, len(export.HealthCheckIns))
				return false
			}

			if len(export.Medications) != counts.Medications {
				t.Logf("Medications count mismatch: expected %d, got %d", counts.Medications, len(export.Medications))
				return false
			}

			if len(export.MenstruationCycles) != counts.MenstruationCycles {
				t.Logf("Menstruation cycles count mismatch: expected %d, got %d", counts.MenstruationCycles, len(export.MenstruationCycles))
				return false
			}

			if len(export.BloodPressureReadings) != counts.BloodPressureReadings {
				t.Logf("Blood pressure readings count mismatch: expected %d, got %d", counts.BloodPressureReadings, len(export.BloodPressureReadings))
				return false
			}

			if len(export.FitnessData) != counts.FitnessData {
				t.Logf("Fitness data count mismatch: expected %d, got %d", counts.FitnessData, len(export.FitnessData))
				return false
			}

			if len(export.Reports) != counts.Reports {
				t.Logf("Reports count mismatch: expected %d, got %d", counts.Reports, len(export.Reports))
				return false
			}

			// Verify export timestamp is set
			if export.ExportedAt.IsZero() {
				t.Logf("ExportedAt timestamp not set")
				return false
			}

			return true
		},
		gen.Const(uuid.New().String()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Property 23: Audit Log Creation
// When any data modification occurs, an audit log entry must be created
// Validates: Requirements 10.5
func TestProperty_AuditLogCreation(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("audit log is created for all data modifications", prop.ForAll(
		func(userID string, operationType string, resourceType string, resourceID string) bool {
			ctx := context.Background()
			db, cleanup := setupTestDB(t)
			defer cleanup()

			auditLogger := audit.NewLogger(db, zap.NewNop())

			// Create audit log entry
			entry := audit.AuditLog{
				UserID:        userID,
				OperationType: audit.OperationType(operationType),
				ResourceType:  audit.ResourceType(resourceType),
				ResourceID:    resourceID,
				IPAddress:     "127.0.0.1",
				UserAgent:     "test-agent",
			}

			err := auditLogger.Log(ctx, entry)
			if err != nil {
				t.Logf("Failed to create audit log: %v", err)
				return false
			}

			// Verify audit log was created
			logs, err := auditLogger.GetAuditLogs(ctx, userID, 10)
			if err != nil {
				t.Logf("Failed to retrieve audit logs: %v", err)
				return false
			}

			if len(logs) == 0 {
				t.Logf("No audit logs found for user %s", userID)
				return false
			}

			// Verify the log entry matches
			found := false
			for _, log := range logs {
				if log.UserID == userID &&
					log.OperationType == audit.OperationType(operationType) &&
					log.ResourceType == audit.ResourceType(resourceType) &&
					log.ResourceID == resourceID {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Audit log entry not found with expected values")
				return false
			}

			return true
		},
		gen.Const(uuid.New().String()),
		gen.OneConstOf("CREATE", "UPDATE", "DELETE"),
		gen.OneConstOf("health_check_in", "medication", "menstruation_cycle", "blood_pressure_reading"),
		gen.Const(uuid.New().String()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper types and functions

type DataCounts struct {
	HealthCheckIns        int
	Medications           int
	MenstruationCycles    int
	BloodPressureReadings int
	FitnessData           int
	Reports               int
}

func createTestUserData(t *testing.T, db *pgxpool.Pool, userID string) {
	ctx := context.Background()

	// Create user
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, "Test User", "test@example.com", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create health check-in
	_, err = db.Exec(ctx, `
		INSERT INTO health_check_ins (id, user_id, check_in_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New().String(), userID, time.Now(), time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create health check-in: %v", err)
	}

	// Create medication
	_, err = db.Exec(ctx, `
		INSERT INTO medications (id, user_id, name, dosage, frequency, start_date, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, uuid.New().String(), userID, "Test Med", "10mg", "daily", time.Now(), true, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create medication: %v", err)
	}

	// Create menstruation cycle
	_, err = db.Exec(ctx, `
		INSERT INTO menstruation_cycles (id, user_id, start_date, flow_intensity, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New().String(), userID, time.Now(), "moderate", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create menstruation cycle: %v", err)
	}

	// Create blood pressure reading
	_, err = db.Exec(ctx, `
		INSERT INTO blood_pressure_readings (id, user_id, systolic, diastolic, pulse, measured_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New().String(), userID, 120, 80, 70, time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create blood pressure reading: %v", err)
	}

	// Create fitness data
	_, err = db.Exec(ctx, `
		INSERT INTO fitness_data (id, user_id, date, data_type, value, unit, source, source_data_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, uuid.New().String(), userID, time.Now(), "steps", 10000.0, "count", "test", uuid.New().String(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create fitness data: %v", err)
	}

	// Create report
	_, err = db.Exec(ctx, `
		INSERT INTO reports (id, user_id, date_range_start, date_range_end, file_path, generated_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, uuid.New().String(), userID, time.Now().AddDate(0, 0, -7), time.Now(), "/reports/test.pdf", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create report: %v", err)
	}

	// Create check-in session
	_, err = db.Exec(ctx, `
		INSERT INTO check_in_sessions (id, user_id, started_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, uuid.New().String(), userID, time.Now(), "completed", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create check-in session: %v", err)
	}
}

func createTestUserDataWithCounts(t *testing.T, db *pgxpool.Pool, userID string) DataCounts {
	ctx := context.Background()

	counts := DataCounts{
		HealthCheckIns:        2,
		Medications:           3,
		MenstruationCycles:    2,
		BloodPressureReadings: 4,
		FitnessData:           5,
		Reports:               1,
	}

	// Create user
	_, err := db.Exec(ctx, `
		INSERT INTO users (id, name, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, userID, "Test User", "test@example.com", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Create health check-ins
	for i := 0; i < counts.HealthCheckIns; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO health_check_ins (id, user_id, check_in_date, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
		`, uuid.New().String(), userID, time.Now().AddDate(0, 0, -i), time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to create health check-in: %v", err)
		}
	}

	// Create medications
	for i := 0; i < counts.Medications; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO medications (id, user_id, name, dosage, frequency, start_date, active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, uuid.New().String(), userID, "Test Med", "10mg", "daily", time.Now(), true, time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to create medication: %v", err)
		}
	}

	// Create menstruation cycles
	for i := 0; i < counts.MenstruationCycles; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO menstruation_cycles (id, user_id, start_date, flow_intensity, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, uuid.New().String(), userID, time.Now().AddDate(0, 0, -i*30), "moderate", time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to create menstruation cycle: %v", err)
		}
	}

	// Create blood pressure readings
	for i := 0; i < counts.BloodPressureReadings; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO blood_pressure_readings (id, user_id, systolic, diastolic, pulse, measured_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.New().String(), userID, 120, 80, 70, time.Now().AddDate(0, 0, -i), time.Now())
		if err != nil {
			t.Fatalf("Failed to create blood pressure reading: %v", err)
		}
	}

	// Create fitness data
	for i := 0; i < counts.FitnessData; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO fitness_data (id, user_id, date, data_type, value, unit, source, source_data_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`, uuid.New().String(), userID, time.Now().AddDate(0, 0, -i), "steps", 10000.0, "count", "test", uuid.New().String(), time.Now())
		if err != nil {
			t.Fatalf("Failed to create fitness data: %v", err)
		}
	}

	// Create reports
	for i := 0; i < counts.Reports; i++ {
		_, err = db.Exec(ctx, `
			INSERT INTO reports (id, user_id, date_range_start, date_range_end, file_path, generated_at, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, uuid.New().String(), userID, time.Now().AddDate(0, 0, -7), time.Now(), "/reports/test.pdf", time.Now(), time.Now())
		if err != nil {
			t.Fatalf("Failed to create report: %v", err)
		}
	}

	return counts
}

func verifyUserDataExists(t *testing.T, db *pgxpool.Pool, userID string) bool {
	ctx := context.Background()
	var count int

	// Check health check-ins
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM health_check_ins WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count == 0 {
		return false
	}

	// Check medications
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM medications WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count == 0 {
		return false
	}

	return true
}

func verifyUserDataDeleted(t *testing.T, db *pgxpool.Pool, userID string) bool {
	ctx := context.Background()
	var count int

	// Check health check-ins deleted
	err := db.QueryRow(ctx, "SELECT COUNT(*) FROM health_check_ins WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Health check-ins not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check medications deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM medications WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Medications not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check menstruation cycles deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM menstruation_cycles WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Menstruation cycles not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check blood pressure readings deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM blood_pressure_readings WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Blood pressure readings not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check fitness data deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM fitness_data WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Fitness data not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check reports deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM reports WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Reports not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check check-in sessions deleted
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM check_in_sessions WHERE user_id = $1", userID).Scan(&count)
	if err != nil || count != 0 {
		t.Logf("Check-in sessions not deleted: count=%d, err=%v", count, err)
		return false
	}

	// Check user is marked as deleted (soft delete)
	var deletedAt *time.Time
	err = db.QueryRow(ctx, "SELECT deleted_at FROM users WHERE id = $1", userID).Scan(&deletedAt)
	if err != nil || deletedAt == nil {
		t.Logf("User not marked as deleted: deletedAt=%v, err=%v", deletedAt, err)
		return false
	}

	return true
}
