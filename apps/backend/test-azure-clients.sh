#!/bin/bash

# Test Azure Clients Script
# This script extracts credentials from Terraform output and tests all Azure clients

set -e

echo "🔍 Extracting Azure credentials from Terraform..."

# Navigate to infra directory
TERRAFORM_DIR="../../infra"

# Check if infra directory exists
if [ ! -d "$TERRAFORM_DIR" ]; then
    echo "❌ Infra directory not found at $TERRAFORM_DIR"
    echo "Please run this script from apps/backend/"
    exit 1
fi

# Extract values from terragrunt output
cd "$TERRAFORM_DIR"

export AZURE_OPENAI_ENDPOINT=$(terragrunt output -raw openai_endpoint 2>/dev/null || echo "")
export AZURE_OPENAI_KEY=$(terragrunt output -raw openai_key 2>/dev/null || echo "")
export AZURE_OPENAI_DEPLOYMENT="gpt-4o"

export AZURE_SPEECH_KEY=$(terragrunt output -raw speech_key 2>/dev/null || echo "")
export AZURE_SPEECH_REGION=$(terragrunt output -raw speech_region 2>/dev/null || echo "")

export AZURE_STORAGE_ACCOUNT_NAME=$(terragrunt output -raw storage_account_name 2>/dev/null || echo "")
# Extract storage key from connection string
STORAGE_CONNECTION_STRING=$(terragrunt output -raw storage_connection_string 2>/dev/null || echo "")
# Use sed instead of grep -P for macOS compatibility
export AZURE_STORAGE_ACCOUNT_KEY=$(echo "$STORAGE_CONNECTION_STRING" | sed -n 's/.*AccountKey=\([^;]*\).*/\1/p' || echo "")

cd - > /dev/null

# Validate credentials
echo ""
echo "📋 Validating credentials..."
MISSING_CREDS=0

if [ -z "$AZURE_OPENAI_ENDPOINT" ]; then
    echo "❌ Missing AZURE_OPENAI_ENDPOINT"
    MISSING_CREDS=1
fi

if [ -z "$AZURE_OPENAI_KEY" ]; then
    echo "❌ Missing AZURE_OPENAI_KEY"
    MISSING_CREDS=1
fi

if [ -z "$AZURE_SPEECH_KEY" ]; then
    echo "❌ Missing AZURE_SPEECH_KEY"
    MISSING_CREDS=1
fi

if [ -z "$AZURE_SPEECH_REGION" ]; then
    echo "❌ Missing AZURE_SPEECH_REGION"
    MISSING_CREDS=1
fi

if [ -z "$AZURE_STORAGE_ACCOUNT_NAME" ]; then
    echo "❌ Missing AZURE_STORAGE_ACCOUNT_NAME"
    MISSING_CREDS=1
fi

if [ -z "$AZURE_STORAGE_ACCOUNT_KEY" ]; then
    echo "❌ Missing AZURE_STORAGE_ACCOUNT_KEY"
    MISSING_CREDS=1
fi

if [ $MISSING_CREDS -eq 1 ]; then
    echo ""
    echo "❌ Some credentials are missing. Please check your Terraform outputs."
    exit 1
fi

echo "✅ All credentials found"
echo ""
echo "🔧 Configuration:"
echo "  OpenAI Endpoint: $AZURE_OPENAI_ENDPOINT"
echo "  OpenAI Deployment: $AZURE_OPENAI_DEPLOYMENT"
echo "  Speech Region: $AZURE_SPEECH_REGION"
echo "  Storage Account: $AZURE_STORAGE_ACCOUNT_NAME"
echo ""

# Build and run the test
echo "🏗️  Building test application..."
go build -o /tmp/test-azure-clients ./cmd/test-azure-clients

echo ""
echo "🚀 Running Azure client tests..."
echo "================================================"
echo ""

/tmp/test-azure-clients

echo ""
echo "================================================"
echo "✅ Test completed!"
echo ""
echo "💡 Tip: Check /tmp/test-speech-output.mp3 to hear the Hungarian TTS output"
