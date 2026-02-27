resource "azurerm_network_security_group" "main" {
  name                = var.name
  location            = var.location
  resource_group_name = var.resource_group_name

  tags = var.tags
}

# Dynamic security rules
resource "azurerm_network_security_rule" "rules" {
  for_each = { for idx, rule in var.security_rules : rule.name => rule }

  name                        = each.value.name
  priority                    = each.value.priority
  direction                   = each.value.direction
  access                      = each.value.access
  protocol                    = each.value.protocol
  source_port_range           = each.value.source_port_range
  destination_port_range      = each.value.destination_port_range
  source_address_prefix       = each.value.source_address_prefix
  destination_address_prefix  = each.value.destination_address_prefix
  resource_group_name         = var.resource_group_name
  network_security_group_name = azurerm_network_security_group.main.name
}

# Subnet association
resource "azurerm_subnet_network_security_group_association" "main" {
  count = var.subnet_id != null ? 1 : 0

  subnet_id                 = var.subnet_id
  network_security_group_id = azurerm_network_security_group.main.id
}
