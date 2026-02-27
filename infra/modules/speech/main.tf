resource "azurerm_cognitive_account" "speech" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  kind                = "SpeechServices"
  sku_name            = var.sku_name

  custom_subdomain_name = var.custom_subdomain_name

  dynamic "network_acls" {
    for_each = var.allowed_subnet_ids != null ? [1] : []
    content {
      default_action = "Deny"
      virtual_network_rules {
        subnet_id                            = var.allowed_subnet_ids[0]
        ignore_missing_vnet_service_endpoint = false
      }
    }
  }

  public_network_access_enabled = var.public_network_access_enabled

  tags = var.tags
}

# Private DNS Zone for Cognitive Services
resource "azurerm_private_dns_zone" "cognitive" {
  count = var.enable_private_endpoint ? 1 : 0

  name                = "privatelink.cognitiveservices.azure.com"
  resource_group_name = var.resource_group_name

  tags = var.tags
}

# Link Private DNS Zone to VNet
resource "azurerm_private_dns_zone_virtual_network_link" "cognitive" {
  count = var.enable_private_endpoint ? 1 : 0

  name                  = "${var.name}-dns-link"
  resource_group_name   = var.resource_group_name
  private_dns_zone_name = azurerm_private_dns_zone.cognitive[0].name
  virtual_network_id    = var.vnet_id

  tags = var.tags
}

# Private Endpoint
resource "azurerm_private_endpoint" "speech" {
  count = var.enable_private_endpoint ? 1 : 0

  name                = "${var.name}-pe"
  location            = var.location
  resource_group_name = var.resource_group_name
  subnet_id           = var.private_endpoint_subnet_id

  private_service_connection {
    name                           = "${var.name}-psc"
    private_connection_resource_id = azurerm_cognitive_account.speech.id
    is_manual_connection           = false
    subresource_names              = ["account"]
  }

  private_dns_zone_group {
    name                 = "cognitive-dns-zone-group"
    private_dns_zone_ids = [azurerm_private_dns_zone.cognitive[0].id]
  }

  tags = var.tags
}
