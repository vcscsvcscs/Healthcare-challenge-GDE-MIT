package service

import (
	"context"
	"fmt"

	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"go.uber.org/zap"
)

// DashboardRepositoryInterface defines the interface for dashboard data access
type DashboardRepositoryInterface interface {
	GetAggregatedMetrics(ctx context.Context, userID string, days int) (*repository.AggregatedMetrics, error)
	GetDailyMetrics(ctx context.Context, userID string, days int) ([]repository.DailyMetrics, error)
}

// DashboardService manages dashboard data aggregation and trends
type DashboardService struct {
	repo   DashboardRepositoryInterface
	logger *zap.Logger
}

// NewDashboardService creates a new DashboardService
func NewDashboardService(repo DashboardRepositoryInterface, logger *zap.Logger) *DashboardService {
	return &DashboardService{
		repo:   repo,
		logger: logger,
	}
}

// DashboardSummary represents aggregated dashboard data
type DashboardSummary struct {
	Period           string                    `json:"period"`
	AveragePain      float64                   `json:"average_pain"`
	MoodDistribution map[string]int            `json:"mood_distribution"`
	EnergyLevels     map[string]int            `json:"energy_levels"`
	CheckInCount     int                       `json:"check_in_count"`
	TimeSeriesData   []repository.DailyMetrics `json:"time_series_data"`
}

// TrendAnalysis represents trend analysis data
type TrendAnalysis struct {
	Period           string                    `json:"period"`
	AveragePain      float64                   `json:"average_pain"`
	MoodDistribution map[string]int            `json:"mood_distribution"`
	EnergyLevels     map[string]int            `json:"energy_levels"`
	TimeSeriesData   []repository.DailyMetrics `json:"time_series_data"`
}

// GetSummary retrieves dashboard summary with time range filtering
func (s *DashboardService) GetSummary(ctx context.Context, userID string, days int) (*DashboardSummary, error) {
	s.logger.Info("getting dashboard summary",
		zap.String("user_id", userID),
		zap.Int("days", days),
	)

	// Validate days parameter
	if days != 7 && days != 30 && days != 90 {
		s.logger.Warn("invalid days parameter, defaulting to 7",
			zap.Int("days", days),
		)
		days = 7
	}

	// Get aggregated metrics
	metrics, err := s.repo.GetAggregatedMetrics(ctx, userID, days)
	if err != nil {
		s.logger.Error("failed to get aggregated metrics",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}

	// Get time-series data
	dailyMetrics, err := s.repo.GetDailyMetrics(ctx, userID, days)
	if err != nil {
		s.logger.Error("failed to get daily metrics",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	// Handle empty datasets gracefully
	if metrics.CheckInCount == 0 {
		s.logger.Info("no check-ins found for user in time period",
			zap.String("user_id", userID),
			zap.Int("days", days),
		)
		return &DashboardSummary{
			Period:           fmt.Sprintf("%d days", days),
			AveragePain:      0,
			MoodDistribution: make(map[string]int),
			EnergyLevels:     make(map[string]int),
			CheckInCount:     0,
			TimeSeriesData:   []repository.DailyMetrics{},
		}, nil
	}

	summary := &DashboardSummary{
		Period:           fmt.Sprintf("%d days", days),
		AveragePain:      metrics.AveragePainLevel,
		MoodDistribution: metrics.MoodDistribution,
		EnergyLevels:     metrics.EnergyLevels,
		CheckInCount:     metrics.CheckInCount,
		TimeSeriesData:   dailyMetrics,
	}

	s.logger.Info("dashboard summary retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("check_in_count", summary.CheckInCount),
	)

	return summary, nil
}

// GetTrends retrieves trend analysis with aggregations
func (s *DashboardService) GetTrends(ctx context.Context, userID string, days int) (*TrendAnalysis, error) {
	s.logger.Info("getting trend analysis",
		zap.String("user_id", userID),
		zap.Int("days", days),
	)

	// Validate days parameter
	if days != 7 && days != 30 && days != 90 {
		s.logger.Warn("invalid days parameter, defaulting to 7",
			zap.Int("days", days),
		)
		days = 7
	}

	// Get aggregated metrics
	metrics, err := s.repo.GetAggregatedMetrics(ctx, userID, days)
	if err != nil {
		s.logger.Error("failed to get aggregated metrics for trends",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get aggregated metrics: %w", err)
	}

	// Get time-series data for trends
	dailyMetrics, err := s.repo.GetDailyMetrics(ctx, userID, days)
	if err != nil {
		s.logger.Error("failed to get daily metrics for trends",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get daily metrics: %w", err)
	}

	// Handle empty datasets gracefully
	if len(dailyMetrics) == 0 {
		s.logger.Info("no data found for trend analysis",
			zap.String("user_id", userID),
			zap.Int("days", days),
		)
		return &TrendAnalysis{
			Period:           fmt.Sprintf("%d days", days),
			AveragePain:      0,
			MoodDistribution: make(map[string]int),
			EnergyLevels:     make(map[string]int),
			TimeSeriesData:   []repository.DailyMetrics{},
		}, nil
	}

	trends := &TrendAnalysis{
		Period:           fmt.Sprintf("%d days", days),
		AveragePain:      metrics.AveragePainLevel,
		MoodDistribution: metrics.MoodDistribution,
		EnergyLevels:     metrics.EnergyLevels,
		TimeSeriesData:   dailyMetrics,
	}

	s.logger.Info("trend analysis retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("data_points", len(dailyMetrics)),
	)

	return trends, nil
}
