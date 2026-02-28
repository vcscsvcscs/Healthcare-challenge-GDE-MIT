package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// CheckInHandler implements check-in API endpoints
type CheckInHandler struct {
	service *service.CheckInService
	logger  *zap.Logger
}

// NewCheckInHandler creates a new CheckInHandler
func NewCheckInHandler(service *service.CheckInService, logger *zap.Logger) *CheckInHandler {
	return &CheckInHandler{
		service: service,
		logger:  logger,
	}
}

// PostApiV1CheckinStart starts a new check-in session
func (h *CheckInHandler) PostApiV1CheckinStart(c *gin.Context) {
	var req api.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	userID := uuidToString(req.UserId)

	// Start session
	sessionWithAudio, err := h.service.StartSession(c.Request.Context(), userID)
	if err != nil {
		h.logger.Error("failed to start session",
			zap.Error(err),
			zap.String("user_id", userID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to start check-in session",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	status := api.SessionResponseStatus(sessionWithAudio.Session.Status)
	response := api.SessionResponse{
		SessionId:    stringToUUID(sessionWithAudio.Session.ID),
		QuestionText: stringPtr(sessionWithAudio.QuestionText),
		QuestionId:   stringPtr(sessionWithAudio.QuestionID),
		Status:       &status,
		UserId:       stringToUUID(userID),
		StartedAt:    timePtr(sessionWithAudio.Session.StartedAt),
	}

	h.logger.Info("check-in session started",
		zap.String("session_id", sessionWithAudio.Session.ID),
		zap.String("user_id", userID),
	)

	c.JSON(http.StatusOK, response)
}

// PostApiV1CheckinAudioStream handles audio streaming for real-time transcription
func (h *CheckInHandler) PostApiV1CheckinAudioStream(c *gin.Context, params api.PostApiV1CheckinAudioStreamParams) {
	sessionID := params.SessionId.String()

	h.logger.Info("audio stream started",
		zap.String("session_id", sessionID),
	)

	// Read audio stream from request body
	audioStream := c.Request.Body
	defer c.Request.Body.Close()

	// Stream audio to speech service for transcription
	transcription, err := h.service.StreamAudioToSpeech(c.Request.Context(), sessionID, audioStream)
	if err != nil {
		h.logger.Error("audio streaming failed",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to transcribe audio",
			Details: stringPtr(err.Error()),
		})
		return
	}

	h.logger.Info("audio transcribed successfully",
		zap.String("session_id", sessionID),
		zap.Int("transcription_length", len(transcription)),
	)

	c.JSON(http.StatusOK, gin.H{
		"transcription": transcription,
	})
}

// PostApiV1CheckinRespond processes user response and returns next question
func (h *CheckInHandler) PostApiV1CheckinRespond(c *gin.Context) {
	var req api.RespondRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	sessionID := uuidToString(req.SessionId)

	// Validate request
	if req.Response == "" {
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Response is required",
		})
		return
	}

	// Process response
	conversationState, err := h.service.ProcessResponse(c.Request.Context(), sessionID, req.Response)
	if err != nil {
		h.logger.Error("failed to process response",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process response",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.ConversationStateResponse{
		SessionId:    stringToUUID(conversationState.SessionID),
		QuestionText: stringPtr(conversationState.QuestionText),
		QuestionId:   stringPtr(conversationState.QuestionID),
		IsComplete:   boolPtr(conversationState.IsComplete),
	}

	h.logger.Info("response processed",
		zap.String("session_id", sessionID),
		zap.Bool("is_complete", conversationState.IsComplete),
	)

	c.JSON(http.StatusOK, response)
}

// GetApiV1CheckinStatusSessionId retrieves session status
func (h *CheckInHandler) GetApiV1CheckinStatusSessionId(c *gin.Context, sessionId uuid.UUID) {
	sessionIDStr := sessionId.String()

	h.logger.Info("getting session status",
		zap.String("session_id", sessionIDStr),
	)

	// Get session status
	status, err := h.service.GetSessionStatus(c.Request.Context(), sessionIDStr)
	if err != nil {
		h.logger.Error("failed to get session status",
			zap.Error(err),
			zap.String("session_id", sessionIDStr),
		)
		c.JSON(http.StatusNotFound, api.ErrorResponse{
			Code:    "NOT_FOUND",
			Message: "Session not found",
		})
		return
	}

	// Convert to API response
	statusEnum := api.SessionStatusStatus(status.Status)
	response := api.SessionStatus{
		SessionId:         stringToUUID(status.SessionID),
		Status:            &statusEnum,
		QuestionsAnswered: intPtr(status.CurrentQuestion),
		TotalQuestions:    intPtr(status.TotalQuestions),
		StartedAt:         timePtr(status.StartedAt),
		CompletedAt:       status.CompletedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetApiV1CheckinQuestionAudioSessionIdQuestionId retrieves question audio
func (h *CheckInHandler) GetApiV1CheckinQuestionAudioSessionIdQuestionId(c *gin.Context, sessionId uuid.UUID, questionId string) {
	sessionIDStr := sessionId.String()

	h.logger.Info("getting question audio",
		zap.String("session_id", sessionIDStr),
		zap.String("question_id", questionId),
	)

	// Get question audio
	audioData, err := h.service.GetQuestionAudio(c.Request.Context(), sessionIDStr, questionId)
	if err != nil {
		h.logger.Error("failed to get question audio",
			zap.Error(err),
			zap.String("session_id", sessionIDStr),
			zap.String("question_id", questionId),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "AUDIO_GENERATION_FAILED",
			Message: "Failed to generate question audio",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Return audio as WAV
	c.Header("Content-Type", "audio/wav")
	c.Header("Content-Length", fmt.Sprintf("%d", len(audioData)))
	c.Data(http.StatusOK, "audio/wav", audioData)
}

// PostApiV1CheckinComplete completes a check-in session
func (h *CheckInHandler) PostApiV1CheckinComplete(c *gin.Context) {
	var req api.CompleteSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, api.ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request body",
			Details: stringPtr(err.Error()),
		})
		return
	}

	sessionID := uuidToString(req.SessionId)

	// Complete session
	healthCheckIn, err := h.service.CompleteSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Error("failed to complete session",
			zap.Error(err),
			zap.String("session_id", sessionID),
		)
		c.JSON(http.StatusInternalServerError, api.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to complete check-in session",
			Details: stringPtr(err.Error()),
		})
		return
	}

	// Convert to API response
	response := api.HealthCheckInResponse{
		Id:               stringToUUID(healthCheckIn.ID),
		UserId:           stringToUUID(healthCheckIn.UserID),
		CheckInDate:      timeToDate(healthCheckIn.CheckInDate),
		Symptoms:         &healthCheckIn.Symptoms,
		Mood:             (*api.HealthCheckInResponseMood)(healthCheckIn.Mood),
		PainLevel:        healthCheckIn.PainLevel,
		EnergyLevel:      (*api.HealthCheckInResponseEnergyLevel)(healthCheckIn.EnergyLevel),
		SleepQuality:     (*api.HealthCheckInResponseSleepQuality)(healthCheckIn.SleepQuality),
		MedicationTaken:  (*api.HealthCheckInResponseMedicationTaken)(healthCheckIn.MedicationTaken),
		PhysicalActivity: &healthCheckIn.PhysicalActivity,
		GeneralFeeling:   healthCheckIn.GeneralFeeling,
		AdditionalNotes:  healthCheckIn.AdditionalNotes,
		CreatedAt:        timePtr(healthCheckIn.CreatedAt),
	}

	// Add meals as nested struct
	if healthCheckIn.Breakfast != nil || healthCheckIn.Lunch != nil || healthCheckIn.Dinner != nil {
		response.Meals = &struct {
			Breakfast *string `json:"breakfast,omitempty"`
			Dinner    *string `json:"dinner,omitempty"`
			Lunch     *string `json:"lunch,omitempty"`
		}{
			Breakfast: healthCheckIn.Breakfast,
			Lunch:     healthCheckIn.Lunch,
			Dinner:    healthCheckIn.Dinner,
		}
	}

	h.logger.Info("check-in session completed",
		zap.String("session_id", sessionID),
		zap.String("check_in_id", healthCheckIn.ID),
	)

	c.JSON(http.StatusOK, response)
}
