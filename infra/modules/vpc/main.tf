resource "azurerm_virtual_network" "main" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  address_space       = var.address_space

  tags = var.tags
}

# Public subnets for NAT Gateway (one per AZ)
resource "azurerm_subnet" "public" {
  count = length(var.availability_zones)

  name                 = "${var.name}-public-az${var.availability_zones[count.index]}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.public_subnet_cidrs[count.index]]
}

# Private subnets for services (one per AZ)
resource "azurerm_subnet" "private_services" {
  count = length(var.availability_zones)

  name                 = "${var.name}-private-services-az${var.availability_zones[count.index]}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.private_services_subnet_cidrs[count.index]]
}

# Private subnets for databases (one per AZ)
resource "azurerm_subnet" "private_db" {
  count = length(var.availability_zones)

  name                 = "${var.name}-private-db-az${var.availability_zones[count.index]}"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.private_db_subnet_cidrs[count.index]]
}

# Private endpoint subnet for Azure services
resource "azurerm_subnet" "private_endpoints" {
  name                 = "${var.name}-private-endpoints"
  resource_group_name  = var.resource_group_name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = [var.private_endpoint_subnet_cidr]
}

# NAT Gateway Public IPs (one per AZ)
resource "azurerm_public_ip" "nat" {
  count = length(var.availability_zones)

  name                = "${var.name}-nat-pip-az${var.availability_zones[count.index]}"
  location            = var.location
  resource_group_name = var.resource_group_name
  allocation_method   = "Static"
  sku                 = "Standard"
  zones               = [var.availability_zones[count.index]]

  tags = var.tags
}

# NAT Gateways (one per AZ)
resource "azurerm_nat_gateway" "main" {
  count = length(var.availability_zones)

  name                = "${var.name}-nat-az${var.availability_zones[count.index]}"
  location            = var.location
  resource_group_name = var.resource_group_name
  sku_name            = "Standard"
  zones               = [var.availability_zones[count.index]]

  tags = var.tags
}

# Associate NAT Gateways with Public IPs
resource "azurerm_nat_gateway_public_ip_association" "main" {
  count = length(var.availability_zones)

  nat_gateway_id       = azurerm_nat_gateway.main[count.index].id
  public_ip_address_id = azurerm_public_ip.nat[count.index].id
}

# Associate NAT Gateways with private services subnets
resource "azurerm_subnet_nat_gateway_association" "services" {
  count = length(var.availability_zones)

  subnet_id      = azurerm_subnet.private_services[count.index].id
  nat_gateway_id = azurerm_nat_gateway.main[count.index].id
}
