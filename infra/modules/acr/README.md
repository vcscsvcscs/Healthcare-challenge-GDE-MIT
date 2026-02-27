# Azure Container Registry (ACR) Module

This module creates an Azure Container Registry with optional private endpoint support.

## Features

- Azure Container Registry with configurable SKU
- Optional private endpoint for secure access
- Private DNS zone integration
- Network rules support
- Admin user configuration

## Usage

```hcl
module "acr" {
  source = "./modules/acr"

  name                = "acrhealthcaredev"
  location            = "swedencentral"
  resource_group_name = "Solo-1"
  sku                 = "Basic"
  admin_enabled       = true

  # Optional: Private endpoint
  enable_private_endpoint    = true
  private_endpoint_subnet_id = module.vpc.private_endpoint_subnet_id
  vnet_id                    = module.vpc.vnet_id

  tags = {
    Environment = "dev"
    Project     = "Healthcare"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | Name of the ACR | string | - | yes |
| resource_group_name | Resource group name | string | - | yes |
| location | Azure region | string | - | yes |
| sku | ACR SKU (Basic, Standard, Premium) | string | "Basic" | no |
| admin_enabled | Enable admin user | bool | false | no |
| enable_private_endpoint | Enable private endpoint | bool | false | no |

## Outputs

| Name | Description |
|------|-------------|
| id | ACR ID |
| name | ACR name |
| login_server | ACR login server URL |
| admin_username | ACR admin username (sensitive) |
| admin_password | ACR admin password (sensitive) |
