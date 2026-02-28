package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// MedicationService handles medication management business logic
type MedicationService struct {
	repo   *repository.MedicationRepository
	logger *zap.Logger
}

// NewMedicationService creates a new MedicationService
func NewMedicationService(repo *repository.MedicationRepository, logger *zap.Logger) *MedicationService {
	return &MedicationService{
		repo:   repo,
		logger: logger,
	}
}

// AddMedication adds a new medication for a user
func (s *MedicationService) AddMedication(ctx context.Context, userID string, med *model.Medication) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}
	if med.Name == "" {
		return fmt.Errorf("medication name is required")
	}
	if med.Dosage == "" {
		return fmt.Errorf("medication dosage is required")
	}
	if med.Frequency == "" {
		return fmt.Errorf("medication frequency is required")
	}

	// Generate ID if not provided
	if med.ID == "" {
		med.ID = uuid.New().String()
	}

	// Set user ID
	med.UserID = userID

	// Set active status based on end date
	med.Active = true
	if med.EndDate != nil && med.EndDate.Before(time.Now()) {
		med.Active = false
	}

	// Set timestamps
	now := time.Now()
	med.CreatedAt = now
	med.UpdatedAt = now

	if err := s.repo.Create(ctx, med); err != nil {
		s.logger.Error("failed to add medication",
			zap.Error(err),
			zap.String("user_id", userID),
			zap.String("medication_name", med.Name),
		)
		return fmt.Errorf("failed to add medication: %w", err)
	}

	s.logger.Info("medication added successfully",
		zap.String("medication_id", med.ID),
		zap.String("user_id", userID),
		zap.String("name", med.Name),
	)

	return nil
}

// ListMedications retrieves all medications for a user
func (s *MedicationService) ListMedications(ctx context.Context, userID string) ([]model.Medication, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	medications, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		s.logger.Error("failed to list medications",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		return nil, fmt.Errorf("failed to list medications: %w", err)
	}

	// Update active status for medications with past end dates
	now := time.Now()
	for i := range medications {
		if medications[i].EndDate != nil && medications[i].EndDate.Before(now) && medications[i].Active {
			medications[i].Active = false
			// Update in database
			if err := s.repo.Update(ctx, &medications[i]); err != nil {
				s.logger.Warn("failed to update medication active status",
					zap.Error(err),
					zap.String("medication_id", medications[i].ID),
				)
			}
		}
	}

	s.logger.Info("medications listed successfully",
		zap.String("user_id", userID),
		zap.Int("count", len(medications)),
	)

	return medications, nil
}

// UpdateMedication updates an existing medication
func (s *MedicationService) UpdateMedication(ctx context.Context, medID string, updates *model.Medication) error {
	if medID == "" {
		return fmt.Errorf("medication ID is required")
	}

	// Fetch existing medication to preserve ID and user_id
	existing, err := s.repo.FindByID(ctx, medID)
	if err != nil {
		s.logger.Error("failed to find medication for update",
			zap.Error(err),
			zap.String("medication_id", medID),
		)
		return fmt.Errorf("medication not found: %w", err)
	}

	// Preserve ID and user_id
	updates.ID = existing.ID
	updates.UserID = existing.UserID

	// Update active status based on end date
	if updates.EndDate != nil && updates.EndDate.Before(time.Now()) {
		updates.Active = false
	} else {
		updates.Active = true
	}

	// Update timestamp
	updates.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, updates); err != nil {
		s.logger.Error("failed to update medication",
			zap.Error(err),
			zap.String("medication_id", medID),
		)
		return fmt.Errorf("failed to update medication: %w", err)
	}

	s.logger.Info("medication updated successfully",
		zap.String("medication_id", medID),
		zap.String("name", updates.Name),
	)

	return nil
}

// DeleteMedication deletes a medication
func (s *MedicationService) DeleteMedication(ctx context.Context, medID string) error {
	if medID == "" {
		return fmt.Errorf("medication ID is required")
	}

	if err := s.repo.Delete(ctx, medID); err != nil {
		s.logger.Error("failed to delete medication",
			zap.Error(err),
			zap.String("medication_id", medID),
		)
		return fmt.Errorf("failed to delete medication: %w", err)
	}

	s.logger.Info("medication deleted successfully",
		zap.String("medication_id", medID),
	)

	return nil
}

// LogAdherence logs medication adherence
func (s *MedicationService) LogAdherence(ctx context.Context, medicationID string, takenAt time.Time, adherence bool) error {
	if medicationID == "" {
		return fmt.Errorf("medication ID is required")
	}

	log := &model.MedicationLog{
		ID:           uuid.New().String(),
		MedicationID: medicationID,
		TakenAt:      takenAt,
		Adherence:    adherence,
		CreatedAt:    time.Now(),
	}

	if err := s.repo.LogAdherence(ctx, log); err != nil {
		s.logger.Error("failed to log medication adherence",
			zap.Error(err),
			zap.String("medication_id", medicationID),
		)
		return fmt.Errorf("failed to log adherence: %w", err)
	}

	s.logger.Info("medication adherence logged",
		zap.String("medication_id", medicationID),
		zap.Bool("adherence", adherence),
	)

	return nil
}
