package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"go.uber.org/zap"
)

// MockDashboardRepository is a mock implementation of DashboardRepository
type MockDashboardRepository struct {
	mock.Mock
}

func (m *MockDashboardRepository) GetAggregatedMetrics(ctx context.Context, userID string, days int) (*repository.AggregatedMetrics, error) {
	args := m.Called(ctx, userID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AggregatedMetrics), args.Error(1)
}

func (m *MockDashboardRepository) GetDailyMetrics(ctx context.Context, userID string, days int) ([]repository.DailyMetrics, error) {
	args := m.Called(ctx, userID, days)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.DailyMetrics), args.Error(1)
}

func TestDashboardService_GetSummary_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDashboardRepository)
	logger := zap.NewNop()
	service := NewDashboardService(mockRepo, logger)

	ctx := context.Background()
	userID := "test-user-id"
	days := 7

	expectedMetrics := &repository.AggregatedMetrics{
		AveragePainLevel: 3.5,
		MoodDistribution: map[string]int{"positive": 5, "neutral": 2},
		EnergyLevels:     map[string]int{"high": 4, "medium": 3},
		CheckInCount:     7,
	}

	painLevel := 3
	mood := "positive"
	energyLevel := "high"
	expectedDailyMetrics := []repository.DailyMetrics{
		{
			Date:        time.Now().AddDate(0, 0, -1),
			PainLevel:   &painLevel,
			Mood:        &mood,
			EnergyLevel: &energyLevel,
		},
	}

	mockRepo.On("GetAggregatedMetrics", ctx, userID, days).Return(expectedMetrics, nil)
	mockRepo.On("GetDailyMetrics", ctx, userID, days).Return(expectedDailyMetrics, nil)

	// Act
	summary, err := service.GetSummary(ctx, userID, days)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "7 days", summary.Period)
	assert.Equal(t, 3.5, summary.AveragePain)
	assert.Equal(t, 7, summary.CheckInCount)
	assert.Equal(t, 5, summary.MoodDistribution["positive"])
	assert.Equal(t, 4, summary.EnergyLevels["high"])
	assert.Len(t, summary.TimeSeriesData, 1)

	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_EmptyDataset(t *testing.T) {
	// Arrange
	mockRepo := new(MockDashboardRepository)
	logger := zap.NewNop()
	service := NewDashboardService(mockRepo, logger)

	ctx := context.Background()
	userID := "test-user-id"
	days := 7

	emptyMetrics := &repository.AggregatedMetrics{
		AveragePainLevel: 0,
		MoodDistribution: make(map[string]int),
		EnergyLevels:     make(map[string]int),
		CheckInCount:     0,
	}

	emptyDailyMetrics := []repository.DailyMetrics{}

	mockRepo.On("GetAggregatedMetrics", ctx, userID, days).Return(emptyMetrics, nil)
	mockRepo.On("GetDailyMetrics", ctx, userID, days).Return(emptyDailyMetrics, nil)

	// Act
	summary, err := service.GetSummary(ctx, userID, days)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "7 days", summary.Period)
	assert.Equal(t, 0.0, summary.AveragePain)
	assert.Equal(t, 0, summary.CheckInCount)
	assert.Empty(t, summary.MoodDistribution)
	assert.Empty(t, summary.EnergyLevels)
	assert.Empty(t, summary.TimeSeriesData)

	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetSummary_InvalidDays(t *testing.T) {
	// Arrange
	mockRepo := new(MockDashboardRepository)
	logger := zap.NewNop()
	service := NewDashboardService(mockRepo, logger)

	ctx := context.Background()
	userID := "test-user-id"
	invalidDays := 15 // Not 7, 30, or 90

	emptyMetrics := &repository.AggregatedMetrics{
		AveragePainLevel: 0,
		MoodDistribution: make(map[string]int),
		EnergyLevels:     make(map[string]int),
		CheckInCount:     0,
	}

	emptyDailyMetrics := []repository.DailyMetrics{}

	// Should default to 7 days
	mockRepo.On("GetAggregatedMetrics", ctx, userID, 7).Return(emptyMetrics, nil)
	mockRepo.On("GetDailyMetrics", ctx, userID, 7).Return(emptyDailyMetrics, nil)

	// Act
	summary, err := service.GetSummary(ctx, userID, invalidDays)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, summary)
	assert.Equal(t, "7 days", summary.Period) // Should default to 7

	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetTrends_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockDashboardRepository)
	logger := zap.NewNop()
	service := NewDashboardService(mockRepo, logger)

	ctx := context.Background()
	userID := "test-user-id"
	days := 30

	expectedMetrics := &repository.AggregatedMetrics{
		AveragePainLevel: 4.2,
		MoodDistribution: map[string]int{"positive": 15, "neutral": 10, "negative": 5},
		EnergyLevels:     map[string]int{"high": 12, "medium": 13, "low": 5},
		CheckInCount:     30,
	}

	painLevel := 4
	mood := "positive"
	energyLevel := "medium"
	expectedDailyMetrics := []repository.DailyMetrics{
		{
			Date:        time.Now().AddDate(0, 0, -1),
			PainLevel:   &painLevel,
			Mood:        &mood,
			EnergyLevel: &energyLevel,
		},
		{
			Date:        time.Now().AddDate(0, 0, -2),
			PainLevel:   &painLevel,
			Mood:        &mood,
			EnergyLevel: &energyLevel,
		},
	}

	mockRepo.On("GetAggregatedMetrics", ctx, userID, days).Return(expectedMetrics, nil)
	mockRepo.On("GetDailyMetrics", ctx, userID, days).Return(expectedDailyMetrics, nil)

	// Act
	trends, err := service.GetTrends(ctx, userID, days)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "30 days", trends.Period)
	assert.Equal(t, 4.2, trends.AveragePain)
	assert.Equal(t, 15, trends.MoodDistribution["positive"])
	assert.Equal(t, 12, trends.EnergyLevels["high"])
	assert.Len(t, trends.TimeSeriesData, 2)

	mockRepo.AssertExpectations(t)
}

func TestDashboardService_GetTrends_EmptyDataset(t *testing.T) {
	// Arrange
	mockRepo := new(MockDashboardRepository)
	logger := zap.NewNop()
	service := NewDashboardService(mockRepo, logger)

	ctx := context.Background()
	userID := "test-user-id"
	days := 90

	emptyMetrics := &repository.AggregatedMetrics{
		AveragePainLevel: 0,
		MoodDistribution: make(map[string]int),
		EnergyLevels:     make(map[string]int),
		CheckInCount:     0,
	}

	emptyDailyMetrics := []repository.DailyMetrics{}

	mockRepo.On("GetAggregatedMetrics", ctx, userID, days).Return(emptyMetrics, nil)
	mockRepo.On("GetDailyMetrics", ctx, userID, days).Return(emptyDailyMetrics, nil)

	// Act
	trends, err := service.GetTrends(ctx, userID, days)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, trends)
	assert.Equal(t, "90 days", trends.Period)
	assert.Equal(t, 0.0, trends.AveragePain)
	assert.Empty(t, trends.MoodDistribution)
	assert.Empty(t, trends.EnergyLevels)
	assert.Empty(t, trends.TimeSeriesData)

	mockRepo.AssertExpectations(t)
}
