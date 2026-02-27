package azure

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/azure"
	"go.uber.org/zap"
)

// OpenAIClient wraps Azure OpenAI SDK with retry logic and logging
type OpenAIClient struct {
	client     *openai.Client
	deployment string
	logger     *zap.Logger
	maxRetries int
	baseDelay  time.Duration
}

// NewOpenAIClient creates a new Azure OpenAI client using the openai-go SDK with Azure extensions
func NewOpenAIClient(endpoint, apiKey, deployment string, logger *zap.Logger) (*OpenAIClient, error) {
	if endpoint == "" || apiKey == "" || deployment == "" {
		return nil, fmt.Errorf("endpoint, apiKey, and deployment are required")
	}

	// Create OpenAI client with Azure configuration
	client := openai.NewClient(
		azure.WithEndpoint(endpoint, "2024-08-01-preview"),
		azure.WithAPIKey(apiKey),
	)

	return &OpenAIClient{
		client:     &client,
		deployment: deployment,
		logger:     logger,
		maxRetries: 3,
		baseDelay:  time.Second,
	}, nil
}

// Complete sends a chat completion request to Azure OpenAI with retry logic
func (c *OpenAIClient) Complete(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	startTime := time.Now()
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			delay := c.baseDelay * time.Duration(1<<uint(attempt-1))
			c.logger.Info("retrying Azure OpenAI request",
				zap.Int("attempt", attempt+1),
				zap.Duration("delay", delay),
			)
			time.Sleep(delay)
		}

		result, err := c.complete(ctx, messages)
		if err == nil {
			processingTime := time.Since(startTime)
			c.logger.Info("Azure OpenAI request completed",
				zap.Duration("processing_time", processingTime),
				zap.Int("attempts", attempt+1),
			)
			return result, nil
		}

		lastErr = err
		if !c.isRetryable(err) {
			c.logger.Error("non-retryable Azure OpenAI error",
				zap.Error(err),
				zap.Int("attempt", attempt+1),
			)
			break
		}

		c.logger.Warn("Azure OpenAI request failed, will retry",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
		)
	}

	processingTime := time.Since(startTime)
	c.logger.Error("Azure OpenAI request failed after retries",
		zap.Error(lastErr),
		zap.Duration("total_time", processingTime),
		zap.Int("max_retries", c.maxRetries),
	)

	return "", fmt.Errorf("Azure OpenAI request failed after %d attempts: %w", c.maxRetries, lastErr)
}

// complete performs a single chat completion request
func (c *OpenAIClient) complete(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (string, error) {
	requestStart := time.Now()

	resp, err := c.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModel(c.deployment),
		Messages: messages,
	})

	if err != nil {
		return "", fmt.Errorf("chat completion request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices returned from Azure OpenAI")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	// Log token usage and processing time
	requestTime := time.Since(requestStart)
	c.logger.Info("Azure OpenAI token usage",
		zap.Int64("prompt_tokens", resp.Usage.PromptTokens),
		zap.Int64("completion_tokens", resp.Usage.CompletionTokens),
		zap.Int64("total_tokens", resp.Usage.TotalTokens),
		zap.Duration("request_time", requestTime),
	)

	return content, nil
}

// isRetryable determines if an error should trigger a retry
func (c *OpenAIClient) isRetryable(err error) bool {
	// Retry on network errors, timeouts, and rate limits
	// Don't retry on authentication errors or invalid requests
	if err == nil {
		return false
	}

	// Check for context cancellation or deadline exceeded
	if ctx := context.Background(); ctx.Err() != nil {
		return false
	}

	// For now, retry on all errors except explicit non-retryable ones
	// In production, you'd want more sophisticated error classification
	errStr := err.Error()

	// Don't retry authentication errors
	if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "unauthorized") || strings.Contains(errStr, "401") {
		return false
	}

	// Don't retry invalid request errors
	if strings.Contains(errStr, "invalid") || strings.Contains(errStr, "bad request") || strings.Contains(errStr, "400") {
		return false
	}

	// Retry everything else (rate limits, timeouts, network errors)
	return true
}
