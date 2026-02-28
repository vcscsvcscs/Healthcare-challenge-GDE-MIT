package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/pkg/api"
	"go.uber.org/zap"
)

// Property 24: Error Response Structure
// Feature: eva-health-backend, Property 24: Error Response Structure
// **Validates: Requirements 11.1, 11.2, 11.3, 11.5, 11.6**
func TestProperty_ErrorResponseStructure(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	logger := zap.NewNop()

	// Test various error scenarios that trigger validation errors at JSON binding level
	properties.Property("All error responses follow standard structure with code, message, and optional details", prop.ForAll(
		func(errorScenario string) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			var expectedCode string
			var expectedStatus int

			switch errorScenario {
			case "invalid_json_checkin":
				// Test invalid JSON in check-in start
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString("{invalid json"))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

				expectedCode = "VALIDATION_ERROR"
				expectedStatus = http.StatusBadRequest

			case "invalid_json_medication":
				// Test invalid JSON in medication creation
				handler := &MedicationHandler{logger: logger}
				router.POST("/test", handler.PostApiV1HealthMedications)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"name": "test", "dosage": }`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

				expectedCode = "VALIDATION_ERROR"
				expectedStatus = http.StatusBadRequest

			case "empty_response_checkin":
				// Test empty response validation
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinRespond)

				sessionID := uuid.New()
				reqBody := fmt.Sprintf(`{"session_id":"%s","response":""}`, sessionID.String())
				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

				expectedCode = "VALIDATION_ERROR"
				expectedStatus = http.StatusBadRequest

			case "invalid_uuid_format":
				// Test invalid UUID format
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"user_id":"not-a-uuid"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

				expectedCode = "VALIDATION_ERROR"
				expectedStatus = http.StatusBadRequest

			case "malformed_json_array":
				// Test malformed JSON array
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`[1,2,3`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

				expectedCode = "VALIDATION_ERROR"
				expectedStatus = http.StatusBadRequest

			default:
				return true
			}

			// Verify status code
			if w.Code != expectedStatus {
				t.Logf("Scenario %s: Expected status code %d, got %d", errorScenario, expectedStatus, w.Code)
				return false
			}

			// Parse response body
			var errorResp api.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
				t.Logf("Scenario %s: Failed to parse error response: %v, body: %s", errorScenario, err, w.Body.String())
				return false
			}

			// Verify required fields exist
			if errorResp.Code == "" {
				t.Logf("Scenario %s: Error response missing 'code' field", errorScenario)
				return false
			}

			if errorResp.Message == "" {
				t.Logf("Scenario %s: Error response missing 'message' field", errorScenario)
				return false
			}

			// Verify code matches expected
			if errorResp.Code != expectedCode {
				t.Logf("Scenario %s: Expected error code '%s', got '%s'", errorScenario, expectedCode, errorResp.Code)
				return false
			}

			// Verify response is valid JSON with correct structure
			// Details field is optional, but if present should be a string pointer
			// (already validated by JSON unmarshaling)

			return true
		},
		gen.OneConstOf(
			"invalid_json_checkin",
			"invalid_json_medication",
			"empty_response_checkin",
			"invalid_uuid_format",
			"malformed_json_array",
		),
	))

	properties.TestingRun(t)
}

// Property 25: Request Validation Completeness
// Feature: eva-health-backend, Property 25: Request Validation Completeness
// **Validates: Requirements 11.4**
func TestProperty_RequestValidationCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	logger := zap.NewNop()

	// Test validation across different endpoints with various invalid inputs
	// Focus on JSON binding errors that don't require service calls
	properties.Property("Request validation catches all invalid inputs and returns appropriate error responses", prop.ForAll(
		func(validationType string, invalidValue int) bool {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, router := gin.CreateTestContext(w)

			switch validationType {
			case "invalid_json_structure":
				// Test malformed JSON
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{invalid json`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			case "invalid_uuid_type":
				// Test wrong data type (string instead of UUID)
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"user_id":"not-a-uuid"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			case "invalid_date_format":
				// Test invalid date format
				handler := &MedicationHandler{logger: logger}
				router.POST("/test", handler.PostApiV1HealthMedications)

				userID := uuid.New()
				reqBody := fmt.Sprintf(`{"user_id":"%s","name":"Test","dosage":"10mg","frequency":"daily","start_date":"not-a-date"}`, userID.String())
				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(reqBody))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			case "incomplete_json_object":
				// Test incomplete JSON object
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"user_id":`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			case "wrong_json_type":
				// Test wrong JSON type (array instead of object)
				handler := &CheckInHandler{logger: logger}
				router.POST("/test", handler.PostApiV1CheckinStart)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`[]`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			case "malformed_json_quotes":
				// Test malformed JSON with unescaped quotes
				handler := &MedicationHandler{logger: logger}
				router.POST("/test", handler.PostApiV1HealthMedications)

				c.Request = httptest.NewRequest("POST", "/test", bytes.NewBufferString(`{"name": "test"value"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				router.ServeHTTP(w, c.Request)

			default:
				return true
			}

			// Verify that a 400 Bad Request was returned
			if w.Code != http.StatusBadRequest {
				t.Logf("Validation type %s: Expected status 400 for validation error, got %d", validationType, w.Code)
				return false
			}

			// Parse error response
			var errorResp api.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &errorResp); err != nil {
				t.Logf("Validation type %s: Failed to parse error response: %v, body: %s", validationType, err, w.Body.String())
				return false
			}

			// Verify error code is VALIDATION_ERROR
			if errorResp.Code != "VALIDATION_ERROR" {
				t.Logf("Validation type %s: Expected error code 'VALIDATION_ERROR', got '%s'", validationType, errorResp.Code)
				return false
			}

			// Verify message is present and descriptive
			if errorResp.Message == "" {
				t.Logf("Validation type %s: Error message is empty", validationType)
				return false
			}

			// Verify the response structure is consistent
			if errorResp.Code == "" || errorResp.Message == "" {
				t.Logf("Validation type %s: Error response missing required fields", validationType)
				return false
			}

			return true
		},
		gen.OneConstOf(
			"invalid_json_structure",
			"invalid_uuid_type",
			"invalid_date_format",
			"incomplete_json_object",
			"wrong_json_type",
			"malformed_json_quotes",
		),
		gen.IntRange(0, 100), // Dummy parameter for variety
	))

	properties.TestingRun(t)
}
