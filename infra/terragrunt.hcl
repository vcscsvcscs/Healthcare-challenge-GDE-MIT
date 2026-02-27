# Terragrunt config – use local Terraform in this directory.
# Replace source with a module path (e.g. "./modules/networking") when you add modules.

terraform {
  source = "."
}

# inputs = {}
# Add inputs when you have variables to pass, e.g.:
# inputs = {
#   environment = "dev"
#   location     = "eastus"
# }
