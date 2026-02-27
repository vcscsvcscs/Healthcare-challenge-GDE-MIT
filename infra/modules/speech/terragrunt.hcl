terraform {
  source = "."
}

include "root" {
  path = find_in_parent_folders()
}

inputs = {
  name                = "speech-${get_env("ENVIRONMENT", "dev")}"
  location            = "westeurope"  # Central Europe region
  resource_group_name = dependency.resource_group.outputs.name
  sku_name            = "S0"
  
  custom_subdomain_name = "speech-${get_env("ENVIRONMENT", "dev")}-subdomain"
  
  public_network_access_enabled = true
  
  tags = {
    Environment = get_env("ENVIRONMENT", "dev")
    ManagedBy   = "Terragrunt"
    Region      = "Central Europe"
  }
}

# Uncomment if you have a resource group module
# dependency "resource_group" {
#   config_path = "../resource-group"
# }
