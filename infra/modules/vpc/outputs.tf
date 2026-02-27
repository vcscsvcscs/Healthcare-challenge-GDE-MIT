output "vnet_id" {
  description = "Virtual network ID"
  value       = azurerm_virtual_network.main.id
}

output "vnet_name" {
  description = "Virtual network name"
  value       = azurerm_virtual_network.main.name
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = azurerm_subnet.public[*].id
}

output "private_services_subnet_ids" {
  description = "Private services subnet IDs"
  value       = azurerm_subnet.private_services[*].id
}

output "private_db_subnet_ids" {
  description = "Private database subnet IDs"
  value       = azurerm_subnet.private_db[*].id
}

output "nat_gateway_ids" {
  description = "NAT Gateway IDs"
  value       = azurerm_nat_gateway.main[*].id
}

output "nat_public_ips" {
  description = "NAT Gateway public IP addresses"
  value       = azurerm_public_ip.nat[*].ip_address
}
