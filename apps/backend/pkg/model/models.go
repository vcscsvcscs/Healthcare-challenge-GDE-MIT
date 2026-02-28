package model

import "time"

// User represents a user in the system
type User struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// SessionStatus represents the status of a check-in session
type SessionStatus string

const (
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCompleted SessionStatus = "completed"
	SessionStatusExpired   SessionStatus = "expired"
)

// Session represents a check-in session
type Session struct {
	ID          string        `json:"id"`
	UserID      string        `json:"user_id"`
	StartedAt   time.Time     `json:"started_at"`
	CompletedAt *time.Time    `json:"completed_at,omitempty"`
	ExpiredAt   *time.Time    `json:"expired_at,omitempty"`
	Status      SessionStatus `json:"status"`
	Messages    []Message     `json:"messages,omitempty"`
}

// MessageRole represents the role of a message sender
type MessageRole string

const (
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleUser      MessageRole = "user"
)

// Message represents a conversation message
type Message struct {
	ID            string      `json:"id"`
	SessionID     string      `json:"session_id"`
	Role          MessageRole `json:"role"`
	Content       string      `json:"content"`
	AudioFilePath *string     `json:"audio_file_path,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// AudioRecording represents an audio recording
type AudioRecording struct {
	ID              string    `json:"id"`
	SessionID       string    `json:"session_id"`
	MessageID       *string   `json:"message_id,omitempty"`
	FilePath        string    `json:"file_path"`
	DurationSeconds *float64  `json:"duration_seconds,omitempty"`
	Transcription   *string   `json:"transcription,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// HealthCheckIn represents a completed health check-in with extracted data
type HealthCheckIn struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	SessionID        *string   `json:"session_id,omitempty"`
	CheckInDate      time.Time `json:"check_in_date"`
	Symptoms         []string  `json:"symptoms,omitempty"`
	Mood             *string   `json:"mood,omitempty"`
	PainLevel        *int      `json:"pain_level,omitempty"`
	EnergyLevel      *string   `json:"energy_level,omitempty"`
	SleepQuality     *string   `json:"sleep_quality,omitempty"`
	MedicationTaken  *string   `json:"medication_taken,omitempty"`
	PhysicalActivity []string  `json:"physical_activity,omitempty"`
	Breakfast        *string   `json:"breakfast,omitempty"`
	Lunch            *string   `json:"lunch,omitempty"`
	Dinner           *string   `json:"dinner,omitempty"`
	GeneralFeeling   *string   `json:"general_feeling,omitempty"`
	AdditionalNotes  *string   `json:"additional_notes,omitempty"`
	RawTranscript    *string   `json:"raw_transcript,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// Medication represents a medication record
type Medication struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	Dosage    string     `json:"dosage"`
	Frequency string     `json:"frequency"`
	StartDate time.Time  `json:"start_date"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Notes     *string    `json:"notes,omitempty"`
	Active    bool       `json:"active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// MedicationLog represents a medication adherence log entry
type MedicationLog struct {
	ID           string    `json:"id"`
	MedicationID string    `json:"medication_id"`
	TakenAt      time.Time `json:"taken_at"`
	Adherence    bool      `json:"adherence"`
	CreatedAt    time.Time `json:"created_at"`
}

// MenstruationCycle represents a menstruation cycle record
type MenstruationCycle struct {
	ID            string     `json:"id"`
	UserID        string     `json:"user_id"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	FlowIntensity *string    `json:"flow_intensity,omitempty"`
	Symptoms      []string   `json:"symptoms,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// BloodPressureReading represents a blood pressure measurement
type BloodPressureReading struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Systolic   int       `json:"systolic"`
	Diastolic  int       `json:"diastolic"`
	Pulse      int       `json:"pulse"`
	MeasuredAt time.Time `json:"measured_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// FitnessDataPoint represents a fitness data point from Health Connect
type FitnessDataPoint struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Date         time.Time `json:"date"`
	DataType     string    `json:"data_type"` // steps, heart_rate, sleep, calories, distance, active_minutes
	Value        float64   `json:"value"`
	Unit         string    `json:"unit"`           // count, bpm, minutes, kcal, meters
	Source       string    `json:"source"`         // health_connect, google_fit
	SourceDataID string    `json:"source_data_id"` // Original ID from Health Connect
	CreatedAt    time.Time `json:"created_at"`
}

// Report represents a generated health report
type Report struct {
	ID             string    `json:"id"`
	UserID         string    `json:"user_id"`
	DateRangeStart time.Time `json:"date_range_start"`
	DateRangeEnd   time.Time `json:"date_range_end"`
	FilePath       string    `json:"file_path"`
	GeneratedAt    time.Time `json:"generated_at"`
	CreatedAt      time.Time `json:"created_at"`
}
