package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// SpeechServiceClient wraps Azure Speech Service REST API for speech-to-text and text-to-speech
type SpeechServiceClient struct {
	subscriptionKey string
	region          string
	endpoint        string
	ttsEndpoint     string // For testing purposes
	httpClient      *http.Client
	logger          *zap.Logger
}

// NewSpeechServiceClient creates a new Azure Speech Service client
func NewSpeechServiceClient(subscriptionKey, region string, logger *zap.Logger) (*SpeechServiceClient, error) {
	if subscriptionKey == "" || region == "" {
		return nil, fmt.Errorf("subscriptionKey and region are required")
	}

	endpoint := fmt.Sprintf("https://%s.stt.speech.microsoft.com", region)

	return &SpeechServiceClient{
		subscriptionKey: subscriptionKey,
		region:          region,
		endpoint:        endpoint,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		logger: logger,
	}, nil
}

// StreamAudioToText performs real-time speech-to-text transcription from an audio stream
// Note: This implementation uses the REST API for simplicity. For production streaming,
// consider using WebSocket-based streaming or the native SDK with proper C library setup.
func (c *SpeechServiceClient) StreamAudioToText(ctx context.Context, audioStream io.Reader) (string, error) {
	c.logger.Info("starting speech-to-text transcription")

	// Read audio data from stream
	audioData, err := io.ReadAll(audioStream)
	if err != nil {
		return "", fmt.Errorf("failed to read audio stream: %w", err)
	}

	// Create request to Speech-to-Text REST API
	url := fmt.Sprintf("%s/speech/recognition/conversation/cognitiveservices/v1?language=hu-HU", c.endpoint)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(audioData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Ocp-Apim-Subscription-Key", c.subscriptionKey)
	req.Header.Set("Content-Type", "audio/wav; codecs=audio/pcm; samplerate=16000")
	req.Header.Set("Accept", "application/json")

	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("speech-to-text request failed", zap.Error(err))
		return "", fmt.Errorf("speech-to-text request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("speech-to-text request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return "", fmt.Errorf("speech-to-text request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		RecognitionStatus string `json:"RecognitionStatus"`
		DisplayText       string `json:"DisplayText"`
		Offset            int64  `json:"Offset"`
		Duration          int64  `json:"Duration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	processingTime := time.Since(startTime)
	c.logger.Info("speech-to-text transcription completed",
		zap.String("status", result.RecognitionStatus),
		zap.Duration("processing_time", processingTime),
		zap.Int("audio_size_bytes", len(audioData)),
	)

	if result.RecognitionStatus != "Success" {
		return "", fmt.Errorf("recognition failed with status: %s", result.RecognitionStatus)
	}

	return result.DisplayText, nil
}

// TextToSpeech converts text to speech audio in Hungarian
func (c *SpeechServiceClient) TextToSpeech(ctx context.Context, text string, language string) ([]byte, error) {
	c.logger.Info("starting text-to-speech synthesis",
		zap.String("language", language),
		zap.Int("text_length", len(text)),
	)

	// Determine voice name based on language
	voiceName := "hu-HU-NoemiNeural"
	if language != "hu-HU" {
		voiceName = fmt.Sprintf("%s-Standard-A", language)
	}

	// Create SSML request
	ssml := fmt.Sprintf(`<speak version='1.0' xml:lang='%s'>
		<voice xml:lang='%s' name='%s'>
			%s
		</voice>
	</speak>`, language, language, voiceName, text)

	// Create request to Text-to-Speech REST API
	url := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", c.region)
	if c.ttsEndpoint != "" {
		url = c.ttsEndpoint + "/cognitiveservices/v1"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(ssml))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Ocp-Apim-Subscription-Key", c.subscriptionKey)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "audio-16khz-32kbitrate-mono-mp3")
	req.Header.Set("User-Agent", "Eva-Health-Backend")

	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("text-to-speech request failed", zap.Error(err))
		return nil, fmt.Errorf("text-to-speech request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("text-to-speech request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("text-to-speech request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	processingTime := time.Since(startTime)
	c.logger.Info("text-to-speech synthesis completed",
		zap.Int("audio_size_bytes", len(audioData)),
		zap.Duration("processing_time", processingTime),
	)

	return audioData, nil
}

// TextToSpeechWAV converts text to speech audio in WAV format (for speech-to-text compatibility)
func (c *SpeechServiceClient) TextToSpeechWAV(ctx context.Context, text string, language string) ([]byte, error) {
	c.logger.Info("starting text-to-speech synthesis (WAV format)",
		zap.String("language", language),
		zap.Int("text_length", len(text)),
	)

	// Determine voice name based on language
	voiceName := "hu-HU-NoemiNeural"
	if language != "hu-HU" {
		voiceName = fmt.Sprintf("%s-Standard-A", language)
	}

	// Create SSML request
	ssml := fmt.Sprintf(`<speak version='1.0' xml:lang='%s'>
		<voice xml:lang='%s' name='%s'>
			%s
		</voice>
	</speak>`, language, language, voiceName, text)

	// Create request to Text-to-Speech REST API
	url := fmt.Sprintf("https://%s.tts.speech.microsoft.com/cognitiveservices/v1", c.region)
	if c.ttsEndpoint != "" {
		url = c.ttsEndpoint + "/cognitiveservices/v1"
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBufferString(ssml))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers for WAV format
	req.Header.Set("Ocp-Apim-Subscription-Key", c.subscriptionKey)
	req.Header.Set("Content-Type", "application/ssml+xml")
	req.Header.Set("X-Microsoft-OutputFormat", "riff-16khz-16bit-mono-pcm") // WAV format
	req.Header.Set("User-Agent", "Eva-Health-Backend")

	// Send request
	startTime := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("text-to-speech request failed", zap.Error(err))
		return nil, fmt.Errorf("text-to-speech request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.Error("text-to-speech request failed",
			zap.Int("status_code", resp.StatusCode),
			zap.String("response", string(body)),
		)
		return nil, fmt.Errorf("text-to-speech request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read audio data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	processingTime := time.Since(startTime)
	c.logger.Info("text-to-speech synthesis (WAV) completed",
		zap.Int("audio_size_bytes", len(audioData)),
		zap.Duration("processing_time", processingTime),
	)

	return audioData, nil
}
