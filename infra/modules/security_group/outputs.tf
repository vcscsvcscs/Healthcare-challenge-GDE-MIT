output "nsg_id" {
  description = "Network security group ID"
  value       = azurerm_network_security_group.main.id
}

output "nsg_name" {
  description = "Network security group name"
  value       = azurerm_network_security_group.main.name
}
