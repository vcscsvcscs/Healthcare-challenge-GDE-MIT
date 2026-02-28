package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// DashboardRepository manages dashboard data aggregations
type DashboardRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

// NewDashboardRepository creates a new DashboardRepository
func NewDashboardRepository(db *pgxpool.Pool, logger *zap.Logger) *DashboardRepository {
	return &DashboardRepository{
		db:     db,
		logger: logger,
	}
}

// AggregatedMetrics represents aggregated health metrics
type AggregatedMetrics struct {
	AveragePainLevel float64
	MoodDistribution map[string]int
	EnergyLevels     map[string]int
	CheckInCount     int
}

// DailyMetrics represents health metrics for a single day
type DailyMetrics struct {
	Date            time.Time
	PainLevel       *int
	Mood            *string
	EnergyLevel     *string
	SleepQuality    *string
	MedicationTaken *string
	SymptomCount    int
	ActivityCount   int
}

// GetHealthCheckIns retrieves health check-ins for a user within a date range
func (r *DashboardRepository) GetHealthCheckIns(ctx context.Context, userID string, startDate, endDate time.Time) ([]model.HealthCheckIn, error) {
	query := `
		SELECT 
			id, user_id, session_id, check_in_date,
			symptoms, mood, pain_level, energy_level, sleep_quality,
			medication_taken, physical_activity,
			breakfast, lunch, dinner,
			general_feeling, additional_notes, raw_transcript,
			created_at, updated_at
		FROM health_check_ins
		WHERE user_id = $1 AND check_in_date >= $2 AND check_in_date <= $3
		ORDER BY check_in_date DESC
	`

	rows, err := r.db.Query(ctx, query, userID, startDate, endDate)
	if err != nil {
		r.logger.Error("failed to get health check-ins for dashboard",
			zap.Error(err),
			zap.String("user_id", userID),
		)
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

// GetAggregatedMetrics computes aggregated metrics for a user over a time period
func (r *DashboardRepository) GetAggregatedMetrics(ctx context.Context, userID string, days int) (*AggregatedMetrics, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			AVG(CASE WHEN pain_level IS NOT NULL THEN pain_level ELSE 0 END) as avg_pain,
			COUNT(*) as check_in_count,
			mood,
			energy_level
		FROM health_check_ins
		WHERE user_id = $1 AND check_in_date >= $2
		GROUP BY mood, energy_level
	`

	rows, err := r.db.Query(ctx, query, userID, startDate)
	if err != nil {
		r.logger.Error("failed to get aggregated metrics",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}
	defer rows.Close()

	metrics := &AggregatedMetrics{
		MoodDistribution: make(map[string]int),
		EnergyLevels:     make(map[string]int),
	}

	var totalPain float64
	var painCount int

	for rows.Next() {
		var avgPain float64
		var count int
		var mood, energyLevel *string

		err := rows.Scan(&avgPain, &count, &mood, &energyLevel)
		if err != nil {
			r.logger.Error("failed to scan aggregated metrics", zap.Error(err))
			continue
		}

		if avgPain > 0 {
			totalPain += avgPain * float64(count)
			painCount += count
		}

		metrics.CheckInCount += count

		if mood != nil && *mood != "" {
			metrics.MoodDistribution[*mood] += count
		}

		if energyLevel != nil && *energyLevel != "" {
			metrics.EnergyLevels[*energyLevel] += count
		}
	}

	if painCount > 0 {
		metrics.AveragePainLevel = totalPain / float64(painCount)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating aggregated metrics", zap.Error(err))
		return nil, fmt.Errorf("error iterating aggregated metrics: %w", err)
	}

	return metrics, nil
}

// GetDailyMetrics retrieves daily metrics for time-series data
func (r *DashboardRepository) GetDailyMetrics(ctx context.Context, userID string, days int) ([]DailyMetrics, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	query := `
		SELECT 
			check_in_date,
			pain_level,
			mood,
			energy_level,
			sleep_quality,
			medication_taken,
			COALESCE(array_length(symptoms, 1), 0) as symptom_count,
			COALESCE(array_length(physical_activity, 1), 0) as activity_count
		FROM health_check_ins
		WHERE user_id = $1 AND check_in_date >= $2
		ORDER BY check_in_date ASC
	`

	rows, err := r.db.Query(ctx, query, userID, startDate)
	if err != nil {
		r.logger.Error("failed to get daily metrics",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}
	defer rows.Close()

	var dailyMetrics []DailyMetrics
	for rows.Next() {
		var dm DailyMetrics
		err := rows.Scan(
			&dm.Date,
			&dm.PainLevel,
			&dm.Mood,
			&dm.EnergyLevel,
			&dm.SleepQuality,
			&dm.MedicationTaken,
			&dm.SymptomCount,
			&dm.ActivityCount,
		)
		if err != nil {
			r.logger.Error("failed to scan daily metrics", zap.Error(err))
			continue
		}
		dailyMetrics = append(dailyMetrics, dm)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating daily metrics", zap.Error(err))
		return nil, fmt.Errorf("error iterating daily metrics: %w", err)
	}

	return dailyMetrics, nil
}

// SaveReport saves a report record
func (r *DashboardRepository) SaveReport(ctx context.Context, report *model.Report) error {
	query := `
		INSERT INTO reports (
			id, user_id, start_date, end_date,
			file_path, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	status := "completed" // Default status for generated reports

	_, err := r.db.Exec(ctx, query,
		report.ID,
		report.UserID,
		report.DateRangeStart,
		report.DateRangeEnd,
		report.FilePath,
		status,
	)

	if err != nil {
		r.logger.Error("failed to save report",
			zap.Error(err),
			zap.String("report_id", report.ID),
			zap.String("user_id", report.UserID),
		)
		return fmt.Errorf("failed to save report: %w", err)
	}

	return nil
}

// GetReportByID retrieves a report by ID
func (r *DashboardRepository) GetReportByID(ctx context.Context, reportID string) (*model.Report, error) {
	query := `
		SELECT 
			id, user_id, start_date, end_date,
			file_path, created_at
		FROM reports
		WHERE id = $1
	`

	var report model.Report
	err := r.db.QueryRow(ctx, query, reportID).Scan(
		&report.ID,
		&report.UserID,
		&report.DateRangeStart,
		&report.DateRangeEnd,
		&report.FilePath,
		&report.CreatedAt,
	)

	if err != nil {
		r.logger.Error("failed to get report", zap.Error(err), zap.String("report_id", reportID))
		return nil, fmt.Errorf("failed to get report: %w", err)
	}

	// Set GeneratedAt to CreatedAt for compatibility
	report.GeneratedAt = report.CreatedAt

	return &report, nil
}

// GetReportsByUserID retrieves all reports for a user
func (r *DashboardRepository) GetReportsByUserID(ctx context.Context, userID string) ([]model.Report, error) {
	query := `
		SELECT 
			id, user_id, start_date, end_date,
			file_path, created_at
		FROM reports
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		r.logger.Error("failed to get reports", zap.Error(err), zap.String("user_id", userID))
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}
	defer rows.Close()

	var reports []model.Report
	for rows.Next() {
		var report model.Report
		err := rows.Scan(
			&report.ID,
			&report.UserID,
			&report.DateRangeStart,
			&report.DateRangeEnd,
			&report.FilePath,
			&report.CreatedAt,
		)
		if err != nil {
			r.logger.Error("failed to scan report", zap.Error(err))
			continue
		}
		// Set GeneratedAt to CreatedAt for compatibility
		report.GeneratedAt = report.CreatedAt
		reports = append(reports, report)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("error iterating reports", zap.Error(err))
		return nil, fmt.Errorf("error iterating reports: %w", err)
	}

	return reports, nil
}
