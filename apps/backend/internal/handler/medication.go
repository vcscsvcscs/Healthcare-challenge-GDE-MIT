package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// MedicationHandler implements medication API endpoints
type MedicationHandler struct {
	service *service.MedicationService
	logger  *zap.Logger
}

// NewMedicationHandler creates a new MedicationHandler
func NewMedicationHandler(service *service.MedicationService, logger *zap.Logger) *MedicationHandler {
	return &MedicationHandler{
		service: service,
		logger:  logger,
	}
}

// PostApiV1HealthMedications adds a new medication
func (h *MedicationHandler) PostApiV1HealthMedications(c *gin.Context) {
	var req api.CreateMedicationRequest
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
	medication := &model.Medication{
		Name:      req.Name,
		Dosage:    req.Dosage,
		Frequency: req.Frequency,
		StartDate: dateToTime(req.StartDate),
		EndDate:   nil,
		Notes:     req.Notes,
	}

	if req.EndDate != nil {
		endDate := dateToTime(*req.EndDate)
		medication.EndDate = &endDate
	}

	// Add medication
	if err := h.service.AddMedication(c.Request.Context(), userID, medication); err != nil {
		h.logger.Error("failed to add medication",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to add medication",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.MedicationResponse{
		Id:        stringToUUID(medication.ID),
		UserId:    stringToUUID(medication.UserID),
		Name:      stringPtr(medication.Name),
		Dosage:    stringPtr(medication.Dosage),
		Frequency: stringPtr(medication.Frequency),
		StartDate: timeToDate(medication.StartDate),
		EndDate:   timePtrToDate(medication.EndDate),
		Notes:     medication.Notes,
		Active:    boolPtr(medication.Active),
		CreatedAt: timePtr(medication.CreatedAt),
	}

	h.logger.Info("medication added",
		zap.String("medication_id", medication.ID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, response)
}

// GetApiV1HealthMedications lists all medications for a user
func (h *MedicationHandler) GetApiV1HealthMedications(c *gin.Context, params api.GetApiV1HealthMedicationsParams) {
	userID := uuidToString(params.UserId)

	// Get medications
	medications, err := h.service.ListMedications(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to list medications",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list medications",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	var response []api.MedicationResponse
	for _, med := range medications {
		response = append(response, api.MedicationResponse{
			Id:        stringToUUID(med.ID),
			UserId:    stringToUUID(med.UserID),
			Name:      stringPtr(med.Name),
			Dosage:    stringPtr(med.Dosage),
			Frequency: stringPtr(med.Frequency),
			StartDate: timeToDate(med.StartDate),
			EndDate:   timePtrToDate(med.EndDate),
			Notes:     med.Notes,
			Active:    boolPtr(med.Active),
			CreatedAt: timePtr(med.CreatedAt),
		})
	}

	h.logger.Info("medications listed",
		zap.String("user_id", userID),
		zap.Int("count", len(response)),
	)

	c.JSON(http.StatusOK, response)
}

// PutApiV1HealthMedicationsId updates a medication
func (h *MedicationHandler) PutApiV1HealthMedicationsId(c *gin.Context, id types.UUID) {
	var req api.UpdateMedicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	medicationID := uuidToString(id)

	// Convert API request to model
	medication := &model.Medication{
		Name:      derefString(req.Name),
		Dosage:    derefString(req.Dosage),
		Frequency: derefString(req.Frequency),
		Notes:     req.Notes,
	}

	if req.EndDate != nil {
		endDate := dateToTime(*req.EndDate)
		medication.EndDate = &endDate
	}

	// Update medication
	if err := h.service.UpdateMedication(c.Request.Context(), medicationID, medication); err != nil {
		h.logger.Error("failed to update medication",
			zap.Error(err),
			zap.String("medication_id", medicationID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update medication",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.MedicationResponse{
		Id:        stringToUUID(medication.ID),
		UserId:    stringToUUID(medication.UserID),
		Name:      stringPtr(medication.Name),
		Dosage:    stringPtr(medication.Dosage),
		Frequency: stringPtr(medication.Frequency),
		StartDate: timeToDate(medication.StartDate),
		EndDate:   timePtrToDate(medication.EndDate),
		Notes:     medication.Notes,
		Active:    boolPtr(medication.Active),
		CreatedAt: timePtr(medication.CreatedAt),
	}

	h.logger.Info("medication updated",
		zap.String("medication_id", medicationID),
	)

	c.JSON(http.StatusOK, response)
}

// DeleteApiV1HealthMedicationsId deletes a medication
func (h *MedicationHandler) DeleteApiV1HealthMedicationsId(c *gin.Context, id types.UUID) {
	medicationID := uuidToString(id)

	// Delete medication
	if err := h.service.DeleteMedication(c.Request.Context(), medicationID); err != nil {
		h.logger.Error("failed to delete medication",
			zap.Error(err),
			zap.String("medication_id", medicationID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete medication",
			Details: stringPtr(err.Error()),
		})
		return
	}

	h.logger.Info("medication deleted",
		zap.String("medication_id", medicationID),
	)

	c.Status(http.StatusNoContent)
}

// derefString safely dereferences a string pointer, returning empty string if nil
func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
