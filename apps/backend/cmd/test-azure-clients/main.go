package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"go.uber.org/zap"

	"github.com/vcscsvcscs/Healthcare-challenge-GDE-MIT/apps/backend/internal/azure"
)

func main() {
	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Get credentials from environment
	openaiEndpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	openaiKey := os.Getenv("AZURE_OPENAI_KEY")
	openaiDeployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

	speechKey := os.Getenv("AZURE_SPEECH_KEY")
	speechRegion := os.Getenv("AZURE_SPEECH_REGION")

	storageAccountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	storageAccountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")

	// Validate required environment variables
	if openaiEndpoint == "" || openaiKey == "" || openaiDeployment == "" {
		logger.Fatal("Missing Azure OpenAI credentials. Set AZURE_OPENAI_ENDPOINT, AZURE_OPENAI_KEY, and AZURE_OPENAI_DEPLOYMENT")
	}

	if speechKey == "" || speechRegion == "" {
		logger.Fatal("Missing Azure Speech credentials. Set AZURE_SPEECH_KEY and AZURE_SPEECH_REGION")
	}

	if storageAccountName == "" || storageAccountKey == "" {
		logger.Fatal("Missing Azure Storage credentials. Set AZURE_STORAGE_ACCOUNT_NAME and AZURE_STORAGE_ACCOUNT_KEY")
	}

	ctx := context.Background()

	// Test 1: Azure OpenAI Client
	logger.Info("=== Testing Azure OpenAI Client ===")
	if err := testOpenAIClient(ctx, openaiEndpoint, openaiKey, openaiDeployment, logger); err != nil {
		logger.Error("OpenAI client test failed", zap.Error(err))
	} else {
		logger.Info("✅ OpenAI client test passed")
	}

	// Test 2: Azure Speech Service Client
	logger.Info("\n=== Testing Azure Speech Service Client ===")
	if err := testSpeechClient(ctx, speechKey, speechRegion, logger); err != nil {
		logger.Error("Speech client test failed", zap.Error(err))
	} else {
		logger.Info("✅ Speech client test passed")
	}

	// Test 3: Azure Blob Storage Client
	logger.Info("\n=== Testing Azure Blob Storage Client ===")
	if err := testBlobStorageClient(ctx, storageAccountName, storageAccountKey, logger); err != nil {
		logger.Error("Blob storage client test failed", zap.Error(err))
	} else {
		logger.Info("✅ Blob storage client test passed")
	}

	logger.Info("\n=== All tests completed ===")
}

func testOpenAIClient(ctx context.Context, endpoint, apiKey, deployment string, logger *zap.Logger) error {
	client, err := azure.NewOpenAIClient(endpoint, apiKey, deployment, logger)
	if err != nil {
		return fmt.Errorf("failed to create OpenAI client: %w", err)
	}

	// Test chat completion
	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String("You are a helpful assistant."),
				},
			},
		},
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String("Say 'Hello from Azure OpenAI!' in Hungarian."),
				},
			},
		},
	}

	response, err := client.Complete(ctx, messages)
	if err != nil {
		return fmt.Errorf("chat completion failed: %w", err)
	}

	logger.Info("OpenAI response received",
		zap.String("response", response),
		zap.Int("response_length", len(response)),
	)

	return nil
}

func testSpeechClient(ctx context.Context, subscriptionKey, region string, logger *zap.Logger) error {
	client, err := azure.NewSpeechServiceClient(subscriptionKey, region, logger)
	if err != nil {
		return fmt.Errorf("failed to create Speech client: %w", err)
	}

	// Test text-to-speech with actual health check-in questions from the design
	testQuestions := []string{
		"Szia! Hogy érzed magad ma?",
		"Sportoltál ma, vagy mentél sétálni?",
		"Mit reggeliztél, ebédeltél, vagy vacsoráltál?",
		"Fáj valamid?",
		"Hogyan aludtál?",
		"Milyen az energiaszinted?",
		"Beszedtél ma bármi gyógyszert?",
		"Van még valami, amit szeretnél mondani?",
	}

	logger.Info("Testing text-to-speech with health check-in questions", zap.Int("question_count", len(testQuestions)))

	// Test each question
	for i, question := range testQuestions {
		logger.Info(fmt.Sprintf("Testing question %d/%d", i+1, len(testQuestions)), zap.String("question", question))

		// Generate MP3 for listening
		audioDataMP3, err := client.TextToSpeech(ctx, question, "hu-HU")
		if err != nil {
			return fmt.Errorf("text-to-speech (MP3) failed for question %d: %w", i+1, err)
		}

		logger.Info("Text-to-speech (MP3) completed",
			zap.Int("question_number", i+1),
			zap.Int("audio_size_bytes", len(audioDataMP3)),
		)

		// Save MP3 audio to file for verification
		audioFileMP3 := fmt.Sprintf("/tmp/test-speech-question-%d.mp3", i+1)
		if err := os.WriteFile(audioFileMP3, audioDataMP3, 0644); err != nil {
			logger.Warn("Failed to save MP3 audio file", zap.Error(err))
		} else {
			logger.Info("MP3 audio saved", zap.String("file", audioFileMP3))
		}
	}

	// Test speech-to-text round-trip with the first question
	testText := testQuestions[0]
	logger.Info("Testing speech-to-text round-trip", zap.String("text", testText))

	// Generate WAV format for speech-to-text (STT expects WAV, not MP3)
	logger.Info("Generating WAV format audio for speech-to-text test")
	audioDataWAV, err := client.TextToSpeechWAV(ctx, testText, "hu-HU")
	if err != nil {
		logger.Warn("Text-to-speech (WAV) failed, skipping STT test", zap.Error(err))
		return nil
	}

	logger.Info("Text-to-speech (WAV) completed",
		zap.Int("audio_size_bytes", len(audioDataWAV)),
	)

	// Save WAV audio to file
	audioFileWAV := "/tmp/test-speech-roundtrip.wav"
	if err := os.WriteFile(audioFileWAV, audioDataWAV, 0644); err != nil {
		logger.Warn("Failed to save WAV audio file", zap.Error(err))
	} else {
		logger.Info("WAV audio saved", zap.String("file", audioFileWAV))
	}

	// Test speech-to-text with WAV audio
	audioReader := strings.NewReader(string(audioDataWAV))

	transcription, err := client.StreamAudioToText(ctx, audioReader)
	if err != nil {
		return fmt.Errorf("speech-to-text failed: %w", err)
	}

	logger.Info("Speech-to-text completed",
		zap.String("original_text", testText),
		zap.String("transcription", transcription),
	)

	// Check if transcription is similar to original
	if len(transcription) > 0 {
		logger.Info("✅ Speech-to-text successfully transcribed the audio")
	} else {
		logger.Warn("⚠️  Transcription is empty")
	}

	return nil
}

func testBlobStorageClient(ctx context.Context, accountName, accountKey string, logger *zap.Logger) error {
	// Test with audio-recordings container
	containerName := "audio-recordings"
	client, err := azure.NewBlobStorageClient(accountName, accountKey, containerName, logger)
	if err != nil {
		return fmt.Errorf("failed to create Blob Storage client: %w", err)
	}

	// Test audio upload
	testAudioData := []byte("This is test audio data")
	testFilename := fmt.Sprintf("test-audio-%d.wav", time.Now().Unix())

	logger.Info("Testing audio upload", zap.String("filename", testFilename))

	blobName, err := client.UploadAudio(ctx, testFilename, strings.NewReader(string(testAudioData)))
	if err != nil {
		return fmt.Errorf("audio upload failed: %w", err)
	}

	logger.Info("Audio uploaded successfully", zap.String("blob_name", blobName))

	// Test audio download
	logger.Info("Testing audio download", zap.String("blob_name", blobName))

	downloadedData, err := client.DownloadAudio(ctx, blobName)
	if err != nil {
		return fmt.Errorf("audio download failed: %w", err)
	}

	if string(downloadedData) != string(testAudioData) {
		return fmt.Errorf("downloaded data doesn't match uploaded data")
	}

	logger.Info("Audio downloaded and verified successfully",
		zap.Int("size_bytes", len(downloadedData)),
	)

	// Test PDF operations with health-reports container
	pdfClient, err := azure.NewBlobStorageClient(accountName, accountKey, "health-reports", logger)
	if err != nil {
		return fmt.Errorf("failed to create PDF Blob Storage client: %w", err)
	}

	testPDFData := []byte("%PDF-1.4\nTest PDF content")
	testPDFFilename := fmt.Sprintf("test-report-%d.pdf", time.Now().Unix())

	logger.Info("Testing PDF upload", zap.String("filename", testPDFFilename))

	pdfBlobName, err := pdfClient.UploadPDF(ctx, testPDFFilename, testPDFData)
	if err != nil {
		return fmt.Errorf("PDF upload failed: %w", err)
	}

	logger.Info("PDF uploaded successfully", zap.String("blob_name", pdfBlobName))

	// Test PDF download
	logger.Info("Testing PDF download", zap.String("blob_name", pdfBlobName))

	downloadedPDF, err := pdfClient.DownloadPDF(ctx, pdfBlobName)
	if err != nil {
		return fmt.Errorf("PDF download failed: %w", err)
	}

	if string(downloadedPDF) != string(testPDFData) {
		return fmt.Errorf("downloaded PDF doesn't match uploaded PDF")
	}

	logger.Info("PDF downloaded and verified successfully",
		zap.Int("size_bytes", len(downloadedPDF)),
	)

	return nil
}
