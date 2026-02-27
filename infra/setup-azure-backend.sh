#!/bin/bash
set -e

# Azure Backend Setup Script
# This script creates the Azure resources needed for Terraform remote state

# Configuration
RESOURCE_GROUP_NAME="Solo-1"
STORAGE_ACCOUNT_NAME="tfstatesolo1healthcare"
CONTAINER_NAME="tfstate"
LOCATION="swedencentral"
SUBSCRIPTION_ID="61c53454-ceb0-49ba-bc5a-6178761ee50d"

echo "🚀 Setting up Azure backend for Terraform..."
echo "Resource Group: $RESOURCE_GROUP_NAME"
echo "Storage Account: $STORAGE_ACCOUNT_NAME"
echo "Container: $CONTAINER_NAME"
echo "Location: $LOCATION"
echo ""

# Check if logged in to Azure
echo "Checking Azure login status..."
if ! az account show &> /dev/null; then
    echo "❌ Not logged in to Azure. Please run: az login"
    exit 1
fi

# Set the subscription
echo "Setting subscription to: $SUBSCRIPTION_ID"
az account set --subscription "$SUBSCRIPTION_ID"
echo "✅ Using subscription: $SUBSCRIPTION_ID"
echo ""

# Check if resource group exists (using existing resource group)
echo "Checking resource group..."
if az group show --name "$RESOURCE_GROUP_NAME" &> /dev/null; then
    echo "✅ Using existing resource group: $RESOURCE_GROUP_NAME"
else
    echo "❌ Resource group $RESOURCE_GROUP_NAME not found"
    exit 1
fi

# Create storage account
echo "Creating storage account..."
az storage account create \
    --name "$STORAGE_ACCOUNT_NAME" \
    --resource-group "$RESOURCE_GROUP_NAME" \
    --location "$LOCATION" \
    --sku Standard_LRS \
    --encryption-services blob \
    --https-only true \
    --min-tls-version TLS1_2 \
    --allow-blob-public-access false \
    --output none

echo "✅ Storage account created"

# Create blob container
echo "Creating blob container..."
az storage container create \
    --name "$CONTAINER_NAME" \
    --account-name "$STORAGE_ACCOUNT_NAME" \
    --auth-mode login \
    --output none

echo "✅ Blob container created"
echo ""

# Update backend.tfvars with the actual storage account name
echo "Updating backend.tfvars..."
cat > backend.tfvars <<EOF
resource_group_name  = "$RESOURCE_GROUP_NAME"
storage_account_name = "$STORAGE_ACCOUNT_NAME"
container_name       = "$CONTAINER_NAME"
key                  = "terraform.tfstate"
EOF

echo "✅ backend.tfvars updated"
echo ""

echo "🎉 Azure backend setup complete!"
echo ""
echo "Next steps:"
echo "1. Initialize Terraform with the backend:"
echo "   cd infra && terraform init -backend-config=backend.tfvars"
echo ""
echo "2. Or if using Terragrunt:"
echo "   cd infra && terragrunt init"
echo ""
echo "Backend configuration:"
echo "  Resource Group: $RESOURCE_GROUP_NAME"
echo "  Storage Account: $STORAGE_ACCOUNT_NAME"
echo "  Container: $CONTAINER_NAME"
echo "  Subscription: $SUBSCRIPTION_ID"
