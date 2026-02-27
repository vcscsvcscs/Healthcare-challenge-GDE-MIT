output "id" {
  description = "The ID of the OpenAI service"
  value       = azurerm_cognitive_account.openai.id
}

output "endpoint" {
  description = "The endpoint of the OpenAI service"
  value       = azurerm_cognitive_account.openai.endpoint
}

output "primary_access_key" {
  description = "The primary access key for the OpenAI service"
  value       = azurerm_cognitive_account.openai.primary_access_key
  sensitive   = true
}

output "secondary_access_key" {
  description = "The secondary access key for the OpenAI service"
  value       = azurerm_cognitive_account.openai.secondary_access_key
  sensitive   = true
}

output "custom_subdomain" {
  description = "The custom subdomain of the OpenAI service"
  value       = azurerm_cognitive_account.openai.custom_subdomain_name
}

output "deployments" {
  description = "Map of deployment names to their IDs"
  value       = { for k, v in azurerm_cognitive_deployment.deployment : k => v.id }
}
