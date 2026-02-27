# VPC Module

Creates an Azure Virtual Network with multi-AZ support and three subnet types:
- Public subnets for NAT Gateway (one per AZ)
- Private subnets for services with NAT Gateway for outbound internet (one per AZ)
- Private subnets for databases, fully isolated (one per AZ)

## Usage

```hcl
module "vpc" {
  source = "./modules/vpc"

  name                = "vnet-healthcare-dev"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.main.name
  
  address_space      = ["10.0.0.0/16"]
  availability_zones = ["1", "2", "3"]
  
  public_subnet_cidrs            = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  private_services_subnet_cidrs  = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]
  private_db_subnet_cidrs        = ["10.0.21.0/24", "10.0.22.0/24", "10.0.23.0/24"]

  tags = {
    Environment = "dev"
  }
}
```

## Single AZ Usage

```hcl
module "vpc" {
  source = "./modules/vpc"

  name                = "vnet-healthcare-dev"
  location            = "westeurope"
  resource_group_name = azurerm_resource_group.main.name
  
  address_space      = ["10.0.0.0/16"]
  availability_zones = ["1"]
  
  public_subnet_cidrs            = ["10.0.1.0/24"]
  private_services_subnet_cidrs  = ["10.0.11.0/24"]
  private_db_subnet_cidrs        = ["10.0.21.0/24"]

  tags = {
    Environment = "dev"
  }
}
```
