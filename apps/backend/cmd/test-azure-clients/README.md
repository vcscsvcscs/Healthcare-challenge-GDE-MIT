# Azure Clients Test Script

This test script validates all three Azure service clients:
- Azure OpenAI (chat completions)
- Azure Speech Service (text-to-speech in Hungarian)
- Azure Blob Storage (audio and PDF upload/download)

## Quick Start

### Option 1: Using Terraform outputs (Recommended)

From the `apps/backend` directory:

```bash
# Make the script executable
chmod +x test-azure-clients.sh

# Run the test (it will extract credentials from Terraform)
./test-azure-clients.sh
```

### Option 2: Manual credentials

```bash
# Export credentials from Terraform
cd ../../terraform
export AZURE_OPENAI_KEY=$(terraform output -raw openai_key)
export AZURE_SPEECH_KEY=$(terraform output -raw speech_key)
export AZURE_STORAGE_ACCOUNT_KEY=$(terraform output -raw storage_connection_string | grep -oP 'AccountKey=\K[^;]+')
cd -

# Make script executable and run
chmod +x test-azure-clients-manual.sh
./test-azure-clients-manual.sh
```

### Option 3: Direct Go execution

```bash
# Set environment variables
export AZURE_OPENAI_ENDPOINT="https://openai-healthcare-dev.openai.azure.com/"
export AZURE_OPENAI_KEY="your-key"
export AZURE_OPENAI_DEPLOYMENT="gpt-4o"
export AZURE_SPEECH_KEY="your-key"
export AZURE_SPEECH_REGION="swedencentral"
export AZURE_STORAGE_ACCOUNT_NAME="evahealthstoragedev"
export AZURE_STORAGE_ACCOUNT_KEY="your-key"

# Run directly
go run ./cmd/test-azure-clients/main.go
```

## What the test does

1. **OpenAI Client Test**
   - Creates a chat completion request
   - Asks GPT-4o to translate "Hello from Azure OpenAI!" to Hungarian
   - Logs token usage and processing time

2. **Speech Service Test**
   - Converts Hungarian text to speech using NoemiNeural voice
   - Saves the audio to `/tmp/test-speech-output.mp3`
   - You can play this file to verify the Hungarian TTS

3. **Blob Storage Test**
   - Uploads test audio to `audio-recordings` container
   - Downloads and verifies the audio
   - Uploads test PDF to `health-reports` container
   - Downloads and verifies the PDF

## Expected Output

```
=== Testing Azure OpenAI Client ===
INFO    OpenAI response received    {"response": "Helló az Azure OpenAI-tól!", ...}
✅ OpenAI client test passed

=== Testing Azure Speech Service Client ===
INFO    Text-to-speech completed    {"audio_size_bytes": 12345}
INFO    Audio saved for verification    {"file": "/tmp/test-speech-output.mp3"}
✅ Speech client test passed

=== Testing Azure Blob Storage Client ===
INFO    Audio uploaded successfully    {"blob_name": "audio/test-audio-1234567890.wav"}
INFO    Audio downloaded and verified successfully
INFO    PDF uploaded successfully    {"blob_name": "reports/test-report-1234567890.pdf"}
INFO    PDF downloaded and verified successfully
✅ Blob storage client test passed

=== All tests completed ===
```

## Troubleshooting

### Missing credentials
If you see "Missing Azure credentials", make sure:
1. You're in the `apps/backend` directory
2. Terraform has been applied and outputs are available
3. The terraform directory path is correct in the script

### Authentication errors
- Verify your Azure credentials are correct
- Check that the resources exist in your Azure subscription
- Ensure your IP is not blocked by Azure firewall rules

### Storage errors
- Verify the containers `audio-recordings` and `health-reports` exist
- Check storage account access permissions

## Cleanup

The test creates temporary files in blob storage. You can delete them manually or they'll be overwritten on subsequent test runs.

Test files created:
- `/tmp/test-speech-output.mp3` (local)
- `audio/test-audio-*.wav` (blob storage)
- `reports/test-report-*.pdf` (blob storage)
