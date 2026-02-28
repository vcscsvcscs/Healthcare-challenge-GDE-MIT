package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/pdf"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// ReportService manages health report generation
type ReportService struct {
	dashboardRepo  *repository.DashboardRepository
	healthRepo     *repository.HealthDataRepository
	medicationRepo *repository.MedicationRepository
	blobClient     *azure.BlobStorageClient
	pdfGen         *pdf.PDFGenerator
	logger         *zap.Logger
}

// NewReportService creates a new ReportService
func NewReportService(
	dashboardRepo *repository.DashboardRepository,
	healthRepo *repository.HealthDataRepository,
	medicationRepo *repository.MedicationRepository,
	blobClient *azure.BlobStorageClient,
	pdfGen *pdf.PDFGenerator,
	logger *zap.Logger,
) *ReportService {
	return &ReportService{
		dashboardRepo:  dashboardRepo,
		healthRepo:     healthRepo,
		medicationRepo: medicationRepo,
		blobClient:     blobClient,
		pdfGen:         pdfGen,
		logger:         logger,
	}
}

// GenerateReport generates a health report asynchronously
func (s *ReportService) GenerateReport(ctx context.Context, userID string, userName string, startDate, endDate time.Time) (string, error) {
	s.logger.Info("generating health report",
		zap.String("user_id", userID),
		zap.Time("start_date", startDate),
		zap.Time("end_date", endDate),
	)

	// Generate report ID
	reportID := uuid.New().String()

	// Fetch all required data
	checkIns, err := s.dashboardRepo.GetHealthCheckIns(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get health check-ins for report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return "", fmt.Errorf("failed to get health check-ins: %w", err)
	}

	medications, err := s.medicationRepo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get medications for report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return "", fmt.Errorf("failed to get medications: %w", err)
	}

	bloodPressure, err := s.healthRepo.GetBloodPressureByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get blood pressure for report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return "", fmt.Errorf("failed to get blood pressure: %w", err)
	}

	menstruationCycles, err := s.healthRepo.GetMenstruationByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get menstruation cycles for report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return "", fmt.Errorf("failed to get menstruation cycles: %w", err)
	}

	fitnessData, err := s.healthRepo.GetFitnessDataByUserID(ctx, userID, startDate, endDate)
	if err != nil {
		s.logger.Error("failed to get fitness data for report",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return "", fmt.Errorf("failed to get fitness data: %w", err)
	}

	// Prepare report data
	dateRange := fmt.Sprintf("%s to %s", startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
	reportData := &pdf.ReportData{
		UserName:           userName,
		DateRange:          dateRange,
		CheckIns:           checkIns,
		Medications:        medications,
		BloodPressure:      bloodPressure,
		MenstruationCycles: menstruationCycles,
		FitnessData:        fitnessData,
	}

	// Generate PDF
	pdfBytes, err := s.pdfGen.Generate(reportData)
	if err != nil {
		s.logger.Error("failed to generate PDF",
			zap.Error(err),
			zap.String("report_id", reportID),
		)
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	// Upload to Azure Blob Storage
	filename := fmt.Sprintf("%s_%s.pdf", reportID, time.Now().Format("20060102"))
	blobPath, err := s.blobClient.UploadPDF(ctx, filename, pdfBytes)
	if err != nil {
		s.logger.Error("failed to upload PDF to blob storage",
			zap.Error(err),
			zap.String("report_id", reportID),
		)
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	// Create report record in database
	report := &model.Report{
		ID:             reportID,
		UserID:         userID,
		DateRangeStart: startDate,
		DateRangeEnd:   endDate,
		FilePath:       blobPath,
		GeneratedAt:    time.Now(),
	}

	err = s.dashboardRepo.SaveReport(ctx, report)
	if err != nil {
		s.logger.Error("failed to save report record",
			zap.Error(err),
			zap.String("report_id", reportID),
		)
		return "", fmt.Errorf("failed to save report record: %w", err)
	}

	s.logger.Info("health report generated successfully",
		zap.String("report_id", reportID),
		zap.String("user_id", userID),
		zap.String("blob_path", blobPath),
	)

	return reportID, nil
}

// GetReport retrieves a report PDF for download
func (s *ReportService) GetReport(ctx context.Context, reportID string) ([]byte, error) {
	s.logger.Info("retrieving report",
		zap.String("report_id", reportID),
	)

	// Get report record from database
	report, err := s.dashboardRepo.GetReportByID(ctx, reportID)
	if err != nil {
		s.logger.Error("failed to get report record",
			zap.Error(err),
			zap.String("report_id", reportID),
		)
		return nil, fmt.Errorf("failed to get report record: %w", err)
	}

	// Download PDF from Azure Blob Storage
	pdfBytes, err := s.blobClient.DownloadPDF(ctx, report.FilePath)
	if err != nil {
		s.logger.Error("failed to download PDF from blob storage",
			zap.Error(err),
			zap.String("report_id", reportID),
			zap.String("blob_path", report.FilePath),
		)
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}

	s.logger.Info("report retrieved successfully",
		zap.String("report_id", reportID),
		zap.Int("size_bytes", len(pdfBytes)),
	)

	return pdfBytes, nil
}

// GetReportsByUserID retrieves all reports for a user
func (s *ReportService) GetReportsByUserID(ctx context.Context, userID string) ([]model.Report, error) {
	s.logger.Info("retrieving reports for user",
		zap.String("user_id", userID),
	)

	reports, err := s.dashboardRepo.GetReportsByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to get reports for user",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}

	s.logger.Info("reports retrieved successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(reports)),
	)

	return reports, nil
}
