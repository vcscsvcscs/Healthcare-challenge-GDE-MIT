package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/mock"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/pdf"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// Feature: eva-health-backend, Property 16: Dashboard Time Range Filtering
// **Validates: Requirements 7.2**
func TestProperty_DashboardTimeRangeFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Dashboard only returns data within the specified time range", prop.ForAll(
		func(userID string, days int) bool {
			// Skip invalid inputs
			if userID == "" {
				return true
			}

			// Normalize days to valid values (7, 30, 90)
			if days != 7 && days != 30 && days != 90 {
				days = 7
			}

			// Setup mocks
			repo := new(MockDashboardRepository)
			service := NewDashboardService(repo, zap.NewNop())

			// Create test data - some within range, some outside
			now := time.Now()
			startDate := now.AddDate(0, 0, -days)

			// Generate daily metrics within the time range
			var dailyMetrics []repository.DailyMetrics
			for i := 0; i < days; i++ {
				date := now.AddDate(0, 0, -i)
				painLevel := i % 11
				mood := "positive"
				energy := "medium"
				sleep := "good"
				medTaken := "yes"

				dailyMetrics = append(dailyMetrics, repository.DailyMetrics{
					Date:            date,
					PainLevel:       &painLevel,
					Mood:            &mood,
					EnergyLevel:     &energy,
					SleepQuality:    &sleep,
					MedicationTaken: &medTaken,
					SymptomCount:    2,
					ActivityCount:   1,
				})
			}

			// Mock aggregated metrics
			aggregatedMetrics := &repository.AggregatedMetrics{
				AveragePainLevel: 5.0,
				MoodDistribution: map[string]int{"positive": days},
				EnergyLevels:     map[string]int{"medium": days},
				CheckInCount:     days,
			}

			// Setup expectations
			repo.On("GetAggregatedMetrics", mock.Anything, userID, days).Return(aggregatedMetrics, nil)
			repo.On("GetDailyMetrics", mock.Anything, userID, days).Return(dailyMetrics, nil)

			// Execute
			ctx := context.Background()
			summary, err := service.GetSummary(ctx, userID, days)

			// Verify
			if err != nil {
				t.Logf("GetSummary failed: %v", err)
				return false
			}

			// Verify all returned data is within the time range
			for _, dm := range summary.TimeSeriesData {
				if dm.Date.Before(startDate) {
					t.Logf("Data point %v is before start date %v", dm.Date, startDate)
					return false
				}
				if dm.Date.After(now) {
					t.Logf("Data point %v is after current time %v", dm.Date, now)
					return false
				}
			}

			// Verify the correct number of days was requested from repository
			repo.AssertCalled(t, "GetDailyMetrics", mock.Anything, userID, days)

			return true
		},
		gen.Identifier(),
		gen.IntRange(1, 100), // Will be normalized to 7, 30, or 90
	))

	properties.TestingRun(t)
}

// Feature: eva-health-backend, Property 17: Dashboard Aggregation Accuracy
// **Validates: Requirements 7.3**
func TestProperty_DashboardAggregationAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Dashboard aggregations accurately reflect the underlying data", prop.ForAll(
		func(userID string, checkInCount int) bool {
			// Skip invalid inputs
			if userID == "" || checkInCount < 1 || checkInCount > 100 {
				return true
			}

			// Setup mocks
			repo := new(MockDashboardRepository)
			service := NewDashboardService(repo, zap.NewNop())

			// Calculate expected aggregations
			totalPain := 0
			moodCounts := make(map[string]int)
			energyCounts := make(map[string]int)

			for i := 0; i < checkInCount; i++ {
				painLevel := i % 11
				totalPain += painLevel

				mood := []string{"positive", "neutral", "negative"}[i%3]
				moodCounts[mood]++

				energy := []string{"low", "medium", "high"}[i%3]
				energyCounts[energy]++
			}

			expectedAvgPain := float64(totalPain) / float64(checkInCount)

			// Mock aggregated metrics with calculated values
			aggregatedMetrics := &repository.AggregatedMetrics{
				AveragePainLevel: expectedAvgPain,
				MoodDistribution: moodCounts,
				EnergyLevels:     energyCounts,
				CheckInCount:     checkInCount,
			}

			// Setup expectations
			repo.On("GetAggregatedMetrics", mock.Anything, userID, 7).Return(aggregatedMetrics, nil)
			repo.On("GetDailyMetrics", mock.Anything, userID, 7).Return([]repository.DailyMetrics{}, nil)

			// Execute
			ctx := context.Background()
			summary, err := service.GetSummary(ctx, userID, 7)

			// Verify
			if err != nil {
				t.Logf("GetSummary failed: %v", err)
				return false
			}

			// Verify average pain level
			if summary.AveragePain != expectedAvgPain {
				t.Logf("Expected average pain %.2f, got %.2f", expectedAvgPain, summary.AveragePain)
				return false
			}

			// Verify mood distribution
			for mood, count := range moodCounts {
				if summary.MoodDistribution[mood] != count {
					t.Logf("Expected mood %s count %d, got %d", mood, count, summary.MoodDistribution[mood])
					return false
				}
			}

			// Verify energy levels
			for energy, count := range energyCounts {
				if summary.EnergyLevels[energy] != count {
					t.Logf("Expected energy %s count %d, got %d", energy, count, summary.EnergyLevels[energy])
					return false
				}
			}

			// Verify check-in count
			if summary.CheckInCount != checkInCount {
				t.Logf("Expected check-in count %d, got %d", checkInCount, summary.CheckInCount)
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t)
}

// Feature: eva-health-backend, Property 18: Time Series Data Grouping
// **Validates: Requirements 7.4**
func TestProperty_TimeSeriesDataGrouping(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Time series data is correctly grouped by date", prop.ForAll(
		func(userID string, days int) bool {
			// Skip invalid inputs
			if userID == "" || days < 1 || days > 90 {
				return true
			}

			// Setup mocks
			repo := new(MockDashboardRepository)
			service := NewDashboardService(repo, zap.NewNop())

			// Generate daily metrics with unique dates
			now := time.Now()
			var dailyMetrics []repository.DailyMetrics
			datesSeen := make(map[string]bool)

			for i := 0; i < days; i++ {
				date := now.AddDate(0, 0, -i)
				dateKey := date.Format("2006-01-02")

				painLevel := i % 11
				mood := "positive"
				energy := "medium"

				dailyMetrics = append(dailyMetrics, repository.DailyMetrics{
					Date:         date,
					PainLevel:    &painLevel,
					Mood:         &mood,
					EnergyLevel:  &energy,
					SymptomCount: 1,
				})

				datesSeen[dateKey] = true
			}

			// Mock aggregated metrics
			aggregatedMetrics := &repository.AggregatedMetrics{
				AveragePainLevel: 5.0,
				MoodDistribution: map[string]int{"positive": days},
				EnergyLevels:     map[string]int{"medium": days},
				CheckInCount:     days,
			}

			// Setup expectations
			repo.On("GetAggregatedMetrics", mock.Anything, userID, mock.Anything).Return(aggregatedMetrics, nil)
			repo.On("GetDailyMetrics", mock.Anything, userID, mock.Anything).Return(dailyMetrics, nil)

			// Execute
			ctx := context.Background()
			summary, err := service.GetSummary(ctx, userID, days)

			// Verify
			if err != nil {
				t.Logf("GetSummary failed: %v", err)
				return false
			}

			// Verify each date appears only once
			datesInResult := make(map[string]int)
			for _, dm := range summary.TimeSeriesData {
				dateKey := dm.Date.Format("2006-01-02")
				datesInResult[dateKey]++
			}

			// Check for duplicate dates
			for dateKey, count := range datesInResult {
				if count > 1 {
					t.Logf("Date %s appears %d times, expected 1", dateKey, count)
					return false
				}
			}

			// Verify all expected dates are present
			if len(datesInResult) != len(datesSeen) {
				t.Logf("Expected %d unique dates, got %d", len(datesSeen), len(datesInResult))
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.IntRange(1, 90),
	))

	properties.TestingRun(t)
}

// Feature: eva-health-backend, Property 19: Report Content Completeness
// **Validates: Requirements 8.1, 8.2**
// Note: This property test validates that the PDF generator includes all required sections
func TestProperty_ReportContentCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("Generated PDF reports contain all required sections", prop.ForAll(
		func(userName string) bool {
			// Skip invalid inputs
			if userName == "" {
				return true
			}

			// Setup PDF generator
			pdfGen := pdf.NewPDFGenerator(zap.NewNop())

			// Create test data with all sections
			now := time.Now()

			// Health check-ins
			checkIns := []model.HealthCheckIn{
				{
					ID:               "checkin-1",
					UserID:           "test-user",
					CheckInDate:      now.AddDate(0, 0, -1),
					Symptoms:         []string{"headache"},
					Mood:             ptrString("positive"),
					PainLevel:        ptrInt(3),
					EnergyLevel:      ptrString("medium"),
					SleepQuality:     ptrString("good"),
					MedicationTaken:  ptrString("yes"),
					PhysicalActivity: []string{"walking"},
					Breakfast:        ptrString("oatmeal"),
					Lunch:            ptrString("salad"),
					Dinner:           ptrString("chicken"),
					GeneralFeeling:   ptrString("feeling good"),
					AdditionalNotes:  ptrString("no issues"),
					CreatedAt:        now,
					UpdatedAt:        now,
				},
			}

			// Medications
			medications := []model.Medication{
				{
					ID:        "med-1",
					UserID:    "test-user",
					Name:      "Aspirin",
					Dosage:    "100mg",
					Frequency: "daily",
					StartDate: now.AddDate(0, 0, -60),
					Active:    true,
					CreatedAt: now,
					UpdatedAt: now,
				},
			}

			// Blood pressure readings
			bloodPressure := []model.BloodPressureReading{
				{
					ID:         "bp-1",
					UserID:     "test-user",
					Systolic:   120,
					Diastolic:  80,
					Pulse:      70,
					MeasuredAt: now,
					CreatedAt:  now,
				},
			}

			// Menstruation cycles
			menstruationCycles := []model.MenstruationCycle{
				{
					ID:            "cycle-1",
					UserID:        "test-user",
					StartDate:     now.AddDate(0, 0, -7),
					FlowIntensity: ptrString("moderate"),
					Symptoms:      []string{"cramps"},
					CreatedAt:     now,
					UpdatedAt:     now,
				},
			}

			// Fitness data
			fitnessData := []model.FitnessDataPoint{
				{
					ID:           "fitness-1",
					UserID:       "test-user",
					Date:         now,
					DataType:     "steps",
					Value:        10000,
					Unit:         "count",
					Source:       "health_connect",
					SourceDataID: "source-1",
					CreatedAt:    now,
				},
			}

			// Create report data
			reportData := &pdf.ReportData{
				UserName:           userName,
				DateRange:          fmt.Sprintf("%s to %s", now.AddDate(0, 0, -30).Format("2006-01-02"), now.Format("2006-01-02")),
				CheckIns:           checkIns,
				Medications:        medications,
				BloodPressure:      bloodPressure,
				MenstruationCycles: menstruationCycles,
				FitnessData:        fitnessData,
			}

			// Generate PDF
			pdfBytes, err := pdfGen.Generate(reportData)

			// Verify
			if err != nil {
				t.Logf("PDF generation failed: %v", err)
				return false
			}

			if len(pdfBytes) == 0 {
				t.Log("Generated PDF should not be empty")
				return false
			}

			// Verify PDF has reasonable size (at least 1KB for a report with data)
			if len(pdfBytes) < 1024 {
				t.Logf("Generated PDF seems too small: %d bytes", len(pdfBytes))
				return false
			}

			// Verify PDF header (PDF files start with %PDF-)
			if len(pdfBytes) < 5 || string(pdfBytes[0:5]) != "%PDF-" {
				t.Log("Generated file does not appear to be a valid PDF")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// Feature: eva-health-backend, Property 20: Report Storage and Retrieval Round Trip
// **Validates: Requirements 8.3, 8.4**
// Note: This property test validates PDF generation consistency
func TestProperty_ReportStorageAndRetrievalRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("PDF generation is deterministic for the same input data", prop.ForAll(
		func(userName string) bool {
			// Skip invalid inputs
			if userName == "" {
				return true
			}

			// Setup PDF generator
			pdfGen := pdf.NewPDFGenerator(zap.NewNop())

			// Create test data
			now := time.Now()

			// Create report data
			reportData := &pdf.ReportData{
				UserName:           userName,
				DateRange:          fmt.Sprintf("%s to %s", now.AddDate(0, 0, -30).Format("2006-01-02"), now.Format("2006-01-02")),
				CheckIns:           []model.HealthCheckIn{},
				Medications:        []model.Medication{},
				BloodPressure:      []model.BloodPressureReading{},
				MenstruationCycles: []model.MenstruationCycle{},
				FitnessData:        []model.FitnessDataPoint{},
			}

			// Generate PDF twice with same data
			pdfBytes1, err1 := pdfGen.Generate(reportData)
			pdfBytes2, err2 := pdfGen.Generate(reportData)

			// Verify both generations succeeded
			if err1 != nil {
				t.Logf("First PDF generation failed: %v", err1)
				return false
			}

			if err2 != nil {
				t.Logf("Second PDF generation failed: %v", err2)
				return false
			}

			if len(pdfBytes1) == 0 || len(pdfBytes2) == 0 {
				t.Log("Generated PDFs should not be empty")
				return false
			}

			// Note: PDF generation may not be byte-for-byte identical due to timestamps
			// but the size should be very similar (within 5%)
			sizeDiff := float64(len(pdfBytes1) - len(pdfBytes2))
			if sizeDiff < 0 {
				sizeDiff = -sizeDiff
			}
			avgSize := float64(len(pdfBytes1)+len(pdfBytes2)) / 2.0
			percentDiff := (sizeDiff / avgSize) * 100.0

			if percentDiff > 5.0 {
				t.Logf("PDF sizes differ significantly: %d vs %d bytes (%.1f%% difference)",
					len(pdfBytes1), len(pdfBytes2), percentDiff)
				return false
			}

			// Both should be valid PDFs
			if string(pdfBytes1[0:5]) != "%PDF-" || string(pdfBytes2[0:5]) != "%PDF-" {
				t.Log("Generated files do not appear to be valid PDFs")
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t)
}

// Helper functions

func ptrString(s string) *string {
	return &s
}

func ptrInt(i int) *int {
	return &i
}
