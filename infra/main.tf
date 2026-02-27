# Main infrastructure configuration - Simplified for public access only

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "location" {
  description = "Azure region"
  type        = string
  default     = "swedencentral"
}

variable "resource_group_name" {
  description = "Existing resource group name"
  type        = string
  default     = "Solo-1"
}

# Use existing Resource Group
data "azurerm_resource_group" "main" {
  name = var.resource_group_name
}

# Speech Service Module - Public access only
module "speech" {
  source = "./modules/speech"

  name                = "speech-healthcare-${var.environment}"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name
  sku_name            = "S0" # Using Standard tier (F0 free tier already exists in subscription)

  custom_subdomain_name = "speech-healthcare-${var.environment}"

  # Public access enabled
  enable_private_endpoint       = false
  public_network_access_enabled = true

  tags = {
    Environment = var.environment
    Project     = "Healthcare"
    ManagedBy   = "Terraform"
  }
}

# Azure OpenAI Service Module - Public access only
module "openai" {
  source = "./modules/openai"

  name                = "openai-healthcare-${var.environment}"
  location            = var.location
  resource_group_name = data.azurerm_resource_group.main.name
  sku_name            = "S0"

  custom_subdomain_name         = "openai-healthcare-${var.environment}"
  public_network_access_enabled = true

  deployments = [
    {
      name          = "gpt-4o"
      model_name    = "gpt-4o"
      model_version = "2024-08-06"
      scale_type    = "Standard"
      capacity      = 10
    }
  ]

  tags = {
    Environment = var.environment
    Project     = "Healthcare"
    ManagedBy   = "Terraform"
  }
}

# Storage Account Module - For health reports and audio recordings
module "storage" {
  source = "./modules/storage"

  resource_group_name  = data.azurerm_resource_group.main.name
  location             = var.location
  storage_account_name = "evahealthstorage${var.environment}"

  containers = [
    {
      name        = "health-reports"
      access_type = "private"
    },
    {
      name        = "audio-recordings"
      access_type = "private"
    }
  ]

  tags = {
    Environment = var.environment
    Project     = "Healthcare"
    ManagedBy   = "Terraform"
  }
}

# Outputs
output "speech_endpoint" {
  description = "Speech service endpoint"
  value       = module.speech.endpoint
}

output "speech_key" {
  description = "Speech service primary key"
  value       = module.speech.primary_access_key
  sensitive   = true
}

output "speech_region" {
  description = "Speech service region"
  value       = var.location
}

output "openai_endpoint" {
  description = "OpenAI service endpoint"
  value       = module.openai.endpoint
}

output "openai_key" {
  description = "OpenAI service primary key"
  value       = module.openai.primary_access_key
  sensitive   = true
}

output "openai_deployments" {
  description = "OpenAI model deployments"
  value       = module.openai.deployments
}

output "resource_group_name" {
  description = "Resource group name"
  value       = data.azurerm_resource_group.main.name
}

output "storage_account_name" {
  description = "Storage account name"
  value       = module.storage.storage_account_name
}

output "storage_blob_endpoint" {
  description = "Storage blob endpoint"
  value       = module.storage.primary_blob_endpoint
}

output "storage_connection_string" {
  description = "Storage connection string"
  value       = module.storage.primary_connection_string
  sensitive   = true
}

output "storage_containers" {
  description = "Storage container names"
  value       = module.storage.container_names
}
