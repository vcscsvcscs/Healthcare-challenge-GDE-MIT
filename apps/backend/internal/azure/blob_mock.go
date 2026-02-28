package azure

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"go.uber.org/zap"
)

// MockBlobStorageClient is an in-memory implementation of BlobStorageClient for testing
type MockBlobStorageClient struct {
	Storage map[string][]byte
	mu      sync.RWMutex
	logger  *zap.Logger
}

// NewMockBlobStorageClient creates a new mock blob storage client
func NewMockBlobStorageClient(logger *zap.Logger) *MockBlobStorageClient {
	return &MockBlobStorageClient{
		Storage: make(map[string][]byte),
		logger:  logger,
	}
}

// UploadPDF uploads a PDF file to in-memory storage
func (c *MockBlobStorageClient) UploadPDF(ctx context.Context, filename string, data []byte) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	blobName := fmt.Sprintf("reports/%s", filename)
	c.Storage[blobName] = data

	if c.logger != nil {
		c.logger.Info("mock: PDF uploaded",
			zap.String("blob_name", blobName),
			zap.Int("size_bytes", len(data)),
		)
	}

	return blobName, nil
}

// DownloadPDF downloads a PDF file from in-memory storage
func (c *MockBlobStorageClient) DownloadPDF(ctx context.Context, blobName string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, exists := c.Storage[blobName]
	if !exists {
		return nil, fmt.Errorf("blob not found: %s", blobName)
	}

	if c.logger != nil {
		c.logger.Info("mock: PDF downloaded",
			zap.String("blob_name", blobName),
			zap.Int("size_bytes", len(data)),
		)
	}

	return data, nil
}

// UploadAudio uploads an audio file to in-memory storage
func (c *MockBlobStorageClient) UploadAudio(ctx context.Context, filename string, audioStream io.Reader) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	blobName := fmt.Sprintf("audio/%s", filename)

	// Read audio data from stream
	audioData, err := io.ReadAll(audioStream)
	if err != nil {
		return "", fmt.Errorf("failed to read audio stream: %w", err)
	}

	c.Storage[blobName] = audioData

	if c.logger != nil {
		c.logger.Info("mock: audio uploaded",
			zap.String("blob_name", blobName),
			zap.Int("size_bytes", len(audioData)),
		)
	}

	return blobName, nil
}

// DownloadAudio downloads an audio file from in-memory storage
func (c *MockBlobStorageClient) DownloadAudio(ctx context.Context, blobName string) ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, exists := c.Storage[blobName]
	if !exists {
		return nil, fmt.Errorf("blob not found: %s", blobName)
	}

	if c.logger != nil {
		c.logger.Info("mock: audio downloaded",
			zap.String("blob_name", blobName),
			zap.Int("size_bytes", len(data)),
		)
	}

	return bytes.Clone(data), nil
}

// Clear removes all data from in-memory storage
func (c *MockBlobStorageClient) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Storage = make(map[string][]byte)

	if c.logger != nil {
		c.logger.Info("mock: storage cleared")
	}
}

// ListBlobs returns all blob names in storage
func (c *MockBlobStorageClient) ListBlobs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	blobs := make([]string, 0, len(c.Storage))
	for name := range c.Storage {
		blobs = append(blobs, name)
	}

	return blobs
}
