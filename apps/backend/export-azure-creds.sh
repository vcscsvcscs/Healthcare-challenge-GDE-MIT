#!/bin/bash

# Helper script to export Azure credentials from Terraform
# Usage: source ./export-azure-creds.sh

TERRAFORM_DIR="../../infra"

if [ ! -d "$TERRAFORM_DIR" ]; then
    echo "❌ Infra directory not found at $TERRAFORM_DIR"
    return 1 2>/dev/null || exit 1
fi

echo "🔍 Extracting Azure credentials from Terraform..."

cd "$TERRAFORM_DIR"

export AZURE_OPENAI_ENDPOINT=$(terragrunt output -raw openai_endpoint 2>/dev/null)
export AZURE_OPENAI_KEY=$(terragrunt output -raw openai_key 2>/dev/null)
export AZURE_OPENAI_DEPLOYMENT="gpt-4o"

export AZURE_SPEECH_KEY=$(terragrunt output -raw speech_key 2>/dev/null)
export AZURE_SPEECH_REGION=$(terragrunt output -raw speech_region 2>/dev/null)

export AZURE_STORAGE_ACCOUNT_NAME=$(terragrunt output -raw storage_account_name 2>/dev/null)
STORAGE_CONNECTION_STRING=$(terragrunt output -raw storage_connection_string 2>/dev/null)
# Use sed instead of grep -P for macOS compatibility
export AZURE_STORAGE_ACCOUNT_KEY=$(echo "$STORAGE_CONNECTION_STRING" | sed -n 's/.*AccountKey=\([^;]*\).*/\1/p')

cd - > /dev/null

echo "✅ Credentials exported!"
echo ""
echo "You can now run:"
echo "  go run ./cmd/test-azure-clients/main.go"
echo ""
echo "Or use the test script:"
echo "  ./test-azure-clients-manual.sh"
