resource "azurerm_cognitive_account" "speech" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name
  kind                = "SpeechServices"
  sku_name            = var.sku_name

  custom_subdomain_name = var.custom_subdomain_name

  dynamic "network_acls" {
    for_each = var.network_acls != null ? [var.network_acls] : []
    content {
      default_action = network_acls.value.default_action
      ip_rules       = network_acls.value.ip_rules
      virtual_network_rules {
        subnet_id = network_acls.value.subnet_id
      }
    }
  }

  public_network_access_enabled = var.public_network_access_enabled

  tags = var.tags
}
