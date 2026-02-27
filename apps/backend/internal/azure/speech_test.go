package azure

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewSpeechServiceClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name            string
		subscriptionKey string
		region          string
		wantErr         bool
	}{
		{
			name:            "valid configuration",
			subscriptionKey: "test-key",
			region:          "swedencentral",
			wantErr:         false,
		},
		{
			name:            "missing subscription key",
			subscriptionKey: "",
			region:          "swedencentral",
			wantErr:         true,
		},
		{
			name:            "missing region",
			subscriptionKey: "test-key",
			region:          "",
			wantErr:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewSpeechServiceClient(tt.subscriptionKey, tt.region, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSpeechServiceClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewSpeechServiceClient() returned nil client")
			}
			if !tt.wantErr {
				expectedEndpoint := "https://swedencentral.stt.speech.microsoft.com"
				if client.endpoint != expectedEndpoint {
					t.Errorf("endpoint = %v, want %v", client.endpoint, expectedEndpoint)
				}
				if client.region != tt.region {
					t.Errorf("region = %v, want %v", client.region, tt.region)
				}
				if client.httpClient.Timeout != 60*time.Second {
					t.Errorf("timeout = %v, want 60s", client.httpClient.Timeout)
				}
			}
		})
	}
}

func TestSpeechServiceClient_StreamAudioToText_Success(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Ocp-Apim-Subscription-Key") != "test-key" {
			t.Error("Missing or incorrect subscription key header")
		}
		if r.Header.Get("Content-Type") != "audio/wav; codecs=audio/pcm; samplerate=16000" {
			t.Error("Missing or incorrect content type header")
		}

		// Return success response
		response := map[string]interface{}{
			"RecognitionStatus": "Success",
			"DisplayText":       "Ez egy teszt szöveg",
			"Offset":            0,
			"Duration":          10000000,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	// Create mock audio stream
	audioData := []byte("mock audio data")
	audioStream := bytes.NewReader(audioData)

	ctx := context.Background()
	result, err := client.StreamAudioToText(ctx, audioStream)

	if err != nil {
		t.Errorf("StreamAudioToText() error = %v", err)
	}
	if result != "Ez egy teszt szöveg" {
		t.Errorf("StreamAudioToText() = %v, want 'Ez egy teszt szöveg'", result)
	}
}

func TestSpeechServiceClient_StreamAudioToText_RecognitionFailed(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server that returns failed recognition
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"RecognitionStatus": "NoMatch",
			"DisplayText":       "",
			"Offset":            0,
			"Duration":          0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	audioStream := bytes.NewReader([]byte("mock audio data"))
	ctx := context.Background()

	_, err := client.StreamAudioToText(ctx, audioStream)
	if err == nil {
		t.Error("StreamAudioToText() should return error for failed recognition")
	}
}

func TestSpeechServiceClient_StreamAudioToText_HTTPError(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Invalid subscription key"))
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "invalid-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	audioStream := bytes.NewReader([]byte("mock audio data"))
	ctx := context.Background()

	_, err := client.StreamAudioToText(ctx, audioStream)
	if err == nil {
		t.Error("StreamAudioToText() should return error for HTTP error")
	}
}

func TestSpeechServiceClient_StreamAudioToText_InvalidJSON(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server that returns invalid JSON
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	audioStream := bytes.NewReader([]byte("mock audio data"))
	ctx := context.Background()

	_, err := client.StreamAudioToText(ctx, audioStream)
	if err == nil {
		t.Error("StreamAudioToText() should return error for invalid JSON")
	}
}

func TestSpeechServiceClient_TextToSpeech_Success(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request headers
		if r.Header.Get("Ocp-Apim-Subscription-Key") != "test-key" {
			t.Error("Missing or incorrect subscription key header")
		}
		if r.Header.Get("Content-Type") != "application/ssml+xml" {
			t.Error("Missing or incorrect content type header")
		}
		if r.Header.Get("X-Microsoft-OutputFormat") != "audio-16khz-32kbitrate-mono-mp3" {
			t.Error("Missing or incorrect output format header")
		}

		// Verify SSML content
		body, _ := io.ReadAll(r.Body)
		if !bytes.Contains(body, []byte("hu-HU-NoemiNeural")) {
			t.Error("SSML should contain Hungarian voice name")
		}
		if !bytes.Contains(body, []byte("Szia")) {
			t.Error("SSML should contain the text")
		}

		// Return mock audio data
		w.Header().Set("Content-Type", "audio/mpeg")
		w.Write([]byte("mock audio mp3 data"))
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		ttsEndpoint:     server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	ctx := context.Background()
	audioData, err := client.TextToSpeech(ctx, "Szia", "hu-HU")

	if err != nil {
		t.Errorf("TextToSpeech() error = %v", err)
	}
	if len(audioData) == 0 {
		t.Error("TextToSpeech() returned empty audio data")
	}
	if string(audioData) != "mock audio mp3 data" {
		t.Errorf("TextToSpeech() = %v, want 'mock audio mp3 data'", string(audioData))
	}
}

func TestSpeechServiceClient_TextToSpeech_HTTPError(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server that returns HTTP error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid SSML"))
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		ttsEndpoint:     server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	ctx := context.Background()
	_, err := client.TextToSpeech(ctx, "Test", "hu-HU")

	if err == nil {
		t.Error("TextToSpeech() should return error for HTTP error")
	}
}

func TestSpeechServiceClient_TextToSpeechWAV_Success(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify WAV format header
		if r.Header.Get("X-Microsoft-OutputFormat") != "riff-16khz-16bit-mono-pcm" {
			t.Error("Missing or incorrect WAV output format header")
		}

		// Return mock WAV audio data
		w.Header().Set("Content-Type", "audio/wav")
		w.Write([]byte("RIFF mock wav data"))
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		ttsEndpoint:     server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	ctx := context.Background()
	audioData, err := client.TextToSpeechWAV(ctx, "Test", "hu-HU")

	if err != nil {
		t.Errorf("TextToSpeechWAV() error = %v", err)
	}
	if len(audioData) == 0 {
		t.Error("TextToSpeechWAV() returned empty audio data")
	}
	if string(audioData) != "RIFF mock wav data" {
		t.Errorf("TextToSpeechWAV() = %v, want 'RIFF mock wav data'", string(audioData))
	}
}

func TestSpeechServiceClient_ContextCancellation(t *testing.T) {
	logger := zap.NewNop()

	// Create mock server with delay
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := &SpeechServiceClient{
		subscriptionKey: "test-key",
		region:          "swedencentral",
		endpoint:        server.URL,
		httpClient:      &http.Client{Timeout: 60 * time.Second},
		logger:          logger,
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	audioStream := bytes.NewReader([]byte("mock audio data"))
	_, err := client.StreamAudioToText(ctx, audioStream)

	if err == nil {
		t.Error("StreamAudioToText() should return error for cancelled context")
	}
}
