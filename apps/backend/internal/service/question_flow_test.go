package service

import (
	"testing"
)

func TestQuestionFlow_GetNextQuestion(t *testing.T) {
	qf := NewQuestionFlow()

	// Test getting first question
	q1 := qf.GetNextQuestion()
	if q1 == nil {
		t.Fatal("expected first question, got nil")
	}
	if q1.ID != "q1_general_feeling" {
		t.Errorf("expected q1_general_feeling, got %s", q1.ID)
	}
	if q1.TextHU != "Szia! Hogy érzed magad ma?" {
		t.Errorf("unexpected question text: %s", q1.TextHU)
	}

	// Test getting all questions
	for i := 1; i < 8; i++ {
		q := qf.GetNextQuestion()
		if q == nil {
			t.Fatalf("expected question %d, got nil", i+1)
		}
	}

	// Test that we've reached the end
	if !qf.IsComplete() {
		t.Error("expected IsComplete to be true after all questions")
	}

	// Test that GetNextQuestion returns nil after completion
	qNext := qf.GetNextQuestion()
	if qNext != nil {
		t.Error("expected nil after all questions answered")
	}
}

func TestQuestionFlow_GetQuestionByID(t *testing.T) {
	qf := NewQuestionFlow()

	// Test getting question by ID
	q := qf.GetQuestionByID("q4_pain")
	if q == nil {
		t.Fatal("expected question, got nil")
	}
	if q.TextHU != "Fáj valamid?" {
		t.Errorf("unexpected question text: %s", q.TextHU)
	}

	// Test getting non-existent question
	qNil := qf.GetQuestionByID("non_existent")
	if qNil != nil {
		t.Error("expected nil for non-existent question")
	}
}

func TestQuestionFlow_ValidateResponse(t *testing.T) {
	qf := NewQuestionFlow()

	// Test valid response for required question
	err := qf.ValidateResponse("q1_general_feeling", "Jól érzem magam")
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	// Test empty response for required question
	err = qf.ValidateResponse("q1_general_feeling", "")
	if err == nil {
		t.Error("expected error for empty response on required question")
	}

	// Test empty response for optional question
	err = qf.ValidateResponse("q8_additional_notes", "")
	if err != nil {
		t.Errorf("expected no error for empty response on optional question, got: %v", err)
	}

	// Test invalid question ID
	err = qf.ValidateResponse("invalid_id", "response")
	if err == nil {
		t.Error("expected error for invalid question ID")
	}
}

func TestQuestionFlow_Reset(t *testing.T) {
	qf := NewQuestionFlow()

	// Get a few questions
	qf.GetNextQuestion()
	qf.GetNextQuestion()
	qf.GetNextQuestion()

	if qf.GetCurrentQuestionIndex() != 3 {
		t.Errorf("expected current index 3, got %d", qf.GetCurrentQuestionIndex())
	}

	// Reset
	qf.Reset()

	if qf.GetCurrentQuestionIndex() != 0 {
		t.Errorf("expected current index 0 after reset, got %d", qf.GetCurrentQuestionIndex())
	}

	// Verify we can get first question again
	q := qf.GetNextQuestion()
	if q == nil || q.ID != "q1_general_feeling" {
		t.Error("expected to get first question after reset")
	}
}

func TestQuestionFlow_GetTotalQuestions(t *testing.T) {
	qf := NewQuestionFlow()

	total := qf.GetTotalQuestions()
	if total != 8 {
		t.Errorf("expected 8 questions, got %d", total)
	}
}
