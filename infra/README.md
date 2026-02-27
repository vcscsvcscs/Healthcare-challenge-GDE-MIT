# Healthcare Infrastructure - Azure Speech Services

Terraform/Terragrunt infrastructure for Azure Cognitive Services Speech (Speech-to-Text).

## Region
- **Primary**: West Europe (Amsterdam) - Central Europe location

## Prerequisites

1. Install required tools:
   ```bash
   # Terraform
   brew install terraform
   
   # Terragrunt
   brew install terragrunt
   
   # Azure CLI
   brew install azure-cli
   ```

2. Login to Azure:
   ```bash
   az login
   ```

3. Set up backend storage (run once):
   ```bash
   cd infra
   ./setup-azure-backend.sh
   ```

## Usage

### Initialize and Apply with Terragrunt

```bash
cd infra

# Initialize
terragrunt init

# Plan changes
terragrunt plan

# Apply infrastructure
terragrunt apply

# Destroy (when needed)
terragrunt destroy
```

### Using Different Environments

```bash
# Development (uses F0 free tier)
ENVIRONMENT=dev terragrunt apply

# Production (uses S0 standard tier)
ENVIRONMENT=prod terragrunt apply
```

### Get Outputs

```bash
# Get speech endpoint
terragrunt output speech_endpoint

# Get all outputs
terragrunt output
```

## Module Structure

```
infra/
├── main.tf                    # Main infrastructure
├── terraform.tf               # Provider configuration
├── terragrunt.hcl            # Terragrunt root config
├── backend.tfvars            # Backend configuration
├── modules/
│   ├── speech/               # Speech service module
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   ├── openai/               # OpenAI service module
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   └── README.md
│   └── storage/              # Blob storage module
│       ├── main.tf
│       ├── variables.tf
│       ├── outputs.tf
│       └── README.md
```

## Configuration

Edit `terragrunt.hcl` to customize:
- Environment name
- Azure region
- Tags

## Security Notes

- Access keys are marked as sensitive in outputs
- Use `terragrunt output -json` to retrieve keys programmatically
- Consider using Azure Key Vault for production secrets
- Storage containers are private by default
- Blob versioning and soft delete enabled for data protection

## Resources Created

1. **Azure Cognitive Services Speech** - Speech-to-text processing
2. **Azure OpenAI Service** - GPT-4o model deployment
3. **Azure Storage Account** - Blob storage with containers:
   - `health-reports` - Stores generated health reports
   - `audio-recordings` - Stores patient audio recordings

## Costs

- **Speech Service F0 (Free tier)**: Limited to 5 audio hours/month
- **Speech Service S0 (Standard)**: ~$1 per audio hour
- **OpenAI GPT-4o**: Pay-per-token pricing
- **Storage Account**: 
  - Standard LRS: ~$0.02 per GB/month
  - Operations: Minimal cost for read/write operations
- See [Azure Pricing](https://azure.microsoft.com/pricing/)
