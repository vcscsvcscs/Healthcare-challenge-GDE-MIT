# Azure OpenAI Module

This module creates an Azure OpenAI service with model deployments.

## Features

- Azure OpenAI Cognitive Service
- Model deployments (GPT-4o, GPT-4, GPT-3.5-turbo, etc.)
- Configurable public network access
- Custom subdomain support

## Usage

```hcl
module "openai" {
  source = "./modules/openai"

  name                = "openai-healthcare-dev"
  location            = "swedencentral"
  resource_group_name = "Solo-1"
  sku_name            = "S0"

  custom_subdomain_name         = "openai-healthcare-dev"
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
    Environment = "dev"
    Project     = "Healthcare"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | Name of the Azure OpenAI service | string | - | yes |
| location | Azure region | string | - | yes |
| resource_group_name | Resource group name | string | - | yes |
| sku_name | SKU name | string | "S0" | no |
| custom_subdomain_name | Custom subdomain | string | - | yes |
| public_network_access_enabled | Enable public access | bool | true | no |
| deployments | List of model deployments | list(object) | [] | no |
| tags | Resource tags | map(string) | {} | no |

## Outputs

| Name | Description |
|------|-------------|
| id | OpenAI service ID |
| endpoint | OpenAI service endpoint |
| primary_access_key | Primary access key (sensitive) |
| secondary_access_key | Secondary access key (sensitive) |
| custom_subdomain | Custom subdomain name |
| deployments | Map of deployment IDs |
