package azure

import (
	"context"
	"io"
)

// BlobStorage defines the interface for blob storage operations
// This interface allows for easier testing with mock implementations
type BlobStorage interface {
	UploadPDF(ctx context.Context, filename string, data []byte) (string, error)
	DownloadPDF(ctx context.Context, blobName string) ([]byte, error)
	UploadAudio(ctx context.Context, filename string, audioStream io.Reader) (string, error)
	DownloadAudio(ctx context.Context, blobName string) ([]byte, error)
}

// Ensure BlobStorageClient implements BlobStorage interface
var _ BlobStorage = (*BlobStorageClient)(nil)
