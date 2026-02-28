package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// GDPRHandler implements GDPR compliance endpoints
type GDPRHandler struct {
	service *service.GDPRService
	logger  *zap.Logger
}

// NewGDPRHandler creates a new GDPRHandler
func NewGDPRHandler(service *service.GDPRService, logger *zap.Logger) *GDPRHandler {
	return &GDPRHandler{
		service: service,
		logger:  logger,
	}
}

// DeleteUserData handles user data deletion requests (GDPR right to be forgotten)
// DELETE /api/v1/users/:userId/data
func (h *GDPRHandler) DeleteUserData(c *gin.Context) {
	userIDParam := c.Param("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid user ID",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userIDStr := userID.String()
	ipAddress := c.ClientIP()
	userAgent := c.Request.UserAgent()

	h.logger.Info("processing user data deletion request (GDPR)",
		zap.String("user_id", userIDStr),
		zap.String("ip", ipAddress),
	)

	// Delete user data
	if err := h.service.DeleteUserData(c.Request.Context(), userIDStr, ipAddress, userAgent); err != nil {
		h.logger.Error("failed to delete user data",
			zap.Error(err),
			zap.String("user_id", userIDStr),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete user data",
			Details: stringPtr(err.Error()),
		})
		return
	}

	h.logger.Info("user data deleted successfully (GDPR)",
		zap.String("user_id", userIDStr),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "User data deleted successfully",
		"user_id": userIDStr,
	})
}

// ExportUserData handles user data export requests (GDPR right to data portability)
// GET /api/v1/users/:userId/export
func (h *GDPRHandler) ExportUserData(c *gin.Context) {
	userIDParam := c.Param("userId")
	userID, err := uuid.Parse(userIDParam)
	if err != nil {
		h.logger.Error("invalid user ID", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid user ID",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userIDStr := userID.String()

	h.logger.Info("processing user data export request (GDPR)",
		zap.String("user_id", userIDStr),
	)

	// Export user data
	jsonData, err := h.service.ExportUserData(c.Request.Context(), userIDStr)
	if err != nil {
		h.logger.Error("failed to export user data",
			zap.Error(err),
			zap.String("user_id", userIDStr),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to export user data",
			Details: stringPtr(err.Error()),
		})
		return
	}

	h.logger.Info("user data exported successfully (GDPR)",
		zap.String("user_id", userIDStr),
		zap.Int("data_size_bytes", len(jsonData)),
	)

	// Return JSON file as download
	filename := fmt.Sprintf("user_data_%s.json", userIDStr)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Data(http.StatusOK, "application/json", jsonData)
}
