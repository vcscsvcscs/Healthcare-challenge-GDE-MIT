package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"go.uber.org/zap"
)

// ExtractedData represents structured health data extracted from conversation
type ExtractedData struct {
	Symptoms         []string `json:"symptoms"`
	Mood             string   `json:"mood"` // positive, neutral, negative
	PainLevel        *int     `json:"pain_level,omitempty"`
	EnergyLevel      string   `json:"energy_level"`     // low, medium, high
	SleepQuality     string   `json:"sleep_quality"`    // poor, fair, good, excellent
	MedicationTaken  string   `json:"medication_taken"` // yes, no, partial
	PhysicalActivity []string `json:"physical_activity"`
	Meals            MealInfo `json:"meals"`
	GeneralFeeling   string   `json:"general_feeling"`
	AdditionalNotes  string   `json:"additional_notes"`
}

// MealInfo represents meal information
type MealInfo struct {
	Breakfast string `json:"breakfast"`
	Lunch     string `json:"lunch"`
	Dinner    string `json:"dinner"`
}

// DataExtractor extracts structured data from conversation using Azure OpenAI
type DataExtractor struct {
	aiClient *azure.OpenAIClient
	logger   *zap.Logger
}

// NewDataExtractor creates a new DataExtractor
func NewDataExtractor(aiClient *azure.OpenAIClient, logger *zap.Logger) *DataExtractor {
	return &DataExtractor{
		aiClient: aiClient,
		logger:   logger,
	}
}

// Extract extracts structured health data from conversation history
func (de *DataExtractor) Extract(ctx context.Context, conversationHistory []ConversationMessage) (*ExtractedData, error) {
	de.logger.Info("starting data extraction from conversation",
		zap.Int("message_count", len(conversationHistory)),
	)

	// Build conversation history string
	var conversationText strings.Builder
	for _, msg := range conversationHistory {
		conversationText.WriteString(fmt.Sprintf("%s: %s\n", msg.Role, msg.Content))
	}

	// Create AI prompt for data extraction
	prompt := de.buildExtractionPrompt(conversationText.String())

	// Call Azure OpenAI
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt),
		openai.UserMessage("Extract the health data from the conversation above and return it as JSON."),
	}

	response, err := de.aiClient.Complete(ctx, messages)
	if err != nil {
		de.logger.Error("AI extraction failed", zap.Error(err))
		return nil, fmt.Errorf("AI extraction failed: %w", err)
	}

	// Parse JSON response
	extractedData, err := de.parseExtractionResponse(response)
	if err != nil {
		de.logger.Error("failed to parse extraction response",
			zap.Error(err),
			zap.String("response", response),
		)
		return nil, fmt.Errorf("failed to parse extraction response: %w", err)
	}

	de.logger.Info("data extraction completed successfully",
		zap.String("mood", extractedData.Mood),
		zap.String("energy_level", extractedData.EnergyLevel),
		zap.String("sleep_quality", extractedData.SleepQuality),
		zap.Int("symptoms_count", len(extractedData.Symptoms)),
	)

	return extractedData, nil
}

// buildExtractionPrompt creates the AI prompt for data extraction
func (de *DataExtractor) buildExtractionPrompt(conversationHistory string) string {
	return fmt.Sprintf(`You are a medical data extraction assistant. Extract structured health information from the following conversation in Hungarian.

Conversation:
%s

Extract the following information and return it as valid JSON:
{
  "symptoms": ["list of symptoms mentioned"],
  "mood": "positive/neutral/negative",
  "pain_level": 0-10 or null if no pain reported,
  "energy_level": "low/medium/high",
  "sleep_quality": "poor/fair/good/excellent",
  "medication_taken": "yes/no/partial",
  "physical_activity": ["list of activities mentioned"],
  "meals": {
    "breakfast": "description or empty string",
    "lunch": "description or empty string",
    "dinner": "description or empty string"
  },
  "general_feeling": "free text summary of how they feel",
  "additional_notes": "any other relevant information"
}

Rules:
- If information is not mentioned, use empty strings for text fields, empty arrays for lists, or null for pain_level
- Mood should be classified based on the overall tone of the conversation
- Energy level should be inferred from their descriptions
- Sleep quality should be based on their sleep description
- Medication taken should be "yes" if they took all medications, "no" if they took none, "partial" if they took some
- Extract all symptoms and pain descriptions mentioned
- Extract all physical activities mentioned (sports, walks, exercise)
- Return ONLY valid JSON, no additional text

Return the JSON now:`, conversationHistory)
}

// parseExtractionResponse parses the AI response into ExtractedData
func (de *DataExtractor) parseExtractionResponse(response string) (*ExtractedData, error) {
	// Clean up response - sometimes AI adds markdown code blocks
	response = strings.TrimSpace(response)
	response = strings.TrimPrefix(response, "```json")
	response = strings.TrimPrefix(response, "```")
	response = strings.TrimSuffix(response, "```")
	response = strings.TrimSpace(response)

	var data ExtractedData
	if err := json.Unmarshal([]byte(response), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Validate and normalize extracted data
	data = de.normalizeExtractedData(data)

	return &data, nil
}

// normalizeExtractedData validates and normalizes the extracted data
func (de *DataExtractor) normalizeExtractedData(data ExtractedData) ExtractedData {
	// Normalize mood
	data.Mood = strings.ToLower(strings.TrimSpace(data.Mood))
	if data.Mood != "positive" && data.Mood != "neutral" && data.Mood != "negative" {
		de.logger.Warn("invalid mood value, defaulting to neutral", zap.String("mood", data.Mood))
		data.Mood = "neutral"
	}

	// Normalize energy level
	data.EnergyLevel = strings.ToLower(strings.TrimSpace(data.EnergyLevel))
	if data.EnergyLevel != "low" && data.EnergyLevel != "medium" && data.EnergyLevel != "high" {
		de.logger.Warn("invalid energy level, defaulting to medium", zap.String("energy_level", data.EnergyLevel))
		data.EnergyLevel = "medium"
	}

	// Normalize sleep quality
	data.SleepQuality = strings.ToLower(strings.TrimSpace(data.SleepQuality))
	if data.SleepQuality != "poor" && data.SleepQuality != "fair" && data.SleepQuality != "good" && data.SleepQuality != "excellent" {
		de.logger.Warn("invalid sleep quality, defaulting to fair", zap.String("sleep_quality", data.SleepQuality))
		data.SleepQuality = "fair"
	}

	// Normalize medication taken
	data.MedicationTaken = strings.ToLower(strings.TrimSpace(data.MedicationTaken))
	if data.MedicationTaken != "yes" && data.MedicationTaken != "no" && data.MedicationTaken != "partial" {
		de.logger.Warn("invalid medication taken value, defaulting to no", zap.String("medication_taken", data.MedicationTaken))
		data.MedicationTaken = "no"
	}

	// Validate pain level
	if data.PainLevel != nil {
		if *data.PainLevel < 0 {
			de.logger.Warn("pain level below 0, setting to 0", zap.Int("pain_level", *data.PainLevel))
			zero := 0
			data.PainLevel = &zero
		} else if *data.PainLevel > 10 {
			de.logger.Warn("pain level above 10, setting to 10", zap.Int("pain_level", *data.PainLevel))
			ten := 10
			data.PainLevel = &ten
		}
	}

	// Initialize empty arrays if nil
	if data.Symptoms == nil {
		data.Symptoms = []string{}
	}
	if data.PhysicalActivity == nil {
		data.PhysicalActivity = []string{}
	}

	return data
}

// ConversationMessage represents a message in the conversation
type ConversationMessage struct {
	Role    string
	Content string
}
