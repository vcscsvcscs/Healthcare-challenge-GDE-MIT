package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/handler"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/repository"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/service"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// TestCheckInFlowIntegration tests the complete check-in flow from start to completion
// Requirements: 1.1-1.7, 2.1-2.6, 3.1-3.12
func TestCheckInFlowIntegration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test environment
	ctx := context.Background()
	logger := zap.NewNop()

	// Initialize database connection
	db, cleanup := setupTestDatabase(t, ctx)
	defer cleanup()

	// Initialize Azure clients
	azureClients := setupAzureClients(t, logger)

	// Initialize repositories
	checkInRepo := repository.NewCheckInRepository(db, logger)

	// Initialize services
	checkInService := service.NewCheckInService(
		checkInRepo,
		azureClients.OpenAI,
		azureClients.Speech,
		azureClients.Blob,
		logger,
	)

	// Initialize handlers
	checkInHandler := handler.NewCheckInHandler(checkInService, logger)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	registerCheckInRoutes(router, checkInHandler)

	// Test user ID
	userID := uuid.New()

	t.Run("Complete check-in flow", func(t *testing.T) {
		// Step 1: Start a new check-in session
		t.Log("Step 1: Starting check-in session")
		sessionID, firstQuestion := startCheckInSession(t, router, userID)
		require.NotEmpty(t, sessionID, "Session ID should not be empty")
		require.NotEmpty(t, firstQuestion, "First question should not be empty")
		assert.Contains(t, firstQuestion, "Szia", "First question should be in Hungarian")

		// Step 2: Answer all questions in the conversation flow
		t.Log("Step 2: Answering questions")
		responses := []string{
			"Jól érzem magam ma, kicsit fáradt vagyok.",
			"Igen, reggel futottam 5 kilométert.",
			"Reggelire zabkását ettem, ebédre csirkét rizzsel, vacsorára salátát.",
			"Igen, kicsit fáj a fejem.",
			"Jól aludtam, 8 órát.",
			"Közepes az energiaszintem.",
			"Igen, beszedtem minden gyógyszeremet.",
			"Semmi különös, minden rendben.",
		}

		var isComplete bool
		for i, response := range responses {
			t.Logf("  Answering question %d: %s", i+1, response)
			isComplete = answerQuestion(t, router, sessionID, response)

			// Only check completion status after the last question
			if i == len(responses)-1 {
				assert.True(t, isComplete, "Session should be complete after all questions")
			}
		}

		// Step 3: Complete the session and extract data
		t.Log("Step 3: Completing session and extracting data")
		healthCheckIn := completeCheckInSession(t, router, sessionID)
		require.NotNil(t, healthCheckIn, "Health check-in should not be nil")

		// Verify extracted data structure
		t.Log("Step 4: Verifying extracted data")
		verifyExtractedData(t, healthCheckIn)

		// Step 5: Verify data is stored in database
		t.Log("Step 5: Verifying data persistence")
		verifyDataPersistence(t, ctx, checkInRepo, sessionID, userID.String())
	})

	t.Run("Audio streaming and transcription", func(t *testing.T) {
		// Start a new session
		sessionID, _ := startCheckInSession(t, router, userID)

		// Test audio streaming
		t.Log("Testing audio streaming")
		testAudioStreaming(t, router, sessionID, azureClients.Speech)
	})

	t.Run("Session timeout handling", func(t *testing.T) {
		// This test would require manipulating time or waiting 30 minutes
		// For practical purposes, we test the timeout logic with a mock
		t.Skip("Timeout test requires time manipulation - covered in unit tests")
	})
}

// startCheckInSession starts a new check-in session and returns the session ID and first question
func startCheckInSession(t *testing.T, router *gin.Engine, userID uuid.UUID) (string, string) {
	reqBody := api.StartSessionRequest{
		UserId: userID,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/checkin/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response body: %s", w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code, "Start session should return 200 OK")

	var response api.SessionResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")

	require.NotNil(t, response.SessionId, "Session ID should not be nil")
	require.NotNil(t, response.QuestionText, "Question text should not be nil")

	return response.SessionId.String(), *response.QuestionText
}

// answerQuestion submits a response to a question and returns whether the session is complete
func answerQuestion(t *testing.T, router *gin.Engine, sessionID string, response string) bool {
	sessionUUID, err := uuid.Parse(sessionID)
	require.NoError(t, err)

	reqBody := api.RespondRequest{
		SessionId: sessionUUID,
		Response:  response,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/checkin/respond", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Logf("Response status: %d, body: %s", w.Code, w.Body.String())
	}
	assert.Equal(t, http.StatusOK, w.Code, "Respond should return 200 OK")

	var respData api.ConversationStateResponse
	err = json.Unmarshal(w.Body.Bytes(), &respData)
	require.NoError(t, err, "Should be able to parse response")

	require.NotNil(t, respData.IsComplete, "IsComplete should not be nil")
	return *respData.IsComplete
}

// completeCheckInSession completes a check-in session and returns the health check-in data
func completeCheckInSession(t *testing.T, router *gin.Engine, sessionID string) *api.HealthCheckInResponse {
	sessionUUID, err := uuid.Parse(sessionID)
	require.NoError(t, err)

	reqBody := api.CompleteSessionRequest{
		SessionId: sessionUUID,
	}
	body, err := json.Marshal(reqBody)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/checkin/complete", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Complete session should return 200 OK")

	var response api.HealthCheckInResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")

	return &response
}

// verifyExtractedData verifies that the extracted health data has the correct structure
func verifyExtractedData(t *testing.T, checkIn *api.HealthCheckInResponse) {
	// Verify required fields are present
	assert.NotNil(t, checkIn.Id, "Check-in ID should not be nil")
	assert.NotNil(t, checkIn.UserId, "User ID should not be nil")
	assert.NotNil(t, checkIn.CheckInDate, "Check-in date should not be nil")

	// Verify mood is valid enum value
	if checkIn.Mood != nil {
		validMoods := []api.HealthCheckInResponseMood{
			api.Positive,
			api.Neutral,
			api.Negative,
		}
		assert.Contains(t, validMoods, *checkIn.Mood, "Mood should be a valid enum value")
	}

	// Verify energy level is valid enum value
	if checkIn.EnergyLevel != nil {
		validEnergyLevels := []api.HealthCheckInResponseEnergyLevel{
			api.Low,
			api.Medium,
			api.High,
		}
		assert.Contains(t, validEnergyLevels, *checkIn.EnergyLevel, "Energy level should be a valid enum value")
	}

	// Verify sleep quality is valid enum value
	if checkIn.SleepQuality != nil {
		validSleepQualities := []api.HealthCheckInResponseSleepQuality{
			api.Poor,
			api.Fair,
			api.Good,
			api.Excellent,
		}
		assert.Contains(t, validSleepQualities, *checkIn.SleepQuality, "Sleep quality should be a valid enum value")
	}

	// Verify medication taken is valid enum value
	if checkIn.MedicationTaken != nil {
		validMedicationTaken := []api.HealthCheckInResponseMedicationTaken{
			api.Yes,
			api.No,
			api.Partial,
		}
		assert.Contains(t, validMedicationTaken, *checkIn.MedicationTaken, "Medication taken should be a valid enum value")
	}

	// Verify pain level is in valid range if present
	if checkIn.PainLevel != nil {
		assert.GreaterOrEqual(t, *checkIn.PainLevel, 0, "Pain level should be >= 0")
		assert.LessOrEqual(t, *checkIn.PainLevel, 10, "Pain level should be <= 10")
	}

	// Verify arrays are initialized (not nil)
	assert.NotNil(t, checkIn.Symptoms, "Symptoms should not be nil")
	assert.NotNil(t, checkIn.PhysicalActivity, "Physical activity should not be nil")

	t.Logf("Extracted data verification passed:")
	t.Logf("  Mood: %v", checkIn.Mood)
	t.Logf("  Energy Level: %v", checkIn.EnergyLevel)
	t.Logf("  Sleep Quality: %v", checkIn.SleepQuality)
	t.Logf("  Pain Level: %v", checkIn.PainLevel)
	t.Logf("  Symptoms: %v", checkIn.Symptoms)
	t.Logf("  Physical Activity: %v", checkIn.PhysicalActivity)
}

// verifyDataPersistence verifies that data is correctly stored in the database
func verifyDataPersistence(t *testing.T, ctx context.Context, repo *repository.CheckInRepository, sessionID, userID string) {
	// Verify session exists and is completed
	session, err := repo.GetSession(ctx, sessionID)
	require.NoError(t, err, "Should be able to retrieve session")
	assert.Equal(t, "completed", string(session.Status), "Session should be completed")
	assert.NotNil(t, session.CompletedAt, "Session should have completion time")

	// Verify conversation messages are stored
	messages, err := repo.GetConversationMessages(ctx, sessionID)
	require.NoError(t, err, "Should be able to retrieve messages")
	assert.Greater(t, len(messages), 0, "Should have conversation messages")

	// Count assistant and user messages
	assistantCount := 0
	userCount := 0
	for _, msg := range messages {
		if msg.Role == "assistant" {
			assistantCount++
		} else if msg.Role == "user" {
			userCount++
		}
	}
	// We expect 7 assistant messages (questions 1-7) and 8 user messages (responses to all 8 questions)
	// The 8th question triggers completion, so no assistant message is saved for it
	assert.GreaterOrEqual(t, assistantCount, 7, "Should have at least 7 assistant messages (questions)")
	assert.Equal(t, 8, userCount, "Should have 8 user messages (responses)")

	// Verify health check-in is stored
	checkIns, err := repo.GetHealthCheckInsByUserID(ctx, userID)
	require.NoError(t, err, "Should be able to retrieve health check-ins")
	assert.Greater(t, len(checkIns), 0, "Should have at least one health check-in")

	// Find the check-in for this session
	var sessionCheckIn *interface{}
	for i := range checkIns {
		if checkIns[i].SessionID != nil && *checkIns[i].SessionID == sessionID {
			sessionCheckIn = new(interface{})
			break
		}
	}
	assert.NotNil(t, sessionCheckIn, "Should find health check-in for this session")

	t.Log("Data persistence verification passed")
}

// testAudioStreaming tests audio streaming and transcription
func testAudioStreaming(t *testing.T, router *gin.Engine, sessionID string, speechClient *azure.SpeechServiceClient) {
	// Generate test audio using Text-to-Speech
	testText := "Ez egy teszt válasz."
	audioData, err := speechClient.TextToSpeechWAV(context.Background(), testText, "hu-HU")
	require.NoError(t, err, "Should be able to generate test audio")
	require.Greater(t, len(audioData), 0, "Audio data should not be empty")

	// Stream audio to the API
	sessionUUID, err := uuid.Parse(sessionID)
	require.NoError(t, err)

	url := fmt.Sprintf("/api/v1/checkin/audio-stream?session_id=%s", sessionUUID.String())
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewReader(audioData))
	req.Header.Set("Content-Type", "audio/wav")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "Audio stream should return 200 OK")

	var response map[string]string
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "Should be able to parse response")

	transcription, ok := response["transcription"]
	require.True(t, ok, "Response should contain transcription")
	assert.NotEmpty(t, transcription, "Transcription should not be empty")

	t.Logf("Audio streaming test passed. Transcription: %s", transcription)
}

// setupTestDatabase initializes a test database connection
func setupTestDatabase(t *testing.T, ctx context.Context) (*pgxpool.Pool, func()) {
	// Get database URL from environment or use default
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		// Default to local PostgreSQL for testing
		dbURL = "postgres://postgres:postgres@localhost:5432/eva_health_test?sslmode=disable"
	}

	t.Logf("Connecting to database: %s", dbURL)

	// Connect to database
	config, err := pgxpool.ParseConfig(dbURL)
	require.NoError(t, err, "Should be able to parse database URL")

	db, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err, "Should be able to connect to database")

	// Verify connection
	err = db.Ping(ctx)
	require.NoError(t, err, "Should be able to ping database")

	// Verify tables exist
	var tableExists bool
	err = db.QueryRow(ctx, "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'check_in_sessions')").Scan(&tableExists)
	require.NoError(t, err, "Should be able to check if tables exist")

	if !tableExists {
		t.Fatal("Database tables do not exist. Please run migrations first with: mise run migrate-up")
	}

	t.Log("Database connection established and tables verified")

	// Cleanup function
	cleanup := func() {
		db.Close()
		t.Log("Database connection closed")
	}

	return db, cleanup
}

// AzureClients holds all Azure service clients
type AzureClients struct {
	OpenAI *azure.OpenAIClient
	Speech *azure.SpeechServiceClient
	Blob   *azure.BlobStorageClient
	// For testing, we might use a mock that wraps the real client
	BlobMock *azure.MockBlobStorageClient
}

// setupAzureClients initializes Azure service clients
func setupAzureClients(t *testing.T, logger *zap.Logger) *AzureClients {
	// Check if we should use real Azure services or mocks
	useRealAzure := os.Getenv("USE_REAL_AZURE") == "true"

	if !useRealAzure {
		t.Log("Using mock Azure clients (set USE_REAL_AZURE=true to use real services)")
		return setupMockAzureClients(t, logger)
	}

	t.Log("Using real Azure clients")

	// Get Azure credentials from environment
	openAIEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	openAIKey := os.Getenv("AZURE_OPENAI_KEY")
	openAIDeployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

	speechKey := os.Getenv("AZURE_SPEECH_KEY")
	speechRegion := os.Getenv("AZURE_SPEECH_REGION")

	storageAccountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	storageAccountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	storageContainer := os.Getenv("AZURE_STORAGE_CONTAINER")

	// Validate required environment variables
	require.NotEmpty(t, openAIEndpoint, "AZURE_OPENAI_ENDPOINT is required")
	require.NotEmpty(t, openAIKey, "AZURE_OPENAI_KEY is required")
	require.NotEmpty(t, openAIDeployment, "AZURE_OPENAI_DEPLOYMENT is required")
	require.NotEmpty(t, speechKey, "AZURE_SPEECH_KEY is required")
	require.NotEmpty(t, speechRegion, "AZURE_SPEECH_REGION is required")
	require.NotEmpty(t, storageAccountName, "AZURE_STORAGE_ACCOUNT_NAME is required")
	require.NotEmpty(t, storageAccountKey, "AZURE_STORAGE_ACCOUNT_KEY is required")
	require.NotEmpty(t, storageContainer, "AZURE_STORAGE_CONTAINER is required")

	// Initialize clients
	openAIClient, err := azure.NewOpenAIClient(openAIEndpoint, openAIKey, openAIDeployment, logger)
	require.NoError(t, err, "Should be able to create OpenAI client")

	speechClient, err := azure.NewSpeechServiceClient(speechKey, speechRegion, logger)
	require.NoError(t, err, "Should be able to create Speech Service client")

	blobClient, err := azure.NewBlobStorageClient(storageAccountName, storageAccountKey, storageContainer, logger)
	require.NoError(t, err, "Should be able to create Blob Storage client")

	return &AzureClients{
		OpenAI: openAIClient,
		Speech: speechClient,
		Blob:   blobClient,
	}
}

// setupMockAzureClients creates mock Azure clients for testing without real Azure services
func setupMockAzureClients(t *testing.T, logger *zap.Logger) *AzureClients {
	// Create mock OpenAI client with test server
	mockOpenAIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock data extraction response
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": `{
							"symptoms": ["fejfájás", "fáradtság"],
							"mood": "neutral",
							"pain_level": 3,
							"energy_level": "medium",
							"sleep_quality": "good",
							"medication_taken": "yes",
							"physical_activity": ["futás"],
							"meals": {
								"breakfast": "zabkása",
								"lunch": "csirke rizzsel",
								"dinner": "saláta"
							},
							"general_feeling": "Jól érzem magam, kicsit fáradt vagyok",
							"additional_notes": "Semmi különös"
						}`,
					},
				},
			},
			"usage": map[string]interface{}{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(mockOpenAIServer.Close)

	openAIClient, err := azure.NewOpenAIClient(mockOpenAIServer.URL, "test-key", "test-deployment", logger)
	require.NoError(t, err)

	// Create mock Speech Service client with test server
	mockSpeechServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "speech/recognition") {
			// Mock speech-to-text response
			response := map[string]interface{}{
				"RecognitionStatus": "Success",
				"DisplayText":       "Ez egy teszt válasz",
				"Offset":            0,
				"Duration":          1000000,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
		} else {
			// Mock text-to-speech response (return dummy audio data)
			w.Header().Set("Content-Type", "audio/wav")
			w.Write([]byte("RIFF....WAVEfmt ")) // Minimal WAV header
		}
	}))
	t.Cleanup(mockSpeechServer.Close)

	speechClient, err := azure.NewSpeechServiceClient("test-key", "test-region", logger)
	require.NoError(t, err)
	// Override endpoint for testing
	speechClient.SetEndpointForTesting(mockSpeechServer.URL)

	// Create mock Blob Storage client (in-memory storage)
	// Note: For now, we create a nil BlobStorageClient since the mock doesn't match the interface
	// In production, you should refactor the service to use an interface
	// For this test, we'll skip blob operations or use a real client
	blobClient, _ := azure.NewBlobStorageClient("test", "dGVzdA==", "test-container", logger)

	return &AzureClients{
		OpenAI:   openAIClient,
		Speech:   speechClient,
		Blob:     blobClient,
		BlobMock: azure.NewMockBlobStorageClient(logger),
	}
}

// registerCheckInRoutes registers check-in routes on the router
func registerCheckInRoutes(router *gin.Engine, handler *handler.CheckInHandler) {
	v1 := router.Group("/api/v1")
	{
		checkin := v1.Group("/checkin")
		{
			checkin.POST("/start", handler.PostApiV1CheckinStart)
			checkin.POST("/audio-stream", func(c *gin.Context) {
				sessionIDStr := c.Query("session_id")
				sessionID, err := uuid.Parse(sessionIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
					return
				}
				handler.PostApiV1CheckinAudioStream(c, api.PostApiV1CheckinAudioStreamParams{
					SessionId: sessionID,
				})
			})
			checkin.POST("/respond", handler.PostApiV1CheckinRespond)
			checkin.GET("/status/:sessionId", func(c *gin.Context) {
				sessionIDStr := c.Param("sessionId")
				sessionID, err := uuid.Parse(sessionIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
					return
				}
				handler.GetApiV1CheckinStatusSessionId(c, sessionID)
			})
			checkin.GET("/question-audio/:sessionId/:questionId", func(c *gin.Context) {
				sessionIDStr := c.Param("sessionId")
				sessionID, err := uuid.Parse(sessionIDStr)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
					return
				}
				questionID := c.Param("questionId")
				handler.GetApiV1CheckinQuestionAudioSessionIdQuestionId(c, sessionID, questionID)
			})
			checkin.POST("/complete", handler.PostApiV1CheckinComplete)
		}
	}
}
