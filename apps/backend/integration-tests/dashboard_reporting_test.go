package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/handler"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/pdf"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// TestDashboardAndReportingIntegration tests the complete dashboard and reporting flow
// Requirements: 7.1-7.5 (Dashboard), 8.1-8.6 (Reports)
func TestDashboardAndReportingIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	ctx := context.Background()
	logger := zap.NewNop()

	// Initialize database connection
	db, cleanup := setupTestDatabase(t, ctx)
	defer cleanup()

	// Initialize repositories
	healthRepo := repository.NewHealthDataRepository(db, logger)
	dashboardRepo := repository.NewDashboardRepository(db, logger)
	medicationRepo := repository.NewMedicationRepository(db, logger)

	// Initialize services
	healthService := service.NewHealthDataService(healthRepo, logger)
	dashboardService := service.NewDashboardService(dashboardRepo, logger)
	// Initialize PDF generator and mock blob storage for report service
	pdfGen := pdf.NewPDFGenerator(logger)
	mockBlobStorage := NewMockBlobStorageClient(logger)
	reportService := service.NewReportService(dashboardRepo, healthRepo, medicationRepo, mockBlobStorage, pdfGen, logger)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(healthService, logger)
	dashboardHandler := handler.NewDashboardHandler(dashboardService, logger)
	reportHandler := handler.NewReportHandler(reportService, logger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerDashboardAndReportRoutes(router, healthHandler, dashboardHandler, reportHandler)

	// Test user ID
	userID := uuid.New()

	t.Run("Dashboard summary and trends", func(t *testing.T) {
		// Clean up any existing data for this user
		cleanupAllTestData(t, ctx, db, userID.String())

		// Step 1: Seed test data with health check-ins
		t.Log("Step 1: Seeding health check-in data")
		seedHealthCheckInData(t, ctx, db, userID.String())

		// Step 2: Get dashboard summary (default 7 days)
		t.Log("Step 2: Getting dashboard summary for 7 days")
		summary7Days := getDashboardSummary(t, router, userID, nil)
		require.NotNil(t, summary7Days, "Dashboard summary should not be nil")

		// Verify summary contains expected fields
		assert.NotNil(t, summary7Days.Period, "Period should be set")
		assert.NotNil(t, summary7Days.CheckInCount, "Check-in count should be set")
		assert.NotNil(t, summary7Days.AveragePain, "Average pain should be set")

		// Verify check-in count is correct (should have data from last 7 days)
		assert.Greater(t, *summary7Days.CheckInCount, 0, "Should have check-ins in last 7 days")

		// Step 3: Get dashboard summary for 30 days
		t.Log("Step 3: Getting dashboard summary for 30 days")
		days30 := api.GetApiV1DashboardSummaryParamsDays(30)
		summary30Days := getDashboardSummary(t, router, userID, &days30)
		require.NotNil(t, summary30Days, "Dashboard summary should not be nil")

		// 30-day summary should have more or equal check-ins than 7-day
		assert.GreaterOrEqual(t, *summary30Days.CheckInCount, *summary7Days.CheckInCount,
			"30-day summary should have >= check-ins than 7-day")

		// Step 4: Verify aggregations
		t.Log("Step 4: Verifying dashboard aggregations")
		verifyDashboardAggregations(t, summary7Days)

		// Step 5: Verify time series data
		t.Log("Step 5: Verifying time series data")
		if summary7Days.TimeSeriesData != nil {
			assert.Greater(t, len(*summary7Days.TimeSeriesData), 0, "Should have time series data")
			// Verify data is grouped by date
			for _, daily := range *summary7Days.TimeSeriesData {
				assert.NotNil(t, daily.Date, "Each daily metric should have a date")
			}
		}

		// Cleanup
		cleanupAllTestData(t, ctx, db, userID.String())
	})

	t.Run("Report generation and download", func(t *testing.T) {
		// Clean up any existing data for this user
		cleanupAllTestData(t, ctx, db, userID.String())

		// Step 1: Seed comprehensive test data
		t.Log("Step 1: Seeding comprehensive health data")
		seedComprehensiveHealthData(t, ctx, db, router, userID)

		// Step 2: Generate report
		t.Log("Step 2: Generating health report")
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()
		reportID := generateReport(t, router, userID, startDate, endDate)
		require.NotEmpty(t, reportID, "Report ID should not be empty")

		// Step 3: Download report PDF
		t.Log("Step 3: Downloading report PDF")
		pdfBytes := downloadReport(t, router, reportID)
		require.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")

		// Verify PDF header (PDF files start with %PDF-)
		assert.True(t, len(pdfBytes) > 4, "PDF should have at least header bytes")
		assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")

		// Step 4: Verify report content size
		t.Log("Step 4: Verifying report content")
		// A comprehensive report should be at least a few KB
		assert.Greater(t, len(pdfBytes), 1000, "Report should contain substantial content")

		// Cleanup
		cleanupAllTestData(t, ctx, db, userID.String())
	})

	t.Run("Empty data scenarios", func(t *testing.T) {
		// Use a new user with no data
		emptyUserID := uuid.New()

		// Step 1: Get dashboard summary with no data
		t.Log("Step 1: Getting dashboard summary with no data")
		summary := getDashboardSummary(t, router, emptyUserID, nil)
		require.NotNil(t, summary, "Dashboard summary should not be nil even with no data")

		// Should return empty dataset with appropriate metadata
		assert.NotNil(t, summary.CheckInCount, "Check-in count should be set")
		assert.Equal(t, 0, *summary.CheckInCount, "Check-in count should be 0 for empty data")

		// Step 2: Generate report with no data
		t.Log("Step 2: Generating report with no data")
		startDate := time.Now().AddDate(0, 0, -30)
		endDate := time.Now()
		reportID := generateReport(t, router, emptyUserID, startDate, endDate)
		require.NotEmpty(t, reportID, "Report ID should not be empty even with no data")

		// Step 3: Download empty report
		t.Log("Step 3: Downloading empty report")
		pdfBytes := downloadReport(t, router, reportID)
		require.NotEmpty(t, pdfBytes, "PDF bytes should not be empty")

		// Should still be a valid PDF
		assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")
	})
}

// getDashboardSummary retrieves dashboard summary for a user
func getDashboardSummary(t *testing.T, router *gin.Engine, userID uuid.UUID, days *api.GetApiV1DashboardSummaryParamsDays) *api.DashboardSummary {
	url := "/api/v1/dashboard/summary?user_id=" + userID.String()
	if days != nil {
		url += fmt.Sprintf("&days=%d", *days)
	}

	req := httptest.NewRequest(http.MethodGet, url, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code, "Get dashboard summary should return 200 OK")

	var response api.DashboardSummary
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")

	return &response
}

// generateReport generates a health report and returns its ID
func generateReport(t *testing.T, router *gin.Engine, userID uuid.UUID, startDate, endDate time.Time) string {
	reqBody := api.GenerateReportRequest{
		UserId:    userID,
		StartDate: types.Date{Time: startDate},
		EndDate:   types.Date{Time: endDate},
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/reports/generate", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code, "Generate report should return 200 OK")

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")

	reportID, ok := response["report_id"].(string)
	require.True(t, ok, "Response should contain report_id")
	require.NotEmpty(t, reportID, "Report ID should not be empty")

	return reportID
}

// downloadReport downloads a report PDF
func downloadReport(t *testing.T, router *gin.Engine, reportID string) []byte {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/reports/"+reportID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code, "Download report should return 200 OK")

	// Verify content type
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"), "Content-Type should be application/pdf")

	return w.Body.Bytes()
}

// verifyDashboardAggregations verifies that dashboard aggregations are correct
func verifyDashboardAggregations(t *testing.T, summary *api.DashboardSummary) {
	// Verify average pain is within valid range (0-10)
	if summary.AveragePain != nil {
		assert.GreaterOrEqual(t, *summary.AveragePain, 0.0, "Average pain should be >= 0")
		assert.LessOrEqual(t, *summary.AveragePain, 10.0, "Average pain should be <= 10")
	}

	// Verify mood distribution
	if summary.MoodDistribution != nil {
		totalMood := 0
		if summary.MoodDistribution.Positive != nil {
			totalMood += *summary.MoodDistribution.Positive
			assert.GreaterOrEqual(t, *summary.MoodDistribution.Positive, 0, "Positive mood count should be >= 0")
		}
		if summary.MoodDistribution.Neutral != nil {
			totalMood += *summary.MoodDistribution.Neutral
			assert.GreaterOrEqual(t, *summary.MoodDistribution.Neutral, 0, "Neutral mood count should be >= 0")
		}
		if summary.MoodDistribution.Negative != nil {
			totalMood += *summary.MoodDistribution.Negative
			assert.GreaterOrEqual(t, *summary.MoodDistribution.Negative, 0, "Negative mood count should be >= 0")
		}
		// Total mood entries should match or be less than check-in count
		if summary.CheckInCount != nil {
			assert.LessOrEqual(t, totalMood, *summary.CheckInCount, "Total mood entries should be <= check-in count")
		}
	}

	// Verify energy levels
	if summary.EnergyLevels != nil {
		totalEnergy := 0
		if summary.EnergyLevels.High != nil {
			totalEnergy += *summary.EnergyLevels.High
			assert.GreaterOrEqual(t, *summary.EnergyLevels.High, 0, "High energy count should be >= 0")
		}
		if summary.EnergyLevels.Medium != nil {
			totalEnergy += *summary.EnergyLevels.Medium
			assert.GreaterOrEqual(t, *summary.EnergyLevels.Medium, 0, "Medium energy count should be >= 0")
		}
		if summary.EnergyLevels.Low != nil {
			totalEnergy += *summary.EnergyLevels.Low
			assert.GreaterOrEqual(t, *summary.EnergyLevels.Low, 0, "Low energy count should be >= 0")
		}
		// Total energy entries should match or be less than check-in count
		if summary.CheckInCount != nil {
			assert.LessOrEqual(t, totalEnergy, *summary.CheckInCount, "Total energy entries should be <= check-in count")
		}
	}
}

// seedHealthCheckInData seeds health check-in data for testing
func seedHealthCheckInData(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID string) {
	// Insert health check-ins for the last 10 days
	for i := 0; i < 10; i++ {
		checkInDate := time.Now().AddDate(0, 0, -i)

		// Vary the data to test aggregations
		painLevel := (i % 5) + 1 // Pain levels 1-5
		mood := []string{"positive", "neutral", "negative"}[i%3]
		energyLevel := []string{"high", "medium", "low"}[i%3]
		sleepQuality := []string{"excellent", "good", "fair", "poor"}[i%4]

		query := `
			INSERT INTO health_check_ins (
				id, user_id, check_in_date, pain_level, mood, energy_level, sleep_quality,
				symptoms, physical_activity, general_feeling, created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`

		_, err := db.Exec(ctx, query,
			uuid.New().String(),
			userID,
			checkInDate,
			painLevel,
			mood,
			energyLevel,
			sleepQuality,
			[]string{"headache"},
			[]string{"walking"},
			"Feeling okay",
			time.Now(),
			time.Now(),
		)
		require.NoError(t, err, "Should be able to insert health check-in")
	}

	t.Logf("Seeded 10 health check-ins for user %s", userID)
}

// seedComprehensiveHealthData seeds comprehensive health data for report testing
func seedComprehensiveHealthData(t *testing.T, ctx context.Context, db *pgxpool.Pool, router *gin.Engine, userID uuid.UUID) {
	// Seed health check-ins
	seedHealthCheckInData(t, ctx, db, userID.String())

	// Seed blood pressure readings
	for i := 0; i < 5; i++ {
		measuredAt := time.Now().AddDate(0, 0, -i*2)
		logBloodPressureWithDate(t, router, userID, measuredAt, 120+i, 80+i, 70+i)
	}

	// Seed menstruation cycles (if applicable)
	logMenstruationCycleWithDate(t, router, userID, time.Now().AddDate(0, 0, -7), "moderate", []string{"cramps"})

	t.Logf("Seeded comprehensive health data for user %s", userID.String())
}

// registerDashboardAndReportRoutes registers dashboard and report routes on the router
func registerDashboardAndReportRoutes(router *gin.Engine, healthHandler *handler.HealthHandler, dashboardHandler *handler.DashboardHandler, reportHandler *handler.ReportHandler) {
	v1 := router.Group("/api/v1")
	{
		// Health routes (for seeding data)
		health := v1.Group("/health")
		{
			health.POST("/blood-pressure", healthHandler.PostApiV1HealthBloodPressure)
			health.POST("/menstruation", healthHandler.PostApiV1HealthMenstruation)
		}

		// Dashboard routes
		dashboard := v1.Group("/dashboard")
		{
			dashboard.GET("/summary", func(c *gin.Context) {
				userIDStr := c.Query("user_id")
				userID, err := uuid.Parse(userIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
					return
				}

				var days *api.GetApiV1DashboardSummaryParamsDays
				if daysStr := c.Query("days"); daysStr != "" {
					var d int
					if _, err := fmt.Sscanf(daysStr, "%d", &d); err == nil {
						daysParam := api.GetApiV1DashboardSummaryParamsDays(d)
						days = &daysParam
					}
				}

				dashboardHandler.GetApiV1DashboardSummary(c, api.GetApiV1DashboardSummaryParams{
					UserId: userID,
					Days:   days,
				})
			})
		}

		// Report routes
		reports := v1.Group("/reports")
		{
			reports.POST("/generate", reportHandler.PostApiV1ReportsGenerate)
			reports.GET("/:id", func(c *gin.Context) {
				idStr := c.Param("id")
				id, err := uuid.Parse(idStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
					return
				}
				reportHandler.GetApiV1ReportsId(c, types.UUID(id))
			})
		}
	}
}

// cleanupAllTestData removes all test data for a user
func cleanupAllTestData(t *testing.T, ctx context.Context, db *pgxpool.Pool, userID string) {
	// Clean up health check-ins
	_, err := db.Exec(ctx, "DELETE FROM health_check_ins WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to cleanup health check-ins: %v", err)
	}

	// Clean up blood pressure readings
	_, err = db.Exec(ctx, "DELETE FROM blood_pressure_readings WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to cleanup blood pressure readings: %v", err)
	}

	// Clean up menstruation cycles
	_, err = db.Exec(ctx, "DELETE FROM menstruation_cycles WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to cleanup menstruation cycles: %v", err)
	}

	// Clean up reports
	_, err = db.Exec(ctx, "DELETE FROM reports WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Warning: failed to cleanup reports: %v", err)
	}
}
