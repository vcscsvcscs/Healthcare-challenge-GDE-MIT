# Azure Storage Account Module

This module creates an Azure Storage Account with Blob containers for storing health reports and audio recordings.

## Features

- Storage account with secure defaults (HTTPS only, TLS 1.2+)
- Blob versioning enabled
- Soft delete retention policies (7 days)
- Private container access
- Configurable containers

## Usage

```hcl
module "storage" {
  source = "./modules/storage"

  resource_group_name  = "eva-health-rg"
  location             = "eastus"
  storage_account_name = "evahealthstorage"

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
    Environment = "production"
    Project     = "eva-health"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| resource_group_name | Name of the resource group | string | - | yes |
| location | Azure region | string | - | yes |
| storage_account_name | Storage account name (3-24 chars, lowercase alphanumeric) | string | - | yes |
| account_tier | Storage account tier | string | "Standard" | no |
| account_replication_type | Replication type (LRS, GRS, etc.) | string | "LRS" | no |
| containers | List of container configurations | list(object) | See variables.tf | no |
| tags | Resource tags | map(string) | {} | no |

## Outputs

| Name | Description |
|------|-------------|
| storage_account_id | Storage account ID |
| storage_account_name | Storage account name |
| primary_blob_endpoint | Primary blob endpoint URL |
| primary_access_key | Primary access key (sensitive) |
| primary_connection_string | Primary connection string (sensitive) |
| container_names | List of container names |

## Notes

- Storage account names must be globally unique across Azure
- Names must be 3-24 characters, lowercase letters and numbers only
- Blob versioning and soft delete are enabled for data protection
