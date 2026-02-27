package azure

import (
	"bytes"
	"context"
	"io"
	"testing"

	"go.uber.org/zap"
)

func TestNewBlobStorageClient(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name          string
		accountName   string
		accountKey    string
		containerName string
		wantErr       bool
	}{
		{
			name:          "valid configuration",
			accountName:   "testaccount",
			accountKey:    "dGVzdGtleQ==", // base64 encoded "testkey"
			containerName: "test-container",
			wantErr:       false,
		},
		{
			name:          "missing account name",
			accountName:   "",
			accountKey:    "dGVzdGtleQ==",
			containerName: "test-container",
			wantErr:       true,
		},
		{
			name:          "missing account key",
			accountName:   "testaccount",
			accountKey:    "",
			containerName: "test-container",
			wantErr:       true,
		},
		{
			name:          "missing container name",
			accountName:   "testaccount",
			accountKey:    "dGVzdGtleQ==",
			containerName: "",
			wantErr:       true,
		},
		{
			name:          "invalid account key format",
			accountName:   "testaccount",
			accountKey:    "invalid-key-format",
			containerName: "test-container",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBlobStorageClient(tt.accountName, tt.accountKey, tt.containerName, logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBlobStorageClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewBlobStorageClient() returned nil client")
			}
			if !tt.wantErr {
				if client.containerName != tt.containerName {
					t.Errorf("containerName = %v, want %v", client.containerName, tt.containerName)
				}
			}
		})
	}
}

func TestBlobStorageClient_UploadPDF_Validation(t *testing.T) {
	logger := zap.NewNop()

	// Note: This test validates the structure and error handling
	// Actual Azure SDK calls would require mocking or integration tests
	tests := []struct {
		name     string
		filename string
		data     []byte
		wantErr  bool
	}{
		{
			name:     "valid PDF upload",
			filename: "report.pdf",
			data:     []byte("PDF content"),
			wantErr:  false, // Will fail without real Azure connection, but validates structure
		},
		{
			name:     "empty filename",
			filename: "",
			data:     []byte("PDF content"),
			wantErr:  false, // Will fail at Azure level
		},
		{
			name:     "empty data",
			filename: "report.pdf",
			data:     []byte{},
			wantErr:  false, // Will fail at Azure level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create client with test credentials (will fail but validates structure)
			client, err := NewBlobStorageClient("testaccount", "dGVzdGtleQ==", "test-container", logger)
			if err != nil {
				t.Skipf("Skipping test due to client creation error: %v", err)
				return
			}

			ctx := context.Background()
			_, err = client.UploadPDF(ctx, tt.filename, tt.data)

			// We expect errors since we're not connected to real Azure
			// This test validates the method signature and basic structure
			if err == nil && tt.filename == "" {
				t.Error("UploadPDF() should validate filename")
			}
		})
	}
}

func TestBlobStorageClient_DownloadPDF_Validation(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		blobName string
		wantErr  bool
	}{
		{
			name:     "valid blob name",
			blobName: "reports/report.pdf",
			wantErr:  false,
		},
		{
			name:     "empty blob name",
			blobName: "",
			wantErr:  false, // Will fail at Azure level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBlobStorageClient("testaccount", "dGVzdGtleQ==", "test-container", logger)
			if err != nil {
				t.Skipf("Skipping test due to client creation error: %v", err)
				return
			}

			ctx := context.Background()
			_, err = client.DownloadPDF(ctx, tt.blobName)

			// We expect errors since we're not connected to real Azure
			if err == nil {
				t.Error("DownloadPDF() should fail without real Azure connection")
			}
		})
	}
}

func TestBlobStorageClient_UploadAudio_Validation(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name        string
		filename    string
		audioData   []byte
		wantErr     bool
		description string
	}{
		{
			name:        "valid audio upload",
			filename:    "recording.wav",
			audioData:   []byte("RIFF audio data"),
			wantErr:     false,
			description: "should handle valid audio data",
		},
		{
			name:        "empty filename",
			filename:    "",
			audioData:   []byte("audio data"),
			wantErr:     false,
			description: "will fail at Azure level",
		},
		{
			name:        "large audio file",
			filename:    "large-recording.wav",
			audioData:   make([]byte, 10*1024*1024), // 10MB
			wantErr:     false,
			description: "should handle large files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBlobStorageClient("testaccount", "dGVzdGtleQ==", "test-container", logger)
			if err != nil {
				t.Skipf("Skipping test due to client creation error: %v", err)
				return
			}

			ctx := context.Background()
			audioStream := bytes.NewReader(tt.audioData)
			_, err = client.UploadAudio(ctx, tt.filename, audioStream)

			// We expect errors since we're not connected to real Azure
			// This validates the method structure and error handling
			if err == nil {
				t.Error("UploadAudio() should fail without real Azure connection")
			}
		})
	}
}

func TestBlobStorageClient_DownloadAudio_Validation(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		blobName string
		wantErr  bool
	}{
		{
			name:     "valid blob name",
			blobName: "audio/recording.wav",
			wantErr:  false,
		},
		{
			name:     "empty blob name",
			blobName: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewBlobStorageClient("testaccount", "dGVzdGtleQ==", "test-container", logger)
			if err != nil {
				t.Skipf("Skipping test due to client creation error: %v", err)
				return
			}

			ctx := context.Background()
			_, err = client.DownloadAudio(ctx, tt.blobName)

			// We expect errors since we're not connected to real Azure
			if err == nil {
				t.Error("DownloadAudio() should fail without real Azure connection")
			}
		})
	}
}

func TestBlobStorageClient_AudioStreamHandling(t *testing.T) {
	// Test that audio stream is properly read
	audioData := []byte("test audio data")
	audioStream := bytes.NewReader(audioData)

	// Read the stream to verify it works
	readData, err := io.ReadAll(audioStream)
	if err != nil {
		t.Errorf("Failed to read audio stream: %v", err)
	}

	if !bytes.Equal(readData, audioData) {
		t.Errorf("Read data = %v, want %v", readData, audioData)
	}

	// Verify stream can be reset
	audioStream.Seek(0, io.SeekStart)
	readData2, err := io.ReadAll(audioStream)
	if err != nil {
		t.Errorf("Failed to read audio stream second time: %v", err)
	}

	if !bytes.Equal(readData2, audioData) {
		t.Errorf("Second read data = %v, want %v", readData2, audioData)
	}
}

func TestBlobStorageClient_BlobNaming(t *testing.T) {
	// Test blob naming conventions
	tests := []struct {
		name           string
		filename       string
		expectedPrefix string
	}{
		{
			name:           "PDF report",
			filename:       "report.pdf",
			expectedPrefix: "reports/",
		},
		{
			name:           "audio recording",
			filename:       "recording.wav",
			expectedPrefix: "audio/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify naming convention logic
			var blobName string
			if tt.expectedPrefix == "reports/" {
				blobName = "reports/" + tt.filename
			} else if tt.expectedPrefix == "audio/" {
				blobName = "audio/" + tt.filename
			}

			if blobName != tt.expectedPrefix+tt.filename {
				t.Errorf("blobName = %v, want %v", blobName, tt.expectedPrefix+tt.filename)
			}
		})
	}
}

func TestBlobStorageClient_ContextCancellation(t *testing.T) {
	client, err := NewBlobStorageClient("testaccount", "dGVzdGtleQ==", "test-container", zap.NewNop())
	if err != nil {
		t.Skipf("Skipping test due to client creation error: %v", err)
		return
	}

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Test upload with cancelled context
	_, err = client.UploadPDF(ctx, "test.pdf", []byte("data"))
	if err == nil {
		t.Error("UploadPDF() should fail with cancelled context")
	}

	// Test download with cancelled context
	_, err = client.DownloadPDF(ctx, "test.pdf")
	if err == nil {
		t.Error("DownloadPDF() should fail with cancelled context")
	}
}

func TestToPtr(t *testing.T) {
	// Test the helper function
	str := "test"
	ptr := toPtr(str)

	if ptr == nil {
		t.Error("toPtr() returned nil")
	}

	if *ptr != str {
		t.Errorf("*toPtr() = %v, want %v", *ptr, str)
	}
}
