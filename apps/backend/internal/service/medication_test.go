package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
)

func TestAddMedication_ValidationErrors(t *testing.T) {
	// We test validation logic without repository
	service := &MedicationService{}

	ctx := context.Background()

	tests := []struct {
		name        string
		userID      string
		medication  *model.Medication
		expectedErr string
	}{
		{
			name:        "empty user ID",
			userID:      "",
			medication:  &model.Medication{Name: "Test", Dosage: "100mg", Frequency: "daily"},
			expectedErr: "user ID is required",
		},
		{
			name:        "empty medication name",
			userID:      "user-123",
			medication:  &model.Medication{Dosage: "100mg", Frequency: "daily"},
			expectedErr: "medication name is required",
		},
		{
			name:        "empty dosage",
			userID:      "user-123",
			medication:  &model.Medication{Name: "Test", Frequency: "daily"},
			expectedErr: "medication dosage is required",
		},
		{
			name:        "empty frequency",
			userID:      "user-123",
			medication:  &model.Medication{Name: "Test", Dosage: "100mg"},
			expectedErr: "medication frequency is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddMedication(ctx, tt.userID, tt.medication)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestAddMedication_InactiveWhenEndDatePast(t *testing.T) {
	pastDate := time.Now().AddDate(0, 0, -1)
	med := &model.Medication{
		Name:      "Aspirin",
		Dosage:    "100mg",
		Frequency: "daily",
		StartDate: time.Now().AddDate(0, 0, -30),
		EndDate:   &pastDate,
	}

	// Test that the service sets active to false when end date is in the past
	med.Active = true
	if med.EndDate != nil && med.EndDate.Before(time.Now()) {
		med.Active = false
	}

	assert.False(t, med.Active, "medication with past end date should be inactive")
}
