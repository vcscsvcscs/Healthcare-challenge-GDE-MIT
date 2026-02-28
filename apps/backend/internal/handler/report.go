package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/oapi-codegen/runtime/types"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// ReportHandler implements report API endpoints
type ReportHandler struct {
	service *service.ReportService
	logger  *zap.Logger
}

// NewReportHandler creates a new ReportHandler
func NewReportHandler(service *service.ReportService, logger *zap.Logger) *ReportHandler {
	return &ReportHandler{
		service: service,
		logger:  logger,
	}
}

// PostApiV1ReportsGenerate generates a health report
func (h *ReportHandler) PostApiV1ReportsGenerate(c *gin.Context) {
	var req api.GenerateReportRequest
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

	// Convert dates
	startDate := dateToTime(req.StartDate)
	endDate := dateToTime(req.EndDate)

	// Validate date range
	if startDate.After(endDate) {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Start date must be before or equal to end date",
		})
		return
	}

	// Generate report (this could be done asynchronously in production)
	// For now, we'll use a placeholder user name
	userName := "User"
	reportID, err := h.service.GenerateReport(c.Request.Context(), userID, userName, startDate, endDate)
	if err != nil {
		h.logger.Error("failed to generate report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to generate report",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Return report ID
	response := gin.H{
		"report_id": reportID,
		"message":   "Report generated successfully",
	}

	h.logger.Info("report generated",
		zap.String("report_id", reportID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, response)
}

// GetApiV1ReportsId downloads a report
func (h *ReportHandler) GetApiV1ReportsId(c *gin.Context, id types.UUID) {
	reportID := uuidToString(id)

	h.logger.Info("downloading report",
		zap.String("report_id", reportID),
	)

	// Get report PDF
	pdfBytes, err := h.service.GetReport(c.Request.Context(), reportID)
	if err != nil {
		h.logger.Error("failed to get report",
			zap.Error(err),
			zap.String("report_id", reportID),
		)
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Report not found",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Return PDF
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=health_report_%s.pdf", reportID))
	c.Header("Content-Length", fmt.Sprintf("%d", len(pdfBytes)))
	c.Data(http.StatusOK, "application/pdf", pdfBytes)

	h.logger.Info("report downloaded",
		zap.String("report_id", reportID),
		zap.Int("size_bytes", len(pdfBytes)),
	)
}
