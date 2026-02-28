package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/model"
	"go.uber.org/zap"
)

// CheckInService manages conversation flow and data extraction
type CheckInService struct {
	repo           *repository.CheckInRepository
	aiClient       *azure.OpenAIClient
	speechClient   *azure.SpeechServiceClient
	blobClient     *azure.BlobStorageClient
	dataExtractor  *DataExtractor
	logger         *zap.Logger
	sessionTimeout time.Duration
}

// NewCheckInService creates a new CheckInService
func NewCheckInService(
	repo *repository.CheckInRepository,
	aiClient *azure.OpenAIClient,
	speechClient *azure.SpeechServiceClient,
	blobClient *azure.BlobStorageClient,
	logger *zap.Logger,
) *CheckInService {
	return &CheckInService{
		repo:           repo,
		aiClient:       aiClient,
		speechClient:   speechClient,
		blobClient:     blobClient,
		dataExtractor:  NewDataExtractor(aiClient, logger),
		logger:         logger,
		sessionTimeout: 30 * time.Minute,
	}
}

// SessionWithAudio represents a session with audio for the first question
type SessionWithAudio struct {
	Session       *model.Session
	QuestionText  string
	QuestionAudio []byte
	QuestionID    string
}

// ConversationStateWithAudio represents the conversation state with audio
type ConversationStateWithAudio struct {
	SessionID     string
	QuestionText  string
	QuestionAudio []byte
	QuestionID    string
	IsComplete    bool
}

// SessionStatus represents the status of a session
type SessionStatus struct {
	SessionID       string
	Status          model.SessionStatus
	CurrentQuestion int
	TotalQuestions  int
	StartedAt       time.Time
	CompletedAt     *time.Time
	ExpiredAt       *time.Time
	MessageCount    int
}

// StartSession creates a new check-in session and returns the first question with audio
func (s *CheckInService) StartSession(ctx context.Context, userID string) (*SessionWithAudio, error) {
	s.logger.Info("starting new check-in session", zap.String("user_id", userID))

	// Create new session
	session := &model.Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		StartedAt: time.Now(),
		Status:    model.SessionStatusActive,
	}

	// Save session to database
	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Get first question
	questionFlow := NewQuestionFlow()
	firstQuestion := questionFlow.GetNextQuestion()
	if firstQuestion == nil {
		return nil, fmt.Errorf("no questions available")
	}

	// Save first question as assistant message
	assistantMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: session.ID,
		Role:      model.MessageRoleAssistant,
		Content:   firstQuestion.TextHU,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveConversationMessage(ctx, assistantMsg); err != nil {
		s.logger.Warn("failed to save assistant message", zap.Error(err))
	}

	// Generate audio for first question
	audioData, err := s.GetQuestionAudio(ctx, session.ID, firstQuestion.ID)
	if err != nil {
		s.logger.Warn("failed to generate question audio", zap.Error(err))
		// Continue without audio
		audioData = nil
	}

	s.logger.Info("check-in session started successfully",
		zap.String("session_id", session.ID),
		zap.String("question_id", firstQuestion.ID),
	)

	return &SessionWithAudio{
		Session:       session,
		QuestionText:  firstQuestion.TextHU,
		QuestionAudio: audioData,
		QuestionID:    firstQuestion.ID,
	}, nil
}

// StreamAudioToSpeech performs real-time transcription of audio stream
func (s *CheckInService) StreamAudioToSpeech(ctx context.Context, sessionID string, audioStream io.Reader) (string, error) {
	s.logger.Info("starting audio transcription", zap.String("session_id", sessionID))

	// Verify session exists and is active
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	if session.Status != model.SessionStatusActive {
		return "", fmt.Errorf("session is not active: %s", session.Status)
	}

	// Stream audio to Azure Speech Service for transcription
	transcription, err := s.speechClient.StreamAudioToText(ctx, audioStream)
	if err != nil {
		s.logger.Error("speech-to-text failed", zap.String("session_id", sessionID), zap.Error(err))
		return "", fmt.Errorf("transcription failed: %w", err)
	}

	s.logger.Info("audio transcription completed",
		zap.String("session_id", sessionID),
		zap.Int("transcription_length", len(transcription)),
	)

	return transcription, nil
}

// ProcessResponse processes a user response and returns the next question
func (s *CheckInService) ProcessResponse(ctx context.Context, sessionID string, response string) (*ConversationStateWithAudio, error) {
	s.logger.Info("processing user response",
		zap.String("session_id", sessionID),
		zap.Int("response_length", len(response)),
	)

	// Verify session exists and is active
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.Status != model.SessionStatusActive {
		return nil, fmt.Errorf("session is not active: %s", session.Status)
	}

	// Check for session timeout
	if time.Since(session.StartedAt) > s.sessionTimeout {
		s.logger.Warn("session timeout", zap.String("session_id", sessionID))
		session.Status = model.SessionStatusExpired
		now := time.Now()
		session.ExpiredAt = &now
		if err := s.repo.UpdateSession(ctx, session); err != nil {
			s.logger.Error("failed to update expired session", zap.Error(err))
		}
		return nil, fmt.Errorf("session has expired")
	}

	// Validate response is not empty
	if response == "" {
		return nil, fmt.Errorf("response cannot be empty")
	}

	// Save user response
	userMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      model.MessageRoleUser,
		Content:   response,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveConversationMessage(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("failed to save user message: %w", err)
	}

	// Get conversation history to determine current question
	messages, err := s.repo.GetConversationMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Count how many questions have been asked (assistant messages)
	questionCount := 0
	for _, msg := range messages {
		if msg.Role == model.MessageRoleAssistant {
			questionCount++
		}
	}

	// Get next question
	questionFlow := NewQuestionFlow()
	// Advance to current position
	for i := 0; i < questionCount; i++ {
		questionFlow.GetNextQuestion()
	}

	nextQuestion := questionFlow.GetNextQuestion()
	if nextQuestion == nil || questionFlow.IsComplete() {
		// All questions answered
		s.logger.Info("all questions answered", zap.String("session_id", sessionID))
		return &ConversationStateWithAudio{
			SessionID:  sessionID,
			IsComplete: true,
		}, nil
	}

	// Save next question as assistant message
	assistantMsg := &model.Message{
		ID:        uuid.New().String(),
		SessionID: sessionID,
		Role:      model.MessageRoleAssistant,
		Content:   nextQuestion.TextHU,
		CreatedAt: time.Now(),
	}
	if err := s.repo.SaveConversationMessage(ctx, assistantMsg); err != nil {
		s.logger.Warn("failed to save assistant message", zap.Error(err))
	}

	// Generate audio for next question
	audioData, err := s.GetQuestionAudio(ctx, sessionID, nextQuestion.ID)
	if err != nil {
		s.logger.Warn("failed to generate question audio", zap.Error(err))
		audioData = nil
	}

	s.logger.Info("response processed successfully",
		zap.String("session_id", sessionID),
		zap.String("next_question_id", nextQuestion.ID),
	)

	return &ConversationStateWithAudio{
		SessionID:     sessionID,
		QuestionText:  nextQuestion.TextHU,
		QuestionAudio: audioData,
		QuestionID:    nextQuestion.ID,
		IsComplete:    false,
	}, nil
}

// GetQuestionAudio generates or retrieves cached audio for a question
func (s *CheckInService) GetQuestionAudio(ctx context.Context, sessionID string, questionID string) ([]byte, error) {
	s.logger.Info("getting question audio",
		zap.String("session_id", sessionID),
		zap.String("question_id", questionID),
	)

	// Get question text
	questionFlow := NewQuestionFlow()
	question := questionFlow.GetQuestionByID(questionID)
	if question == nil {
		return nil, fmt.Errorf("question not found: %s", questionID)
	}

	// Check if audio is cached in blob storage
	cacheKey := fmt.Sprintf("question-audio/hu-HU/%s.mp3", questionID)
	audioData, err := s.blobClient.DownloadAudio(ctx, cacheKey)
	if err == nil {
		s.logger.Info("question audio retrieved from cache",
			zap.String("question_id", questionID),
			zap.Int("audio_size", len(audioData)),
		)
		return audioData, nil
	}

	// Generate audio using Text-to-Speech
	s.logger.Info("generating question audio", zap.String("question_id", questionID))
	audioData, err = s.speechClient.TextToSpeech(ctx, question.TextHU, "hu-HU")
	if err != nil {
		return nil, fmt.Errorf("TTS failed: %w", err)
	}

	// Cache audio for future use (async)
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if _, err := s.blobClient.UploadAudio(cacheCtx, cacheKey, bytes.NewReader(audioData)); err != nil {
			s.logger.Error("failed to cache question audio",
				zap.String("question_id", questionID),
				zap.Error(err),
			)
		} else {
			s.logger.Info("question audio cached successfully", zap.String("question_id", questionID))
		}
	}()

	return audioData, nil
}

// CompleteSession completes a check-in session and extracts health data
func (s *CheckInService) CompleteSession(ctx context.Context, sessionID string) (*model.HealthCheckIn, error) {
	s.logger.Info("completing check-in session", zap.String("session_id", sessionID))

	// Get session
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	if session.Status != model.SessionStatusActive {
		return nil, fmt.Errorf("session is not active: %s", session.Status)
	}

	// Get conversation history
	messages, err := s.repo.GetConversationMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Build conversation history for extraction
	var conversationHistory []ConversationMessage
	for _, msg := range messages {
		conversationHistory = append(conversationHistory, ConversationMessage{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	// Extract structured data using AI
	extractedData, err := s.dataExtractor.Extract(ctx, conversationHistory)
	if err != nil {
		s.logger.Error("data extraction failed", zap.String("session_id", sessionID), zap.Error(err))

		// Store raw transcript for manual review
		var rawTranscript string
		for _, msg := range messages {
			rawTranscript += fmt.Sprintf("%s: %s\n", msg.Role, msg.Content)
		}

		checkIn := &model.HealthCheckIn{
			ID:            uuid.New().String(),
			UserID:        session.UserID,
			SessionID:     &sessionID,
			CheckInDate:   time.Now(),
			RawTranscript: &rawTranscript,
		}

		if err := s.repo.SaveHealthCheckIn(ctx, checkIn); err != nil {
			return nil, fmt.Errorf("failed to save health check-in with raw transcript: %w", err)
		}

		return nil, fmt.Errorf("data extraction failed, raw transcript saved for manual review: %w", err)
	}

	// Create HealthCheckIn from extracted data
	checkIn := &model.HealthCheckIn{
		ID:               uuid.New().String(),
		UserID:           session.UserID,
		SessionID:        &sessionID,
		CheckInDate:      time.Now(),
		Symptoms:         extractedData.Symptoms,
		Mood:             &extractedData.Mood,
		PainLevel:        extractedData.PainLevel,
		EnergyLevel:      &extractedData.EnergyLevel,
		SleepQuality:     &extractedData.SleepQuality,
		MedicationTaken:  &extractedData.MedicationTaken,
		PhysicalActivity: extractedData.PhysicalActivity,
		Breakfast:        &extractedData.Meals.Breakfast,
		Lunch:            &extractedData.Meals.Lunch,
		Dinner:           &extractedData.Meals.Dinner,
		GeneralFeeling:   &extractedData.GeneralFeeling,
		AdditionalNotes:  &extractedData.AdditionalNotes,
	}

	// Save health check-in
	if err := s.repo.SaveHealthCheckIn(ctx, checkIn); err != nil {
		return nil, fmt.Errorf("failed to save health check-in: %w", err)
	}

	// Update session status to completed
	now := time.Now()
	session.Status = model.SessionStatusCompleted
	session.CompletedAt = &now
	if err := s.repo.UpdateSession(ctx, session); err != nil {
		s.logger.Error("failed to update session status", zap.Error(err))
	}

	// Calculate session duration and message count
	sessionDuration := now.Sub(session.StartedAt)
	messageCount := len(messages)

	// Log session completion with metrics
	// Validates: Requirements 12.4
	s.logger.Info("check-in session completed successfully",
		zap.String("session_id", sessionID),
		zap.String("check_in_id", checkIn.ID),
		zap.Duration("session_duration", sessionDuration),
		zap.Int("message_exchanges", messageCount),
		zap.Time("started_at", session.StartedAt),
		zap.Time("completed_at", now),
	)

	return checkIn, nil
}

// GetSessionStatus returns the current status of a session
func (s *CheckInService) GetSessionStatus(ctx context.Context, sessionID string) (*SessionStatus, error) {
	s.logger.Info("getting session status", zap.String("session_id", sessionID))

	// Get session
	session, err := s.repo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get conversation messages
	messages, err := s.repo.GetConversationMessages(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}

	// Count questions asked
	questionCount := 0
	for _, msg := range messages {
		if msg.Role == model.MessageRoleAssistant {
			questionCount++
		}
	}

	// Get total questions
	questionFlow := NewQuestionFlow()
	totalQuestions := questionFlow.GetTotalQuestions()

	status := &SessionStatus{
		SessionID:       sessionID,
		Status:          session.Status,
		CurrentQuestion: questionCount,
		TotalQuestions:  totalQuestions,
		StartedAt:       session.StartedAt,
		CompletedAt:     session.CompletedAt,
		ExpiredAt:       session.ExpiredAt,
		MessageCount:    len(messages),
	}

	return status, nil
}
