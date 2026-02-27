package repository

import (
	"context"
	"fmt"
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
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
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

// runMigrations runs the database migrations
func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create tables
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
		`CREATE TABLE IF NOT EXISTS conversation_messages (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			session_id UUID NOT NULL REFERENCES check_in_sessions(id) ON DELETE CASCADE,
			role VARCHAR(50) NOT NULL,
			content TEXT NOT NULL,
			audio_file_path VARCHAR(500),
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS health_check_ins (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			session_id UUID REFERENCES check_in_sessions(id) ON DELETE SET NULL,
			check_in_date DATE NOT NULL,
			symptoms TEXT[],
			mood VARCHAR(50),
			pain_level INTEGER CHECK (pain_level >= 0 AND pain_level <= 10),
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
		`CREATE TABLE IF NOT EXISTS medication_logs (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			medication_id UUID NOT NULL REFERENCES medications(id) ON DELETE CASCADE,
			taken_at TIMESTAMP NOT NULL,
			adherence BOOLEAN NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
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
			systolic INTEGER NOT NULL CHECK (systolic >= 70 AND systolic <= 250),
			diastolic INTEGER NOT NULL CHECK (diastolic >= 40 AND diastolic <= 150),
			pulse INTEGER NOT NULL CHECK (pulse >= 30 AND pulse <= 220),
			measured_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		require.NoError(t, err)
	}
}

// createTestUser creates a test user and returns the user ID
func createTestUser(t *testing.T, pool *pgxpool.Pool) string {
	ctx := context.Background()
	userID := uuid.New().String()

	_, err := pool.Exec(ctx,
		`INSERT INTO users (id, name, email) VALUES ($1, $2, $3)`,
		userID, "Test User", fmt.Sprintf("test-%s@example.com", userID))
	require.NoError(t, err)

	return userID
}

// **Validates: Requirements 1.1**
// Feature: eva-health-backend, Property 1: Session Creation Generates Unique IDs
func TestProperty_SessionCreationGeneratesUniqueIDs(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	repo := NewCheckInRepository(pool, logger)

	userID := createTestUser(t, pool)

	properties := gopter.NewProperties(nil)

	properties.Property("session IDs are unique across multiple creations", prop.ForAll(
		func(n int) bool {
			ctx := context.Background()
			sessionIDs := make(map[string]bool)

			// Create n sessions
			for i := 0; i < n; i++ {
				session := &model.Session{
					ID:        uuid.New().String(),
					UserID:    userID,
					StartedAt: time.Now(),
					Status:    model.SessionStatusActive,
				}

				err := repo.CreateSession(ctx, session)
				if err != nil {
					t.Logf("Failed to create session: %v", err)
					return false
				}

				// Check if ID is unique
				if sessionIDs[session.ID] {
					t.Logf("Duplicate session ID found: %s", session.ID)
					return false
				}
				sessionIDs[session.ID] = true
			}

			return len(sessionIDs) == n
		},
		gen.IntRange(1, 20), // Test with 1 to 20 sessions
	))

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// **Validates: Requirements 4.2, 4.3, 4.4, 5.2, 6.5**
// Feature: eva-health-backend, Property 10: Medication CRUD Preserves ID
func TestProperty_MedicationCRUDPreservesID(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	repo := NewMedicationRepository(pool, logger)

	userID := createTestUser(t, pool)

	properties := gopter.NewProperties(nil)

	properties.Property("medication ID is preserved after update", prop.ForAll(
		func(name, dosage, frequency, notes string) bool {
			ctx := context.Background()

			// Create medication
			originalID := uuid.New().String()
			medication := &model.Medication{
				ID:        originalID,
				UserID:    userID,
				Name:      name,
				Dosage:    dosage,
				Frequency: frequency,
				StartDate: time.Now(),
				Notes:     &notes,
				Active:    true,
			}

			err := repo.Create(ctx, medication)
			if err != nil {
				t.Logf("Failed to create medication: %v", err)
				return false
			}

			// Update medication
			newDosage := dosage + " (updated)"
			medication.Dosage = newDosage

			err = repo.Update(ctx, medication)
			if err != nil {
				t.Logf("Failed to update medication: %v", err)
				return false
			}

			// Retrieve medication
			retrieved, err := repo.FindByID(ctx, originalID)
			if err != nil {
				t.Logf("Failed to retrieve medication: %v", err)
				return false
			}

			// Verify ID is preserved and dosage is updated
			return retrieved.ID == originalID && retrieved.Dosage == newDosage
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) < 200 }),
	))

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// **Validates: Requirements 4.4**
// Feature: eva-health-backend, Property 11: Medication Deletion Removes Record
func TestProperty_MedicationDeletionRemovesRecord(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	repo := NewMedicationRepository(pool, logger)

	userID := createTestUser(t, pool)

	properties := gopter.NewProperties(nil)

	properties.Property("deleted medication does not appear in user's medication list", prop.ForAll(
		func(name, dosage, frequency string) bool {
			ctx := context.Background()

			// Create medication
			medicationID := uuid.New().String()
			medication := &model.Medication{
				ID:        medicationID,
				UserID:    userID,
				Name:      name,
				Dosage:    dosage,
				Frequency: frequency,
				StartDate: time.Now(),
				Active:    true,
			}

			err := repo.Create(ctx, medication)
			if err != nil {
				t.Logf("Failed to create medication: %v", err)
				return false
			}

			// Verify medication exists
			medications, err := repo.FindByUserID(ctx, userID)
			if err != nil {
				t.Logf("Failed to find medications: %v", err)
				return false
			}

			found := false
			for _, med := range medications {
				if med.ID == medicationID {
					found = true
					break
				}
			}

			if !found {
				t.Logf("Medication not found before deletion")
				return false
			}

			// Delete medication
			err = repo.Delete(ctx, medicationID)
			if err != nil {
				t.Logf("Failed to delete medication: %v", err)
				return false
			}

			// Verify medication is removed
			medications, err = repo.FindByUserID(ctx, userID)
			if err != nil {
				t.Logf("Failed to find medications after deletion: %v", err)
				return false
			}

			for _, med := range medications {
				if med.ID == medicationID {
					t.Logf("Medication still found after deletion")
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}

// **Validates: Requirements 4.2, 5.2, 6.5**
// Feature: eva-health-backend, Property 13: List Sorting Consistency
func TestProperty_ListSortingConsistency(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	logger, _ := zap.NewDevelopment()
	medRepo := NewMedicationRepository(pool, logger)

	userID := createTestUser(t, pool)

	properties := gopter.NewProperties(nil)

	properties.Property("medications are sorted by start date in descending order", prop.ForAll(
		func(count int) bool {
			ctx := context.Background()

			// Create multiple medications with different start dates
			startDates := make([]time.Time, count)
			for i := 0; i < count; i++ {
				startDate := time.Now().AddDate(0, 0, -i) // Each medication starts one day earlier
				startDates[i] = startDate

				medication := &model.Medication{
					ID:        uuid.New().String(),
					UserID:    userID,
					Name:      fmt.Sprintf("Medication %d", i),
					Dosage:    "100mg",
					Frequency: "daily",
					StartDate: startDate,
					Active:    true,
				}

				err := medRepo.Create(ctx, medication)
				if err != nil {
					t.Logf("Failed to create medication: %v", err)
					return false
				}
			}

			// Retrieve medications
			medications, err := medRepo.FindByUserID(ctx, userID)
			if err != nil {
				t.Logf("Failed to find medications: %v", err)
				return false
			}

			// Verify sorting (descending order by start date)
			for i := 0; i < len(medications)-1; i++ {
				if medications[i].StartDate.Before(medications[i+1].StartDate) {
					t.Logf("Medications not sorted correctly: %v should be after %v",
						medications[i].StartDate, medications[i+1].StartDate)
					return false
				}
			}

			// Clean up for next iteration
			for _, med := range medications {
				_ = medRepo.Delete(ctx, med.ID)
			}

			return true
		},
		gen.IntRange(2, 10), // Test with 2 to 10 medications
	))

	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	properties.TestingRun(t, params)
}
