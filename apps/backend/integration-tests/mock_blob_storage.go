package integration_tests

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
	"go.uber.org/zap"
)

// MockBlobStorageClient is a mock implementation of Azure Blob Storage for testing
type MockBlobStorageClient struct {
	storage map[string][]byte
	mu      sync.RWMutex
	logger  *zap.Logger
}

// Ensure MockBlobStorageClient implements azure.BlobStorage interface
var _ azure.BlobStorage = (*MockBlobStorageClient)(nil)

// NewMockBlobStorageClient creates a new mock blob storage client
func NewMockBlobStorageClient(logger *zap.Logger) *MockBlobStorageClient {
	return &MockBlobStorageClient{
		storage: make(map[string][]byte),
		logger:  logger,
	}
}

// UploadPDF stores a PDF in memory
func (m *MockBlobStorageClient) UploadPDF(ctx context.Context, filename string, data []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	blobPath := fmt.Sprintf("reports/%s", filename)
	m.storage[blobPath] = data

	m.logger.Info("mock: uploaded PDF",
		zap.String("filename", filename),
		zap.String("blob_path", blobPath),
		zap.Int("size", len(data)),
	)

	return blobPath, nil
}

// DownloadPDF retrieves a PDF from memory
func (m *MockBlobStorageClient) DownloadPDF(ctx context.Context, blobPath string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.storage[blobPath]
	if !ok {
		return nil, fmt.Errorf("blob not found: %s", blobPath)
	}

	m.logger.Info("mock: downloaded PDF",
		zap.String("blob_path", blobPath),
		zap.Int("size", len(data)),
	)

	return data, nil
}

// UploadAudio stores audio in memory (not used in this test but required by interface)
func (m *MockBlobStorageClient) UploadAudio(ctx context.Context, filename string, audioStream io.Reader) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Read audio data from stream
	audioData, err := io.ReadAll(audioStream)
	if err != nil {
		return "", fmt.Errorf("failed to read audio stream: %w", err)
	}

	blobPath := fmt.Sprintf("audio/%s", filename)
	m.storage[blobPath] = audioData

	return blobPath, nil
}

// DownloadAudio retrieves audio from memory (not used in this test but required by interface)
func (m *MockBlobStorageClient) DownloadAudio(ctx context.Context, blobPath string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, ok := m.storage[blobPath]
	if !ok {
		return nil, fmt.Errorf("blob not found: %s", blobPath)
	}

	return data, nil
}
