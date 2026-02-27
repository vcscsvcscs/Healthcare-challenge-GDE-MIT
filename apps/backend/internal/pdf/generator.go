package pdf

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// PDFGenerator generates professional medical reports
type PDFGenerator struct {
	logger *zap.Logger
}

// NewPDFGenerator creates a new PDFGenerator
func NewPDFGenerator(logger *zap.Logger) *PDFGenerator {
	return &PDFGenerator{
		logger: logger,
	}
}

// ReportData contains all data needed for report generation
type ReportData struct {
	UserName           string
	DateRange          string
	CheckIns           []model.HealthCheckIn
	Medications        []model.Medication
	BloodPressure      []model.BloodPressureReading
	MenstruationCycles []model.MenstruationCycle
	FitnessData        []model.FitnessDataPoint
}

// Generate creates a PDF report from the provided data
func (g *PDFGenerator) Generate(data *ReportData) ([]byte, error) {
	g.logger.Info("generating PDF report",
		zap.String("user_name", data.UserName),
		zap.String("date_range", data.DateRange),
	)

	// Create new PDF
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.SetAutoPageBreak(true, 20)

	// Add page
	pdf.AddPage()

	// Add title
	g.addTitle(pdf, "Health Report", data.UserName, data.DateRange)

	// Add all sections
	g.addSymptomsTimeline(pdf, data.CheckIns)
	g.addMedicationList(pdf, data.Medications)
	g.addMedicationAdherence(pdf, data.CheckIns)
	g.addBloodPressureTrends(pdf, data.BloodPressure)
	g.addMenstruationCycles(pdf, data.MenstruationCycles)
	g.addPhysicalActivities(pdf, data.CheckIns)
	g.addMealPatterns(pdf, data.CheckIns)
	g.addDailyCheckInSummaries(pdf, data.CheckIns)

	// Generate PDF bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		g.logger.Error("failed to generate PDF", zap.Error(err))
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	g.logger.Info("PDF report generated successfully",
		zap.Int("size_bytes", buf.Len()),
	)

	return buf.Bytes(), nil
}

// addTitle adds the report title and header information
func (g *PDFGenerator) addTitle(pdf *gofpdf.Fpdf, title, userName, dateRange string) {
	pdf.SetFont("Arial", "B", 20)
	pdf.CellFormat(0, 10, title, "", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(0, 8, fmt.Sprintf("Patient: %s", userName), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Period: %s", dateRange), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 8, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04")), "", 1, "L", false, 0, "")
	pdf.Ln(10)
}

// addSectionHeader adds a section header
func (g *PDFGenerator) addSectionHeader(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 14)
	pdf.SetFillColor(230, 230, 230)
	pdf.CellFormat(0, 10, title, "", 1, "L", true, 0, "")
	pdf.Ln(3)
	pdf.SetFont("Arial", "", 10)
}

// addSymptomsTimeline adds symptoms timeline section
func (g *PDFGenerator) addSymptomsTimeline(pdf *gofpdf.Fpdf, checkIns []model.HealthCheckIn) {
	g.addSectionHeader(pdf, "Symptoms Timeline")

	if len(checkIns) == 0 {
		pdf.CellFormat(0, 8, "No symptoms recorded during this period.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	for _, checkIn := range checkIns {
		if len(checkIn.Symptoms) > 0 {
			dateStr := checkIn.CheckInDate.Format("2006-01-02")
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(0, 6, dateStr, "", 1, "L", false, 0, "")
			pdf.SetFont("Arial", "", 10)

			for _, symptom := range checkIn.Symptoms {
				pdf.CellFormat(0, 5, fmt.Sprintf("  - %s", symptom), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}
	}
	pdf.Ln(5)
}

// addMedicationList adds medication list section
func (g *PDFGenerator) addMedicationList(pdf *gofpdf.Fpdf, medications []model.Medication) {
	g.addSectionHeader(pdf, "Medication List")

	if len(medications) == 0 {
		pdf.CellFormat(0, 8, "No medications recorded.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	for _, med := range medications {
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, med.Name, "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(0, 5, fmt.Sprintf("  Dosage: %s", med.Dosage), "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 5, fmt.Sprintf("  Frequency: %s", med.Frequency), "", 1, "L", false, 0, "")
		pdf.CellFormat(0, 5, fmt.Sprintf("  Start Date: %s", med.StartDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
		if med.EndDate != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  End Date: %s", med.EndDate.Format("2006-01-02")), "", 1, "L", false, 0, "")
		}
		if med.Notes != nil && *med.Notes != "" {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Notes: %s", *med.Notes), "", 1, "L", false, 0, "")
		}
		pdf.Ln(3)
	}
	pdf.Ln(5)
}

// addMedicationAdherence adds medication adherence section
func (g *PDFGenerator) addMedicationAdherence(pdf *gofpdf.Fpdf, checkIns []model.HealthCheckIn) {
	g.addSectionHeader(pdf, "Medication Adherence")

	if len(checkIns) == 0 {
		pdf.CellFormat(0, 8, "No adherence data recorded.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	adherenceCount := make(map[string]int)
	for _, checkIn := range checkIns {
		if checkIn.MedicationTaken != nil {
			adherenceCount[*checkIn.MedicationTaken]++
		}
	}

	for status, count := range adherenceCount {
		pdf.CellFormat(0, 6, fmt.Sprintf("%s: %d days", status, count), "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

// addBloodPressureTrends adds blood pressure trends section
func (g *PDFGenerator) addBloodPressureTrends(pdf *gofpdf.Fpdf, readings []model.BloodPressureReading) {
	g.addSectionHeader(pdf, "Blood Pressure Trends")

	if len(readings) == 0 {
		pdf.CellFormat(0, 8, "No blood pressure readings recorded.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	// Calculate averages
	var totalSystolic, totalDiastolic, totalPulse int
	for _, reading := range readings {
		totalSystolic += reading.Systolic
		totalDiastolic += reading.Diastolic
		totalPulse += reading.Pulse
	}

	count := len(readings)
	avgSystolic := float64(totalSystolic) / float64(count)
	avgDiastolic := float64(totalDiastolic) / float64(count)
	avgPulse := float64(totalPulse) / float64(count)

	pdf.CellFormat(0, 6, fmt.Sprintf("Average: %.0f/%.0f mmHg, Pulse: %.0f bpm", avgSystolic, avgDiastolic, avgPulse), "", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Total readings: %d", count), "", 1, "L", false, 0, "")
	pdf.Ln(3)

	// List recent readings
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(0, 6, "Recent Readings:", "", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 10)

	maxReadings := 10
	if len(readings) < maxReadings {
		maxReadings = len(readings)
	}

	for i := 0; i < maxReadings; i++ {
		reading := readings[i]
		dateStr := reading.MeasuredAt.Format("2006-01-02 15:04")
		pdf.CellFormat(0, 5, fmt.Sprintf("%s: %d/%d mmHg, Pulse: %d bpm",
			dateStr, reading.Systolic, reading.Diastolic, reading.Pulse), "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

// addMenstruationCycles adds menstruation cycles section
func (g *PDFGenerator) addMenstruationCycles(pdf *gofpdf.Fpdf, cycles []model.MenstruationCycle) {
	g.addSectionHeader(pdf, "Menstruation Cycles")

	if len(cycles) == 0 {
		pdf.CellFormat(0, 8, "No menstruation data recorded.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	for _, cycle := range cycles {
		startStr := cycle.StartDate.Format("2006-01-02")
		endStr := "ongoing"
		if cycle.EndDate != nil {
			endStr = cycle.EndDate.Format("2006-01-02")
		}

		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, fmt.Sprintf("%s to %s", startStr, endStr), "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)

		if cycle.FlowIntensity != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Flow: %s", *cycle.FlowIntensity), "", 1, "L", false, 0, "")
		}

		if len(cycle.Symptoms) > 0 {
			pdf.CellFormat(0, 5, "  Symptoms:", "", 1, "L", false, 0, "")
			for _, symptom := range cycle.Symptoms {
				pdf.CellFormat(0, 5, fmt.Sprintf("    - %s", symptom), "", 1, "L", false, 0, "")
			}
		}
		pdf.Ln(3)
	}
	pdf.Ln(5)
}

// addPhysicalActivities adds physical activities section
func (g *PDFGenerator) addPhysicalActivities(pdf *gofpdf.Fpdf, checkIns []model.HealthCheckIn) {
	g.addSectionHeader(pdf, "Physical Activities")

	activitiesFound := false
	for _, checkIn := range checkIns {
		if len(checkIn.PhysicalActivity) > 0 {
			activitiesFound = true
			dateStr := checkIn.CheckInDate.Format("2006-01-02")
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(0, 6, dateStr, "", 1, "L", false, 0, "")
			pdf.SetFont("Arial", "", 10)

			for _, activity := range checkIn.PhysicalActivity {
				pdf.CellFormat(0, 5, fmt.Sprintf("  - %s", activity), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}
	}

	if !activitiesFound {
		pdf.CellFormat(0, 8, "No physical activities recorded.", "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

// addMealPatterns adds meal patterns section
func (g *PDFGenerator) addMealPatterns(pdf *gofpdf.Fpdf, checkIns []model.HealthCheckIn) {
	g.addSectionHeader(pdf, "Meal Patterns")

	mealsFound := false
	for _, checkIn := range checkIns {
		if (checkIn.Breakfast != nil && *checkIn.Breakfast != "") ||
			(checkIn.Lunch != nil && *checkIn.Lunch != "") ||
			(checkIn.Dinner != nil && *checkIn.Dinner != "") {
			mealsFound = true
			dateStr := checkIn.CheckInDate.Format("2006-01-02")
			pdf.SetFont("Arial", "B", 10)
			pdf.CellFormat(0, 6, dateStr, "", 1, "L", false, 0, "")
			pdf.SetFont("Arial", "", 10)

			if checkIn.Breakfast != nil && *checkIn.Breakfast != "" {
				pdf.CellFormat(0, 5, fmt.Sprintf("  Breakfast: %s", *checkIn.Breakfast), "", 1, "L", false, 0, "")
			}
			if checkIn.Lunch != nil && *checkIn.Lunch != "" {
				pdf.CellFormat(0, 5, fmt.Sprintf("  Lunch: %s", *checkIn.Lunch), "", 1, "L", false, 0, "")
			}
			if checkIn.Dinner != nil && *checkIn.Dinner != "" {
				pdf.CellFormat(0, 5, fmt.Sprintf("  Dinner: %s", *checkIn.Dinner), "", 1, "L", false, 0, "")
			}
			pdf.Ln(2)
		}
	}

	if !mealsFound {
		pdf.CellFormat(0, 8, "No meal data recorded.", "", 1, "L", false, 0, "")
	}
	pdf.Ln(5)
}

// addDailyCheckInSummaries adds daily check-in summaries section
func (g *PDFGenerator) addDailyCheckInSummaries(pdf *gofpdf.Fpdf, checkIns []model.HealthCheckIn) {
	g.addSectionHeader(pdf, "Daily Check-In Summaries")

	if len(checkIns) == 0 {
		pdf.CellFormat(0, 8, "No check-ins recorded during this period.", "", 1, "L", false, 0, "")
		pdf.Ln(5)
		return
	}

	for _, checkIn := range checkIns {
		dateStr := checkIn.CheckInDate.Format("2006-01-02")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(0, 6, dateStr, "", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 10)

		if checkIn.Mood != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Mood: %s", *checkIn.Mood), "", 1, "L", false, 0, "")
		}
		if checkIn.EnergyLevel != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Energy: %s", *checkIn.EnergyLevel), "", 1, "L", false, 0, "")
		}
		if checkIn.SleepQuality != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Sleep: %s", *checkIn.SleepQuality), "", 1, "L", false, 0, "")
		}
		if checkIn.PainLevel != nil {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Pain Level: %d/10", *checkIn.PainLevel), "", 1, "L", false, 0, "")
		}
		if checkIn.GeneralFeeling != nil && *checkIn.GeneralFeeling != "" {
			pdf.CellFormat(0, 5, fmt.Sprintf("  General Feeling: %s", *checkIn.GeneralFeeling), "", 1, "L", false, 0, "")
		}
		if checkIn.AdditionalNotes != nil && *checkIn.AdditionalNotes != "" {
			pdf.CellFormat(0, 5, fmt.Sprintf("  Notes: %s", *checkIn.AdditionalNotes), "", 1, "L", false, 0, "")
		}
		pdf.Ln(3)
	}
	pdf.Ln(5)
}
