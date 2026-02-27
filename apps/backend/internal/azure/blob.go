package azure

import (
	"context"
	"fmt"
	"io"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"go.uber.org/zap"
)

// BlobStorageClient wraps Azure Blob Storage SDK for file operations
type BlobStorageClient struct {
	client        *azblob.Client
	containerName string
	logger        *zap.Logger
}

// NewBlobStorageClient creates a new Azure Blob Storage client
func NewBlobStorageClient(accountName, accountKey, containerName string, logger *zap.Logger) (*BlobStorageClient, error) {
	if accountName == "" || accountKey == "" || containerName == "" {
		return nil, fmt.Errorf("accountName, accountKey, and containerName are required")
	}

	// Create service URL
	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)

	// Create shared key credential
	credential, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create shared key credential: %w", err)
	}

	// Create blob client
	client, err := azblob.NewClientWithSharedKeyCredential(serviceURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create blob client: %w", err)
	}

	return &BlobStorageClient{
		client:        client,
		containerName: containerName,
		logger:        logger,
	}, nil
}

// UploadPDF uploads a PDF file to Azure Blob Storage
func (c *BlobStorageClient) UploadPDF(ctx context.Context, filename string, data []byte) (string, error) {
	c.logger.Info("uploading PDF to blob storage",
		zap.String("filename", filename),
		zap.Int("size_bytes", len(data)),
	)

	blobName := fmt.Sprintf("reports/%s", filename)

	// Get blob client
	blobClient := c.client.ServiceClient().NewContainerClient(c.containerName).NewBlockBlobClient(blobName)

	// Upload with metadata
	_, err := blobClient.UploadBuffer(ctx, data, &azblob.UploadBufferOptions{
		Metadata: map[string]*string{
			"contenttype": toPtr("application/pdf"),
		},
	})

	if err != nil {
		c.logger.Error("failed to upload PDF",
			zap.String("filename", filename),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to upload PDF: %w", err)
	}

	c.logger.Info("PDF uploaded successfully",
		zap.String("blob_name", blobName),
	)

	return blobName, nil
}

// DownloadPDF downloads a PDF file from Azure Blob Storage
func (c *BlobStorageClient) DownloadPDF(ctx context.Context, blobName string) ([]byte, error) {
	c.logger.Info("downloading PDF from blob storage",
		zap.String("blob_name", blobName),
	)

	// Get blob client
	blobClient := c.client.ServiceClient().NewContainerClient(c.containerName).NewBlockBlobClient(blobName)

	// Download blob
	downloadResponse, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		c.logger.Error("failed to download PDF",
			zap.String("blob_name", blobName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to download PDF: %w", err)
	}
	defer downloadResponse.Body.Close()

	// Read all data
	data, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		c.logger.Error("failed to read PDF data",
			zap.String("blob_name", blobName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}

	c.logger.Info("PDF downloaded successfully",
		zap.String("blob_name", blobName),
		zap.Int("size_bytes", len(data)),
	)

	return data, nil
}

// UploadAudio uploads an audio file to Azure Blob Storage
func (c *BlobStorageClient) UploadAudio(ctx context.Context, filename string, audioStream io.Reader) (string, error) {
	c.logger.Info("uploading audio to blob storage",
		zap.String("filename", filename),
	)

	blobName := fmt.Sprintf("audio/%s", filename)

	// Get blob client
	blobClient := c.client.ServiceClient().NewContainerClient(c.containerName).NewBlockBlobClient(blobName)

	// Read audio data from stream
	audioData, err := io.ReadAll(audioStream)
	if err != nil {
		c.logger.Error("failed to read audio stream",
			zap.String("filename", filename),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to read audio stream: %w", err)
	}

	// Upload with metadata
	_, err = blobClient.UploadBuffer(ctx, audioData, &azblob.UploadBufferOptions{
		Metadata: map[string]*string{
			"contenttype": toPtr("audio/wav"),
		},
	})

	if err != nil {
		c.logger.Error("failed to upload audio",
			zap.String("filename", filename),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to upload audio: %w", err)
	}

	c.logger.Info("audio uploaded successfully",
		zap.String("blob_name", blobName),
		zap.Int("size_bytes", len(audioData)),
	)

	return blobName, nil
}

// DownloadAudio downloads an audio file from Azure Blob Storage
func (c *BlobStorageClient) DownloadAudio(ctx context.Context, blobName string) ([]byte, error) {
	c.logger.Info("downloading audio from blob storage",
		zap.String("blob_name", blobName),
	)

	// Get blob client
	blobClient := c.client.ServiceClient().NewContainerClient(c.containerName).NewBlockBlobClient(blobName)

	// Download blob
	downloadResponse, err := blobClient.DownloadStream(ctx, nil)
	if err != nil {
		c.logger.Error("failed to download audio",
			zap.String("blob_name", blobName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer downloadResponse.Body.Close()

	// Read all data
	data, err := io.ReadAll(downloadResponse.Body)
	if err != nil {
		c.logger.Error("failed to read audio data",
			zap.String("blob_name", blobName),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to read audio data: %w", err)
	}

	c.logger.Info("audio downloaded successfully",
		zap.String("blob_name", blobName),
		zap.Int("size_bytes", len(data)),
	)

	return data, nil
}

// toPtr is a helper function to convert a value to a pointer
func toPtr(s string) *string {
	return &s
}
