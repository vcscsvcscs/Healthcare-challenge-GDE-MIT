package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// DashboardHandler implements dashboard API endpoints
type DashboardHandler struct {
	service *service.DashboardService
	logger  *zap.Logger
}

// NewDashboardHandler creates a new DashboardHandler
func NewDashboardHandler(service *service.DashboardService, logger *zap.Logger) *DashboardHandler {
	return &DashboardHandler{
		service: service,
		logger:  logger,
	}
}

// GetApiV1DashboardSummary retrieves dashboard summary
func (h *DashboardHandler) GetApiV1DashboardSummary(c *gin.Context, params api.GetApiV1DashboardSummaryParams) {
	userID := uuidToString(params.UserId)

	// Default to 7 days if not specified
	days := 7
	if params.Days != nil {
		days = int(*params.Days)
	}

	// Get dashboard summary
	summary, err := h.service.GetSummary(c.Request.Context(), userID, days)
	if err != nil {
		h.logger.Error("failed to get dashboard summary",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.Int("days", days),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get dashboard summary",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.DashboardSummary{
		Period:       stringPtr(summary.Period),
		AveragePain:  &summary.AveragePain,
		CheckInCount: intPtr(summary.CheckInCount),
	}

	// Convert mood distribution
	if summary.MoodDistribution != nil {
		response.MoodDistribution = &struct {
			Negative *int `json:"negative,omitempty"`
			Neutral  *int `json:"neutral,omitempty"`
			Positive *int `json:"positive,omitempty"`
		}{
			Positive: intPtrFromMap(summary.MoodDistribution, "positive"),
			Neutral:  intPtrFromMap(summary.MoodDistribution, "neutral"),
			Negative: intPtrFromMap(summary.MoodDistribution, "negative"),
		}
	}

	// Convert energy levels
	if summary.EnergyLevels != nil {
		response.EnergyLevels = &struct {
			High   *int `json:"high,omitempty"`
			Low    *int `json:"low,omitempty"`
			Medium *int `json:"medium,omitempty"`
		}{
			High:   intPtrFromMap(summary.EnergyLevels, "high"),
			Medium: intPtrFromMap(summary.EnergyLevels, "medium"),
			Low:    intPtrFromMap(summary.EnergyLevels, "low"),
		}
	}

	// Convert time series data
	if summary.TimeSeriesData != nil {
		var timeSeriesData []api.DailyMetrics
		for _, daily := range summary.TimeSeriesData {
			timeSeriesData = append(timeSeriesData, api.DailyMetrics{
				Date:         timeToDate(daily.Date),
				PainLevel:    daily.PainLevel,
				Mood:         daily.Mood,
				EnergyLevel:  daily.EnergyLevel,
				SleepQuality: daily.SleepQuality,
			})
		}
		response.TimeSeriesData = &timeSeriesData
	}

	h.logger.Info("dashboard summary retrieved",
		zap.String("user_id", userID),
		zap.Int("days", days),
		zap.Int("check_in_count", summary.CheckInCount),
	)

	c.JSON(http.StatusOK, response)
}

// intPtrFromMap safely gets an int pointer from a map
func intPtrFromMap(m map[string]int, key string) *int {
	if val, ok := m[key]; ok {
		return intPtr(val)
	}
	return nil
}
