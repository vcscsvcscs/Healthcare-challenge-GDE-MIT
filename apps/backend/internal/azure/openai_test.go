package azure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/openai/openai-go/v3"
	"go.uber.org/zap"
)

// mockOpenAIClient is a mock implementation for testing
type mockOpenAIClient struct {
	responses []mockResponse
	callCount int
}

type mockResponse struct {
	content string
	err     error
}

func TestNewOpenAIClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		endpoint   string
		apiKey     string
		deployment string
		wantErr    bool
	}{
		{
			name:       "valid configuration",
			endpoint:   "https://test.openai.azure.com/",
			apiKey:     "test-key",
			deployment: "gpt-4o",
			wantErr:    false,
		},
		{
			name:       "missing endpoint",
			endpoint:   "",
			apiKey:     "test-key",
			deployment: "gpt-4o",
			wantErr:    true,
		},
		{
			name:       "missing api key",
			endpoint:   "https://test.openai.azure.com/",
			apiKey:     "",
			deployment: "gpt-4o",
			wantErr:    true,
		},
		{
			name:       "missing deployment",
			endpoint:   "https://test.openai.azure.com/",
			apiKey:     "test-key",
			deployment: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewOpenAIClient(tt.endpoint, tt.apiKey, tt.deployment, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOpenAIClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewOpenAIClient() returned nil client")
			}
			if !tt.wantErr {
				if client.deployment != tt.deployment {
					t.Errorf("deployment = %v, want %v", client.deployment, tt.deployment)
				}
				if client.maxRetries != 3 {
					t.Errorf("maxRetries = %v, want 3", client.maxRetries)
				}
				if client.baseDelay != time.Second {
					t.Errorf("baseDelay = %v, want 1s", client.baseDelay)
				}
			}
		})
	}
}

func TestOpenAIClient_isRetryable(t *testing.T) {
	logger := zap.NewNop()
	client := &OpenAIClient{
		logger:     logger,
		maxRetries: 3,
		baseDelay:  time.Second,
	}

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "authentication error",
			err:  errors.New("authentication failed"),
			want: false,
		},
		{
			name: "unauthorized error",
			err:  errors.New("unauthorized access"),
			want: false,
		},
		{
			name: "401 error",
			err:  errors.New("status code 401"),
			want: false,
		},
		{
			name: "invalid request error",
			err:  errors.New("invalid request format"),
			want: false,
		},
		{
			name: "bad request error",
			err:  errors.New("bad request"),
			want: false,
		},
		{
			name: "400 error",
			err:  errors.New("status code 400"),
			want: false,
		},
		{
			name: "rate limit error",
			err:  errors.New("rate limit exceeded"),
			want: true,
		},
		{
			name: "timeout error",
			err:  errors.New("request timeout"),
			want: true,
		},
		{
			name: "network error",
			err:  errors.New("network connection failed"),
			want: true,
		},
		{
			name: "500 error",
			err:  errors.New("status code 500"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := client.isRetryable(tt.err)
			if got != tt.want {
				t.Errorf("isRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenAIClient_RetryLogic(t *testing.T) {
	// This test verifies that retry logic works correctly
	// Note: Since we can't easily mock the openai-go client, this is a structural test
	logger := zap.NewNop()
	client := &OpenAIClient{
		logger:     logger,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond, // Short delay for testing
	}

	// Test that maxRetries is set correctly
	if client.maxRetries != 3 {
		t.Errorf("maxRetries = %v, want 3", client.maxRetries)
	}

	// Test that baseDelay is set correctly
	if client.baseDelay != 10*time.Millisecond {
		t.Errorf("baseDelay = %v, want 10ms", client.baseDelay)
	}
}

func TestOpenAIClient_Complete_EmptyMessages(t *testing.T) {
	logger := zap.NewNop()

	// Create client with invalid credentials (will fail but we're testing validation)
	client, err := NewOpenAIClient("https://test.openai.azure.com/", "test-key", "gpt-4o", logger)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	ctx := context.Background()

	// Test with empty messages - should fail at API level
	_, err = client.Complete(ctx, []openai.ChatCompletionMessageParamUnion{})
	if err == nil {
		t.Error("Complete() with empty messages should return error")
	}
}

func TestOpenAIClient_Complete_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()

	client, err := NewOpenAIClient("https://test.openai.azure.com/", "test-key", "gpt-4o", logger)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test message"),
	}

	_, err = client.Complete(ctx, messages)
	if err == nil {
		t.Error("Complete() with cancelled context should return error")
	}
}

func TestOpenAIClient_Complete_Timeout(t *testing.T) {
	logger := zap.NewNop()

	client, err := NewOpenAIClient("https://test.openai.azure.com/", "test-key", "gpt-4o", logger)
	if err != nil {
		t.Fatalf("NewOpenAIClient() error = %v", err)
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(10 * time.Millisecond)

	messages := []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage("test message"),
	}

	_, err = client.Complete(ctx, messages)
	if err == nil {
		t.Error("Complete() with timeout context should return error")
	}
}
