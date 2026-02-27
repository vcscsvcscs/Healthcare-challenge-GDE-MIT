package pdf

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

func TestPDFGenerator_Generate_Success(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	generator := NewPDFGenerator(logger)

	painLevel := 5
	mood := "positive"
	energyLevel := "high"
	sleepQuality := "good"
	medicationTaken := "yes"
	breakfast := "Oatmeal with fruits"
	generalFeeling := "Feeling good today"

	reportData := &ReportData{
		UserName:  "Test User",
		DateRange: "2024-01-01 to 2024-01-31",
		CheckIns: []model.HealthCheckIn{
			{
				ID:               "checkin-1",
				UserID:           "user-1",
				CheckInDate:      time.Now().AddDate(0, 0, -1),
				Symptoms:         []string{"headache", "fatigue"},
				Mood:             &mood,
				PainLevel:        &painLevel,
				EnergyLevel:      &energyLevel,
				SleepQuality:     &sleepQuality,
				MedicationTaken:  &medicationTaken,
				PhysicalActivity: []string{"walking", "yoga"},
				Breakfast:        &breakfast,
				GeneralFeeling:   &generalFeeling,
			},
		},
		Medications: []model.Medication{
			{
				ID:        "med-1",
				UserID:    "user-1",
				Name:      "Aspirin",
				Dosage:    "100mg",
				Frequency: "Daily",
				StartDate: time.Now().AddDate(0, -1, 0),
				Active:    true,
			},
		},
		BloodPressure: []model.BloodPressureReading{
			{
				ID:         "bp-1",
				UserID:     "user-1",
				Systolic:   120,
				Diastolic:  80,
				Pulse:      70,
				MeasuredAt: time.Now().AddDate(0, 0, -1),
			},
		},
		MenstruationCycles: []model.MenstruationCycle{},
		FitnessData:        []model.FitnessDataPoint{},
	}

	// Act
	pdfBytes, err := generator.Generate(reportData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pdfBytes)
	assert.Greater(t, len(pdfBytes), 0, "PDF should have content")

	// PDF files start with %PDF
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")
}

func TestPDFGenerator_Generate_EmptyData(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	generator := NewPDFGenerator(logger)

	reportData := &ReportData{
		UserName:           "Test User",
		DateRange:          "2024-01-01 to 2024-01-31",
		CheckIns:           []model.HealthCheckIn{},
		Medications:        []model.Medication{},
		BloodPressure:      []model.BloodPressureReading{},
		MenstruationCycles: []model.MenstruationCycle{},
		FitnessData:        []model.FitnessDataPoint{},
	}

	// Act
	pdfBytes, err := generator.Generate(reportData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pdfBytes)
	assert.Greater(t, len(pdfBytes), 0, "PDF should have content even with empty data")

	// PDF files start with %PDF
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")
}

func TestPDFGenerator_Generate_WithMenstruationData(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	generator := NewPDFGenerator(logger)

	flowIntensity := "moderate"
	endDate := time.Now().AddDate(0, 0, -3)

	reportData := &ReportData{
		UserName:      "Test User",
		DateRange:     "2024-01-01 to 2024-01-31",
		CheckIns:      []model.HealthCheckIn{},
		Medications:   []model.Medication{},
		BloodPressure: []model.BloodPressureReading{},
		MenstruationCycles: []model.MenstruationCycle{
			{
				ID:            "cycle-1",
				UserID:        "user-1",
				StartDate:     time.Now().AddDate(0, 0, -7),
				EndDate:       &endDate,
				FlowIntensity: &flowIntensity,
				Symptoms:      []string{"cramps", "fatigue"},
			},
		},
		FitnessData: []model.FitnessDataPoint{},
	}

	// Act
	pdfBytes, err := generator.Generate(reportData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pdfBytes)
	assert.Greater(t, len(pdfBytes), 0, "PDF should have content")
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")
}

func TestPDFGenerator_Generate_WithMultipleBloodPressureReadings(t *testing.T) {
	// Arrange
	logger := zap.NewNop()
	generator := NewPDFGenerator(logger)

	reportData := &ReportData{
		UserName:    "Test User",
		DateRange:   "2024-01-01 to 2024-01-31",
		CheckIns:    []model.HealthCheckIn{},
		Medications: []model.Medication{},
		BloodPressure: []model.BloodPressureReading{
			{
				ID:         "bp-1",
				UserID:     "user-1",
				Systolic:   120,
				Diastolic:  80,
				Pulse:      70,
				MeasuredAt: time.Now().AddDate(0, 0, -1),
			},
			{
				ID:         "bp-2",
				UserID:     "user-1",
				Systolic:   125,
				Diastolic:  82,
				Pulse:      72,
				MeasuredAt: time.Now().AddDate(0, 0, -2),
			},
			{
				ID:         "bp-3",
				UserID:     "user-1",
				Systolic:   118,
				Diastolic:  78,
				Pulse:      68,
				MeasuredAt: time.Now().AddDate(0, 0, -3),
			},
		},
		MenstruationCycles: []model.MenstruationCycle{},
		FitnessData:        []model.FitnessDataPoint{},
	}

	// Act
	pdfBytes, err := generator.Generate(reportData)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, pdfBytes)
	assert.Greater(t, len(pdfBytes), 0, "PDF should have content")
	assert.Equal(t, "%PDF", string(pdfBytes[:4]), "Should be a valid PDF file")
}
