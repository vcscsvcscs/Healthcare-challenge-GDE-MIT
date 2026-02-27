output "id" {
  description = "ACR ID"
  value       = azurerm_container_registry.acr.id
}

output "name" {
  description = "ACR name"
  value       = azurerm_container_registry.acr.name
}

output "login_server" {
  description = "ACR login server URL"
  value       = azurerm_container_registry.acr.login_server
}

output "admin_username" {
  description = "ACR admin username"
  value       = var.admin_enabled ? azurerm_container_registry.acr.admin_username : null
  sensitive   = true
}

output "admin_password" {
  description = "ACR admin password"
  value       = var.admin_enabled ? azurerm_container_registry.acr.admin_password : null
  sensitive   = true
}

output "private_ip_address" {
  description = "Private IP address of the ACR private endpoint"
  value       = var.enable_private_endpoint ? azurerm_private_endpoint.acr[0].private_service_connection[0].private_ip_address : null
}
