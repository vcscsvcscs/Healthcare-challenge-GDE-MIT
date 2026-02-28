package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// HealthHandler implements health data API endpoints
type HealthHandler struct {
	service *service.HealthDataService
	logger  *zap.Logger
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(service *service.HealthDataService, logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		service: service,
		logger:  logger,
	}
}

// PostApiV1HealthMenstruation logs menstruation data
func (h *HealthHandler) PostApiV1HealthMenstruation(c *gin.Context) {
	var req api.MenstruationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userID := uuidToString(req.UserId)

	// Convert API request to model
	cycle := &model.MenstruationCycle{
		StartDate: dateToTime(req.StartDate),
		Symptoms:  []string{},
	}

	if req.EndDate != nil {
		endDate := dateToTime(*req.EndDate)
		cycle.EndDate = &endDate
	}

	if req.FlowIntensity != nil {
		intensity := string(*req.FlowIntensity)
		cycle.FlowIntensity = &intensity
	}

	if req.Symptoms != nil {
		cycle.Symptoms = *req.Symptoms
	}

	// Log menstruation data
	if err := h.service.LogMenstruation(c.Request.Context(), userID, cycle); err != nil {
		h.logger.Error("failed to log menstruation data",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to log menstruation data",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.MenstruationResponse{
		Id:        stringToUUID(cycle.ID),
		UserId:    stringToUUID(cycle.UserID),
		StartDate: timeToDate(cycle.StartDate),
		EndDate:   timePtrToDate(cycle.EndDate),
		Symptoms:  &cycle.Symptoms,
		CreatedAt: timePtr(cycle.CreatedAt),
	}

	if cycle.FlowIntensity != nil {
		intensity := api.MenstruationResponseFlowIntensity(*cycle.FlowIntensity)
		response.FlowIntensity = &intensity
	}

	h.logger.Info("menstruation data logged",
		zap.String("cycle_id", cycle.ID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, response)
}

// GetApiV1HealthMenstruation retrieves menstruation history
func (h *HealthHandler) GetApiV1HealthMenstruation(c *gin.Context, params api.GetApiV1HealthMenstruationParams) {
	userID := uuidToString(params.UserId)

	// Get menstruation history
	cycles, err := h.service.GetMenstruationHistory(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get menstruation history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get menstruation history",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	var response []api.MenstruationResponse
	for _, cycle := range cycles {
		menstruationResp := api.MenstruationResponse{
			Id:        stringToUUID(cycle.ID),
			UserId:    stringToUUID(cycle.UserID),
			StartDate: timeToDate(cycle.StartDate),
			EndDate:   timePtrToDate(cycle.EndDate),
			Symptoms:  &cycle.Symptoms,
			CreatedAt: timePtr(cycle.CreatedAt),
		}

		if cycle.FlowIntensity != nil {
			intensity := api.MenstruationResponseFlowIntensity(*cycle.FlowIntensity)
			menstruationResp.FlowIntensity = &intensity
		}

		response = append(response, menstruationResp)
	}

	h.logger.Info("menstruation history retrieved",
		zap.String("user_id", userID),
		zap.Int("count", len(response)),
	)

	c.JSON(http.StatusOK, response)
}

// PostApiV1HealthBloodPressure logs blood pressure reading
func (h *HealthHandler) PostApiV1HealthBloodPressure(c *gin.Context) {
	var req api.BloodPressureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userID := uuidToString(req.UserId)

	// Convert API request to model
	reading := &model.BloodPressureReading{
		Systolic:   req.Systolic,
		Diastolic:  req.Diastolic,
		Pulse:      req.Pulse,
		MeasuredAt: time.Now(),
	}

	if req.MeasuredAt != nil {
		reading.MeasuredAt = *req.MeasuredAt
	}

	// Log blood pressure
	if err := h.service.LogBloodPressure(c.Request.Context(), userID, reading); err != nil {
		h.logger.Error("failed to log blood pressure",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}

	// Convert to API response
	response := api.BloodPressureResponse{
		Id:         stringToUUID(reading.ID),
		UserId:     stringToUUID(reading.UserID),
		Systolic:   intPtr(reading.Systolic),
		Diastolic:  intPtr(reading.Diastolic),
		Pulse:      intPtr(reading.Pulse),
		MeasuredAt: timePtr(reading.MeasuredAt),
		CreatedAt:  timePtr(reading.CreatedAt),
	}

	h.logger.Info("blood pressure logged",
		zap.String("reading_id", reading.ID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, response)
}

// GetApiV1HealthBloodPressure retrieves blood pressure history
func (h *HealthHandler) GetApiV1HealthBloodPressure(c *gin.Context, params api.GetApiV1HealthBloodPressureParams) {
	userID := uuidToString(params.UserId)

	// Get blood pressure history
	readings, err := h.service.GetBloodPressureHistory(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to get blood pressure history",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get blood pressure history",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	var response []api.BloodPressureResponse
	for _, reading := range readings {
		response = append(response, api.BloodPressureResponse{
			Id:         stringToUUID(reading.ID),
			UserId:     stringToUUID(reading.UserID),
			Systolic:   intPtr(reading.Systolic),
			Diastolic:  intPtr(reading.Diastolic),
			Pulse:      intPtr(reading.Pulse),
			MeasuredAt: timePtr(reading.MeasuredAt),
			CreatedAt:  timePtr(reading.CreatedAt),
		})
	}

	h.logger.Info("blood pressure history retrieved",
		zap.String("user_id", userID),
		zap.Int("count", len(response)),
	)

	c.JSON(http.StatusOK, response)
}

// PostApiV1HealthFitnessSync syncs fitness data from Health Connect
func (h *HealthHandler) PostApiV1HealthFitnessSync(c *gin.Context) {
	var req api.FitnessSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userID := uuidToString(req.UserId)

	// Convert API request to model
	var fitnessData []model.FitnessDataPoint
	for _, data := range req.DataPoints {
		fitnessData = append(fitnessData, model.FitnessDataPoint{
			Date:         dateToTime(data.Date),
			DataType:     string(data.DataType),
			Value:        data.Value,
			Unit:         string(data.Unit),
			Source:       string(data.Source),
			SourceDataID: data.SourceDataId,
		})
	}

	// Sync fitness data
	if err := h.service.SyncFitnessData(c.Request.Context(), userID, fitnessData); err != nil {
		h.logger.Error("failed to sync fitness data",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to sync fitness data",
			Details: stringPtr(err.Error()),
		})
		return
	}

	h.logger.Info("fitness data synced",
		zap.String("user_id", userID),
		zap.Int("count", len(fitnessData)),
	)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Fitness data synced successfully",
		"synced_count": len(fitnessData),
	})
}
