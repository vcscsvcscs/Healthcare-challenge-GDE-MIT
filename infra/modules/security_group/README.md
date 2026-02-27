# Security Group Module

Creates an Azure Network Security Group with configurable rules.

## Usage

```hcl
module "services_nsg" {
  source = "./modules/security_group"

  name                = "nsg-services-dev"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.main.name
  subnet_id           = module.vpc.private_services_subnet_id

  security_rules = [
    {
      name                       = "allow-https"
      priority                   = 100
      direction                  = "Inbound"
      access                     = "Allow"
      protocol                   = "Tcp"
      source_port_range          = "*"
      destination_port_range     = "443"
      source_address_prefix      = "*"
      destination_address_prefix = "*"
    }
  ]

  tags = {
    Environment = "dev"
  }
}
```
