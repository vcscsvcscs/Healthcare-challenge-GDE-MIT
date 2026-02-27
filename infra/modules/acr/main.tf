resource "azurerm_container_registry" "acr" {
  name                = var.name
  resource_group_name = var.resource_group_name
  location            = var.location
  sku                 = var.sku
  admin_enabled       = var.admin_enabled

  public_network_access_enabled = var.public_network_access_enabled

  tags = var.tags
}

# Private endpoint for ACR
resource "azurerm_private_endpoint" "acr" {
  count               = var.enable_private_endpoint ? 1 : 0
  name                = "${var.name}-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id

  private_service_connection {
    name                           = "${var.name}-psc"
    private_connection_resource_id = azurerm_container_registry.acr.id
    is_manual_connection           = false
    subresource_names              = ["registry"]
  }

  private_dns_zone_group {
    name                 = "acr-dns-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.acr[0].id]
  }

  tags = var.tags
}

# Private DNS Zone for ACR
resource "azurerm_private_dns_zone" "acr" {
  count               = var.enable_private_endpoint ? 1 : 0
  name                = "privatelink.azurecr.io"
  resource_group_name = var.resource_group_name

  tags = var.tags
}

# Link Private DNS Zone to VNet
resource "azurerm_private_dns_zone_virtual_network_link" "acr" {
  count                 = var.enable_private_endpoint ? 1 : 0
  name                  = "${var.name}-dns-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.acr[0].name
  virtual_network_id    = var.vnet_id

  tags = var.tags
}
