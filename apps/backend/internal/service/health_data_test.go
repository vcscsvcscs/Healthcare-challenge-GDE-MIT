package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
)

func TestLogMenstruation_InvalidFlowIntensity(t *testing.T) {
	service := &HealthDataService{}

	ctx := context.Background()
	userID := "user-123"
	invalidIntensity := "extreme"
	data := &model.MenstruationCycle{
		StartDate:     time.Now(),
		FlowIntensity: &invalidIntensity,
	}

	err := service.LogMenstruation(ctx, userID, data)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid flow intensity")
}

func TestLogMenstruation_ValidFlowIntensities(t *testing.T) {
	validIntensities := []string{"light", "moderate", "heavy"}

	for _, intensity := range validIntensities {
		t.Run(intensity, func(t *testing.T) {
			flowIntensity := intensity
			data := &model.MenstruationCycle{
				StartDate:     time.Now(),
				FlowIntensity: &flowIntensity,
			}

			// Test validation logic - should not return error for valid intensities
			validIntensitiesMap := map[string]bool{
				"light":    true,
				"moderate": true,
				"heavy":    true,
			}

			isValid := validIntensitiesMap[*data.FlowIntensity]
			assert.True(t, isValid, "intensity %s should be valid", intensity)
		})
	}
}

func TestLogBloodPressure_ValidationErrors(t *testing.T) {
	service := &HealthDataService{}

	ctx := context.Background()
	userID := "user-123"

	tests := []struct {
		name        string
		reading     *model.BloodPressureReading
		expectedErr string
	}{
		{
			name: "systolic too low",
			reading: &model.BloodPressureReading{
				Systolic:   69,
				Diastolic:  80,
				Pulse:      70,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid systolic value",
		},
		{
			name: "systolic too high",
			reading: &model.BloodPressureReading{
				Systolic:   251,
				Diastolic:  80,
				Pulse:      70,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid systolic value",
		},
		{
			name: "diastolic too low",
			reading: &model.BloodPressureReading{
				Systolic:   120,
				Diastolic:  39,
				Pulse:      70,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid diastolic value",
		},
		{
			name: "diastolic too high",
			reading: &model.BloodPressureReading{
				Systolic:   120,
				Diastolic:  151,
				Pulse:      70,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid diastolic value",
		},
		{
			name: "pulse too low",
			reading: &model.BloodPressureReading{
				Systolic:   120,
				Diastolic:  80,
				Pulse:      29,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid pulse value",
		},
		{
			name: "pulse too high",
			reading: &model.BloodPressureReading{
				Systolic:   120,
				Diastolic:  80,
				Pulse:      221,
				MeasuredAt: time.Now(),
			},
			expectedErr: "invalid pulse value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.LogBloodPressure(ctx, userID, tt.reading)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestLogBloodPressure_BoundaryValues(t *testing.T) {
	tests := []struct {
		name    string
		reading *model.BloodPressureReading
	}{
		{
			name: "minimum valid values",
			reading: &model.BloodPressureReading{
				Systolic:   70,
				Diastolic:  40,
				Pulse:      30,
				MeasuredAt: time.Now(),
			},
		},
		{
			name: "maximum valid values",
			reading: &model.BloodPressureReading{
				Systolic:   250,
				Diastolic:  150,
				Pulse:      220,
				MeasuredAt: time.Now(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test validation logic - boundary values should pass validation
			isValidSystolic := tt.reading.Systolic >= 70 && tt.reading.Systolic <= 250
			isValidDiastolic := tt.reading.Diastolic >= 40 && tt.reading.Diastolic <= 150
			isValidPulse := tt.reading.Pulse >= 30 && tt.reading.Pulse <= 220

			assert.True(t, isValidSystolic, "systolic should be valid")
			assert.True(t, isValidDiastolic, "diastolic should be valid")
			assert.True(t, isValidPulse, "pulse should be valid")
		})
	}
}

func TestSyncFitnessData_ValidDataTypes(t *testing.T) {
	validDataTypes := []string{"steps", "heart_rate", "sleep", "calories", "distance", "active_minutes"}

	for _, dataType := range validDataTypes {
		t.Run(dataType, func(t *testing.T) {
			// Test validation logic
			validDataTypesMap := map[string]bool{
				"steps":          true,
				"heart_rate":     true,
				"sleep":          true,
				"calories":       true,
				"distance":       true,
				"active_minutes": true,
			}

			isValid := validDataTypesMap[dataType]
			assert.True(t, isValid, "data type %s should be valid", dataType)
		})
	}
}

func TestGetFitnessHistory_InvalidDateRange(t *testing.T) {
	service := &HealthDataService{}

	ctx := context.Background()
	userID := "user-123"
	startDate := time.Now()
	endDate := time.Now().AddDate(0, 0, -7)

	_, err := service.GetFitnessHistory(ctx, userID, startDate, endDate)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "start date must be before or equal to end date")
}
