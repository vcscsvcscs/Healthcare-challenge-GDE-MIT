# Azure Cognitive Services Speech Module

This Terraform module creates an Azure Cognitive Services Speech account for speech-to-text and text-to-speech capabilities.

## Features

- Creates Azure Cognitive Services Speech account
- Configurable SKU (Free or Standard tier)
- Optional network ACLs for security
- Custom subdomain support
- Public network access control

## Usage

```hcl
module "speech" {
  source = "./modules/speech"

  name                = "my-speech-service"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.main.name
  sku_name            = "S0"

  tags = {
    Environment = "dev"
    Project     = "healthcare"
  }
}
```

## Example with Network Restrictions

```hcl
module "speech" {
  source = "./modules/speech"

  name                = "my-speech-service"
  location            = "eastus"
  resource_group_name = azurerm_resource_group.main.name
  sku_name            = "S0"

  custom_subdomain_name = "my-speech-subdomain"
  
  network_acls = {
    default_action = "Deny"
    ip_rules       = ["203.0.113.0/24"]
    subnet_id      = azurerm_subnet.main.id
  }

  public_network_access_enabled = false

  tags = {
    Environment = "production"
  }
}
```

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|----------|
| name | Name of the Speech service | string | - | yes |
| location | Azure region | string | - | yes |
| resource_group_name | Resource group name | string | - | yes |
| sku_name | SKU (F0 or S0) | string | "S0" | no |
| custom_subdomain_name | Custom subdomain | string | null | no |
| network_acls | Network ACL configuration | object | null | no |
| public_network_access_enabled | Enable public access | bool | true | no |
| tags | Resource tags | map(string) | {} | no |

## Outputs

| Name | Description |
|------|-------------|
| id | Speech service ID |
| endpoint | Service endpoint URL |
| primary_access_key | Primary access key (sensitive) |
| secondary_access_key | Secondary access key (sensitive) |
| name | Service name |

## SKU Options

- **F0**: Free tier (limited requests)
- **S0**: Standard tier (pay-as-you-go)
