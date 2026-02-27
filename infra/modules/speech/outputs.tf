output "id" {
  description = "The ID of the Cognitive Services Speech account"
  value       = azurerm_cognitive_account.speech.id
}

output "endpoint" {
  description = "The endpoint URL for the Speech service"
  value       = azurerm_cognitive_account.speech.endpoint
}

output "primary_access_key" {
  description = "The primary access key for the Speech service"
  value       = azurerm_cognitive_account.speech.primary_access_key
  sensitive   = true
}

output "secondary_access_key" {
  description = "The secondary access key for the Speech service"
  value       = azurerm_cognitive_account.speech.secondary_access_key
  sensitive   = true
}

output "name" {
  description = "The name of the Speech service"
  value       = azurerm_cognitive_account.speech.name
}

output "private_endpoint_id" {
  description = "The ID of the private endpoint"
  value       = var.enable_private_endpoint ? azurerm_private_endpoint.speech[0].id : null
}

output "private_ip_address" {
  description = "The private IP address of the Speech service"
  value       = var.enable_private_endpoint ? azurerm_private_endpoint.speech[0].private_service_connection[0].private_ip_address : null
}
