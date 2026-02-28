package service

import (
	"fmt"
)

// QuestionType represents the type of question
type QuestionType string

const (
	QuestionTypeOpenEnded QuestionType = "open_ended"
	QuestionTypeNumeric   QuestionType = "numeric"
	QuestionTypeYesNo     QuestionType = "yes_no"
)

// Question represents a health question in the conversation flow
type Question struct {
	ID       string
	TextHU   string
	Type     QuestionType
	Required bool
}

// QuestionFlow manages the sequence of health questions
type QuestionFlow struct {
	questions []Question
	current   int
}

// NewQuestionFlow creates a new QuestionFlow with the Hungarian question set
func NewQuestionFlow() *QuestionFlow {
	questions := []Question{
		{
			ID:       "q1_general_feeling",
			TextHU:   "Szia! Hogy érzed magad ma?",
			Type:     QuestionTypeOpenEnded,
			Required: true,
		},
		{
			ID:       "q2_physical_activity",
			TextHU:   "Sportoltál ma, vagy mentél sétálni?",
			Type:     QuestionTypeYesNo,
			Required: true,
		},
		{
			ID:       "q3_meals",
			TextHU:   "Mit reggeliztél, ebédeltél és vacsoráztál?",
			Type:     QuestionTypeOpenEnded,
			Required: true,
		},
		{
			ID:       "q4_pain",
			TextHU:   "Fáj valamid?",
			Type:     QuestionTypeYesNo,
			Required: true,
		},
		{
			ID:       "q5_sleep",
			TextHU:   "Hogyan aludtál?",
			Type:     QuestionTypeOpenEnded,
			Required: true,
		},
		{
			ID:       "q6_energy",
			TextHU:   "Milyen az energiaszinted?",
			Type:     QuestionTypeOpenEnded,
			Required: true,
		},
		{
			ID:       "q7_medication",
			TextHU:   "Beszedtél ma bármi gyógyszert?",
			Type:     QuestionTypeYesNo,
			Required: true,
		},
		{
			ID:       "q8_additional_notes",
			TextHU:   "Van még valami, amit szeretnél mondani?",
			Type:     QuestionTypeOpenEnded,
			Required: false,
		},
	}

	return &QuestionFlow{
		questions: questions,
		current:   0,
	}
}

// GetNextQuestion returns the next question in the flow
func (qf *QuestionFlow) GetNextQuestion() *Question {
	if qf.current >= len(qf.questions) {
		return nil
	}

	question := &qf.questions[qf.current]
	qf.current++
	return question
}

// GetQuestionByID returns a question by its ID
func (qf *QuestionFlow) GetQuestionByID(questionID string) *Question {
	for i := range qf.questions {
		if qf.questions[i].ID == questionID {
			return &qf.questions[i]
		}
	}
	return nil
}

// GetCurrentQuestionIndex returns the current question index (0-based)
func (qf *QuestionFlow) GetCurrentQuestionIndex() int {
	return qf.current
}

// IsComplete returns true if all questions have been answered
func (qf *QuestionFlow) IsComplete() bool {
	return qf.current >= len(qf.questions)
}

// Reset resets the question flow to the beginning
func (qf *QuestionFlow) Reset() {
	qf.current = 0
}

// GetTotalQuestions returns the total number of questions
func (qf *QuestionFlow) GetTotalQuestions() int {
	return len(qf.questions)
}

// ValidateResponse validates a response based on the question type
func (qf *QuestionFlow) ValidateResponse(questionID string, response string) error {
	question := qf.GetQuestionByID(questionID)
	if question == nil {
		return fmt.Errorf("question not found: %s", questionID)
	}

	// Check if response is empty for required questions
	if question.Required && response == "" {
		return fmt.Errorf("response is required for question: %s", questionID)
	}

	// Additional validation can be added here based on question type
	// For now, we accept any non-empty response for required questions

	return nil
}
