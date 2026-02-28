package service

import (
	"testing"

	"go.uber.org/zap"
)

func TestDataExtractor_normalizeExtractedData(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	de := &DataExtractor{logger: logger}

	tests := []struct {
		name     string
		input    ExtractedData
		expected ExtractedData
	}{
		{
			name: "valid data",
			input: ExtractedData{
				Mood:            "positive",
				EnergyLevel:     "high",
				SleepQuality:    "good",
				MedicationTaken: "yes",
				PainLevel:       intPtr(5),
			},
			expected: ExtractedData{
				Mood:             "positive",
				EnergyLevel:      "high",
				SleepQuality:     "good",
				MedicationTaken:  "yes",
				PainLevel:        intPtr(5),
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "invalid mood defaults to neutral",
			input: ExtractedData{
				Mood:            "happy",
				EnergyLevel:     "medium",
				SleepQuality:    "fair",
				MedicationTaken: "no",
			},
			expected: ExtractedData{
				Mood:             "neutral",
				EnergyLevel:      "medium",
				SleepQuality:     "fair",
				MedicationTaken:  "no",
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "invalid energy level defaults to medium",
			input: ExtractedData{
				Mood:            "positive",
				EnergyLevel:     "super",
				SleepQuality:    "good",
				MedicationTaken: "yes",
			},
			expected: ExtractedData{
				Mood:             "positive",
				EnergyLevel:      "medium",
				SleepQuality:     "good",
				MedicationTaken:  "yes",
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "invalid sleep quality defaults to fair",
			input: ExtractedData{
				Mood:            "neutral",
				EnergyLevel:     "low",
				SleepQuality:    "amazing",
				MedicationTaken: "partial",
			},
			expected: ExtractedData{
				Mood:             "neutral",
				EnergyLevel:      "low",
				SleepQuality:     "fair",
				MedicationTaken:  "partial",
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "pain level below 0 clamped to 0",
			input: ExtractedData{
				Mood:            "positive",
				EnergyLevel:     "high",
				SleepQuality:    "excellent",
				MedicationTaken: "yes",
				PainLevel:       intPtr(-5),
			},
			expected: ExtractedData{
				Mood:             "positive",
				EnergyLevel:      "high",
				SleepQuality:     "excellent",
				MedicationTaken:  "yes",
				PainLevel:        intPtr(0),
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "pain level above 10 clamped to 10",
			input: ExtractedData{
				Mood:            "negative",
				EnergyLevel:     "low",
				SleepQuality:    "poor",
				MedicationTaken: "no",
				PainLevel:       intPtr(15),
			},
			expected: ExtractedData{
				Mood:             "negative",
				EnergyLevel:      "low",
				SleepQuality:     "poor",
				MedicationTaken:  "no",
				PainLevel:        intPtr(10),
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
		{
			name: "uppercase values normalized to lowercase",
			input: ExtractedData{
				Mood:            "POSITIVE",
				EnergyLevel:     "HIGH",
				SleepQuality:    "EXCELLENT",
				MedicationTaken: "YES",
			},
			expected: ExtractedData{
				Mood:             "positive",
				EnergyLevel:      "high",
				SleepQuality:     "excellent",
				MedicationTaken:  "yes",
				Symptoms:         []string{},
				PhysicalActivity: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := de.normalizeExtractedData(tt.input)

			if result.Mood != tt.expected.Mood {
				t.Errorf("mood: expected %s, got %s", tt.expected.Mood, result.Mood)
			}
			if result.EnergyLevel != tt.expected.EnergyLevel {
				t.Errorf("energy_level: expected %s, got %s", tt.expected.EnergyLevel, result.EnergyLevel)
			}
			if result.SleepQuality != tt.expected.SleepQuality {
				t.Errorf("sleep_quality: expected %s, got %s", tt.expected.SleepQuality, result.SleepQuality)
			}
			if result.MedicationTaken != tt.expected.MedicationTaken {
				t.Errorf("medication_taken: expected %s, got %s", tt.expected.MedicationTaken, result.MedicationTaken)
			}

			if tt.expected.PainLevel != nil && result.PainLevel != nil {
				if *result.PainLevel != *tt.expected.PainLevel {
					t.Errorf("pain_level: expected %d, got %d", *tt.expected.PainLevel, *result.PainLevel)
				}
			} else if tt.expected.PainLevel != result.PainLevel {
				t.Errorf("pain_level: expected %v, got %v", tt.expected.PainLevel, result.PainLevel)
			}

			if result.Symptoms == nil {
				t.Error("symptoms should be initialized to empty array")
			}
			if result.PhysicalActivity == nil {
				t.Error("physical_activity should be initialized to empty array")
			}
		})
	}
}

func TestDataExtractor_parseExtractionResponse(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	de := &DataExtractor{logger: logger}

	tests := []struct {
		name        string
		response    string
		expectError bool
	}{
		{
			name: "valid JSON",
			response: `{
				"symptoms": ["headache"],
				"mood": "positive",
				"pain_level": 3,
				"energy_level": "high",
				"sleep_quality": "good",
				"medication_taken": "yes",
				"physical_activity": ["walking"],
				"meals": {
					"breakfast": "toast",
					"lunch": "salad",
					"dinner": "pasta"
				},
				"general_feeling": "good",
				"additional_notes": "none"
			}`,
			expectError: false,
		},
		{
			name: "JSON with markdown code blocks",
			response: "```json\n" + `{
				"symptoms": [],
				"mood": "neutral",
				"pain_level": null,
				"energy_level": "medium",
				"sleep_quality": "fair",
				"medication_taken": "no",
				"physical_activity": [],
				"meals": {
					"breakfast": "",
					"lunch": "",
					"dinner": ""
				},
				"general_feeling": "",
				"additional_notes": ""
			}` + "\n```",
			expectError: false,
		},
		{
			name:        "invalid JSON",
			response:    `{invalid json}`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := de.parseExtractionResponse(tt.response)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if result == nil {
					t.Error("expected result, got nil")
				}
			}
		})
	}
}

func TestDataExtractor_buildExtractionPrompt(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	de := &DataExtractor{logger: logger}

	conversationHistory := "assistant: Szia! Hogy érzed magad ma?\nuser: Jól érzem magam"

	prompt := de.buildExtractionPrompt(conversationHistory)

	if prompt == "" {
		t.Error("expected non-empty prompt")
	}

	// Check that prompt contains the conversation history
	if !contains(prompt, conversationHistory) {
		t.Error("prompt should contain conversation history")
	}

	// Check that prompt contains key instructions
	expectedKeywords := []string{"symptoms", "mood", "pain_level", "energy_level", "sleep_quality", "medication_taken"}
	for _, keyword := range expectedKeywords {
		if !contains(prompt, keyword) {
			t.Errorf("prompt should contain keyword: %s", keyword)
		}
	}
}

// Helper functions
func intPtr(i int) *int {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
