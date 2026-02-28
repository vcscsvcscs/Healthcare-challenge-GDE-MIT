package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// HealthDataService handles health data management business logic
type HealthDataService struct {
	repo   *repository.HealthDataRepository
	logger *zap.Logger
}

// NewHealthDataService creates a new HealthDataService
func NewHealthDataService(repo *repository.HealthDataRepository, logger *zap.Logger) *HealthDataService {
	return &HealthDataService{
		repo:   repo,
		logger: logger,
	}
}

// LogMenstruation logs menstruation cycle data
func (s *HealthDataService) LogMenstruation(ctx context.Context, userID string, data *model.MenstruationCycle) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate flow intensity if provided
	if data.FlowIntensity != nil {
		validIntensities := map[string]bool{
			"light":    true,
			"moderate": true,
			"heavy":    true,
		}
		if !validIntensities[*data.FlowIntensity] {
			return fmt.Errorf("invalid flow intensity: must be light, moderate, or heavy")
		}
	}

	// Generate ID if not provided
	if data.ID == "" {
		data.ID = uuid.New().String()
	}

	// Set user ID
	data.UserID = userID

	// Set timestamps
	now := time.Now()
	data.CreatedAt = now
	data.UpdatedAt = now

	if err := s.repo.SaveMenstruation(ctx, data); err != nil {
		s.logger.Error("failed to log menstruation data",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return fmt.Errorf("failed to log menstruation data: %w", err)
	}

	s.logger.Info("menstruation data logged successfully",
		zap.String("cycle_id", data.ID),
		zap.String("user_id", userID),
	)

	return nil
}

// GetMenstruationHistory retrieves menstruation cycle history for a user
func (s *HealthDataService) GetMenstruationHistory(ctx context.Context, userID string) ([]model.MenstruationCycle, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	cycles, err := s.repo.GetMenstruationByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get menstruation history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get menstruation history: %w", err)
	}

	s.logger.Info("menstruation history retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(cycles)),
	)

	return cycles, nil
}

// LogBloodPressure logs a blood pressure reading
func (s *HealthDataService) LogBloodPressure(ctx context.Context, userID string, reading *model.BloodPressureReading) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate blood pressure ranges
	if reading.Systolic < 70 || reading.Systolic > 250 {
		return fmt.Errorf("invalid systolic value: must be between 70 and 250")
	}
	if reading.Diastolic < 40 || reading.Diastolic > 150 {
		return fmt.Errorf("invalid diastolic value: must be between 40 and 150")
	}
	if reading.Pulse < 30 || reading.Pulse > 220 {
		return fmt.Errorf("invalid pulse value: must be between 30 and 220")
	}

	// Generate ID if not provided
	if reading.ID == "" {
		reading.ID = uuid.New().String()
	}

	// Set user ID
	reading.UserID = userID

	// Set timestamp
	reading.CreatedAt = time.Now()

	if err := s.repo.SaveBloodPressure(ctx, reading); err != nil {
		s.logger.Error("failed to log blood pressure reading",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return fmt.Errorf("failed to log blood pressure reading: %w", err)
	}

	s.logger.Info("blood pressure reading logged successfully",
		zap.String("reading_id", reading.ID),
		zap.String("user_id", userID),
		zap.Int("systolic", reading.Systolic),
		zap.Int("diastolic", reading.Diastolic),
	)

	return nil
}

// GetBloodPressureHistory retrieves blood pressure reading history for a user
func (s *HealthDataService) GetBloodPressureHistory(ctx context.Context, userID string) ([]model.BloodPressureReading, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	readings, err := s.repo.GetBloodPressureByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get blood pressure history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get blood pressure history: %w", err)
	}

	s.logger.Info("blood pressure history retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(readings)),
	)

	return readings, nil
}

// SyncFitnessData syncs fitness data from Health Connect with deduplication
func (s *HealthDataService) SyncFitnessData(ctx context.Context, userID string, fitnessData []model.FitnessDataPoint) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	syncedCount := 0
	skippedCount := 0

	for _, dataPoint := range fitnessData {
		// Validate data type
		validDataTypes := map[string]bool{
			"steps":          true,
			"heart_rate":     true,
			"sleep":          true,
			"calories":       true,
			"distance":       true,
			"active_minutes": true,
		}
		if !validDataTypes[dataPoint.DataType] {
			s.logger.Warn("invalid fitness data type",
				zap.String("data_type", dataPoint.DataType),
			)
			continue
		}

		// Check if data point already exists (deduplication by source_data_id)
		if dataPoint.SourceDataID != "" {
			exists, err := s.repo.FitnessDataExists(ctx, dataPoint.SourceDataID)
			if err != nil {
				s.logger.Error("failed to check fitness data existence",
					zap.Error(err),
					zap.String("source_data_id", dataPoint.SourceDataID),
				)
				return fmt.Errorf("failed to check fitness data existence: %w", err)
			}

			if exists {
				s.logger.Debug("fitness data already synced, skipping",
					zap.String("source_data_id", dataPoint.SourceDataID),
				)
				skippedCount++
				continue
			}
		}

		// Generate ID if not provided
		if dataPoint.ID == "" {
			dataPoint.ID = uuid.New().String()
		}

		// Set user ID
		dataPoint.UserID = userID

		// Set timestamp
		dataPoint.CreatedAt = time.Now()

		// Save new data point
		if err := s.repo.SaveFitnessData(ctx, &dataPoint); err != nil {
			s.logger.Error("failed to save fitness data",
				zap.Error(err),
				zap.String("user_id", userID),
				zap.String("data_type", dataPoint.DataType),
			)
			return fmt.Errorf("failed to save fitness data: %w", err)
		}

		syncedCount++
	}

	s.logger.Info("fitness data synced successfully",
		zap.String("user_id", userID),
		zap.Int("synced_count", syncedCount),
		zap.Int("skipped_count", skippedCount),
		zap.Int("total_count", len(fitnessData)),
	)

	return nil
}

// GetFitnessHistory retrieves fitness data history for a user within a date range
func (s *HealthDataService) GetFitnessHistory(ctx context.Context, userID string, startDate, endDate time.Time) ([]model.FitnessDataPoint, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	// Validate date range
	if startDate.After(endDate) {
		return nil, fmt.Errorf("start date must be before or equal to end date")
	}

	dataPoints, err := s.repo.GetFitnessDataByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get fitness history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get fitness history: %w", err)
	}

	s.logger.Info("fitness history retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(dataPoints)),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	return dataPoints, nil
}
