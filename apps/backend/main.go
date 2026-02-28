package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/audit"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/config"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/handler"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/middleware"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/pdf"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

var (
	logger *zap.Logger
	pool   *pgxpool.Pool
	cfg    *config.Config
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize Zap logger
	if cfg.Server.Environment == "production" {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Configuration loaded successfully",
		zap.String("environment", cfg.Server.Environment),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database connection pool with pgx
	pool, err = pgxpool.New(context.Background(), cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer pool.Close()

	// Test database connection
	if err := pool.Ping(context.Background()); err != nil {
		logger.Fatal("Failed to ping database", zap.Error(err))
	}
	logger.Info("Successfully connected to database")

	// Initialize Azure clients
	openAIClient, err := azure.NewOpenAIClient(
		cfg.Azure.OpenAI.Endpoint,
		cfg.Azure.OpenAI.APIKey,
		cfg.Azure.OpenAI.Deployment,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize Azure OpenAI client", zap.Error(err))
	}

	speechClient, err := azure.NewSpeechServiceClient(
		cfg.Azure.Speech.SubscriptionKey,
		cfg.Azure.Speech.Region,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize Azure Speech Service client", zap.Error(err))
	}

	blobClient, err := azure.NewBlobStorageClient(
		cfg.Azure.Storage.AccountName,
		cfg.Azure.Storage.AccountKey,
		cfg.Azure.Storage.AudioContainer,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize Azure Blob Storage client", zap.Error(err))
	}

	// Initialize repositories
	checkInRepo := repository.NewCheckInRepository(pool, logger)
	medicationRepo := repository.NewMedicationRepository(pool, logger)
	healthDataRepo := repository.NewHealthDataRepository(pool, logger)
	dashboardRepo := repository.NewDashboardRepository(pool, logger)

	// Initialize services
	checkInService := service.NewCheckInService(
		checkInRepo,
		openAIClient,
		speechClient,
		blobClient,
		logger,
	)
	medicationService := service.NewMedicationService(medicationRepo, logger)
	healthDataService := service.NewHealthDataService(healthDataRepo, logger)
	dashboardService := service.NewDashboardService(dashboardRepo, logger)

	// Initialize PDF generator
	pdfGenerator := pdf.NewPDFGenerator(logger)

	// Initialize report service with separate blob client for reports
	reportBlobClient, err := azure.NewBlobStorageClient(
		cfg.Azure.Storage.AccountName,
		cfg.Azure.Storage.AccountKey,
		cfg.Azure.Storage.ReportContainer,
		logger,
	)
	if err != nil {
		logger.Fatal("Failed to initialize report blob storage client", zap.Error(err))
	}

	reportService := service.NewReportService(
		dashboardRepo,
		healthDataRepo,
		medicationRepo,
		reportBlobClient,
		pdfGenerator,
		logger,
	)

	// Initialize GDPR service
	auditLogger := audit.NewLogger(pool, logger)
	gdprService := service.NewGDPRService(
		pool,
		auditLogger,
		logger,
	)

	// Initialize handlers
	checkInHandler := handler.NewCheckInHandler(checkInService, logger)
	medicationHandler := handler.NewMedicationHandler(medicationService, logger)
	healthHandler := handler.NewHealthHandler(healthDataService, logger)
	dashboardHandler := handler.NewDashboardHandler(dashboardService, logger)
	reportHandler := handler.NewReportHandler(reportService, logger)
	gdprHandler := handler.NewGDPRHandler(gdprService, logger)

	// Create a unified handler that implements the ServerInterface
	apiHandler := &APIHandler{
		checkIn:    checkInHandler,
		medication: medicationHandler,
		health:     healthHandler,
		dashboard:  dashboardHandler,
		report:     reportHandler,
		gdpr:       gdprHandler,
		pool:       pool,
		logger:     logger,
	}

	// Set Gin mode
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	r := gin.New()

	// Add recovery middleware (must be first)
	r.Use(middleware.RecoveryMiddleware(logger))

	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Configure appropriately for production
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-ID", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Add request ID middleware
	r.Use(middleware.RequestIDMiddleware())

	// Add tracing middleware
	r.Use(middleware.TracingMiddleware())

	// Add request logging middleware
	r.Use(middleware.RequestLoggingMiddleware(logger))

	// Add error logging middleware
	r.Use(middleware.ErrorLoggingMiddleware(logger))

	// Add slow query logging middleware
	r.Use(middleware.SlowQueryLoggingMiddleware(logger, 1*time.Second))

	// Register generated API handlers
	api.RegisterHandlers(r, apiHandler)

	// Start server with graceful shutdown
	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Close database connections
	pool.Close()

	logger.Info("Server exited")
}

// APIHandler implements the generated ServerInterface by delegating to individual handlers
type APIHandler struct {
	checkIn    *handler.CheckInHandler
	medication *handler.MedicationHandler
	health     *handler.HealthHandler
	dashboard  *handler.DashboardHandler
	report     *handler.ReportHandler
	gdpr       *handler.GDPRHandler
	pool       *pgxpool.Pool
	logger     *zap.Logger
}

// Check-in endpoints
func (h *APIHandler) PostApiV1CheckinStart(c *gin.Context) {
	h.checkIn.PostApiV1CheckinStart(c)
}

func (h *APIHandler) PostApiV1CheckinAudioStream(c *gin.Context, params api.PostApiV1CheckinAudioStreamParams) {
	h.checkIn.PostApiV1CheckinAudioStream(c, params)
}

func (h *APIHandler) PostApiV1CheckinRespond(c *gin.Context) {
	h.checkIn.PostApiV1CheckinRespond(c)
}

func (h *APIHandler) GetApiV1CheckinStatusSessionId(c *gin.Context, sessionId openapi_types.UUID) {
	h.checkIn.GetApiV1CheckinStatusSessionId(c, sessionId)
}

func (h *APIHandler) GetApiV1CheckinQuestionAudioSessionIdQuestionId(c *gin.Context, sessionId openapi_types.UUID, questionId string) {
	h.checkIn.GetApiV1CheckinQuestionAudioSessionIdQuestionId(c, sessionId, questionId)
}

func (h *APIHandler) PostApiV1CheckinComplete(c *gin.Context) {
	h.checkIn.PostApiV1CheckinComplete(c)
}

// Dashboard endpoints
func (h *APIHandler) GetApiV1DashboardSummary(c *gin.Context, params api.GetApiV1DashboardSummaryParams) {
	h.dashboard.GetApiV1DashboardSummary(c, params)
}

// Health data endpoints
func (h *APIHandler) GetApiV1HealthBloodPressure(c *gin.Context, params api.GetApiV1HealthBloodPressureParams) {
	h.health.GetApiV1HealthBloodPressure(c, params)
}

func (h *APIHandler) PostApiV1HealthBloodPressure(c *gin.Context) {
	h.health.PostApiV1HealthBloodPressure(c)
}

func (h *APIHandler) PostApiV1HealthFitnessSync(c *gin.Context) {
	h.health.PostApiV1HealthFitnessSync(c)
}

func (h *APIHandler) GetApiV1HealthMedications(c *gin.Context, params api.GetApiV1HealthMedicationsParams) {
	h.medication.GetApiV1HealthMedications(c, params)
}

func (h *APIHandler) PostApiV1HealthMedications(c *gin.Context) {
	h.medication.PostApiV1HealthMedications(c)
}

func (h *APIHandler) DeleteApiV1HealthMedicationsId(c *gin.Context, id openapi_types.UUID) {
	h.medication.DeleteApiV1HealthMedicationsId(c, id)
}

func (h *APIHandler) PutApiV1HealthMedicationsId(c *gin.Context, id openapi_types.UUID) {
	h.medication.PutApiV1HealthMedicationsId(c, id)
}

func (h *APIHandler) GetApiV1HealthMenstruation(c *gin.Context, params api.GetApiV1HealthMenstruationParams) {
	h.health.GetApiV1HealthMenstruation(c, params)
}

func (h *APIHandler) PostApiV1HealthMenstruation(c *gin.Context) {
	h.health.PostApiV1HealthMenstruation(c)
}

// Report endpoints
func (h *APIHandler) PostApiV1ReportsGenerate(c *gin.Context) {
	h.report.PostApiV1ReportsGenerate(c)
}

func (h *APIHandler) GetApiV1ReportsId(c *gin.Context, id openapi_types.UUID) {
	h.report.GetApiV1ReportsId(c, id)
}

// GetHealth implements the health check endpoint
// Requirements: Deployment, 12.2
func (h *APIHandler) GetHealth(c *gin.Context) {
	ctx := c.Request.Context()

	// Check database connectivity
	if err := h.pool.Ping(ctx); err != nil {
		h.logger.Error("health check failed: database unreachable", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":   "unhealthy",
			"database": "disconnected",
			"error":    err.Error(),
		})
		return
	}

	// Return healthy status
	c.JSON(http.StatusOK, gin.H{
		"status":   "healthy",
		"database": "connected",
		"service":  "eva-health-backend",
		"version":  "1.0.0",
	})
}
