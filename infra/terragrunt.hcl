# Terragrunt config – use local Terraform in this directory.
# Replace source with a module path (e.g. "./modules/networking") when you add modules.

terraform {
  source = "."
  
  # Configure backend from backend.tfvars file (only for init)
  extra_arguments "backend" {
    commands = [
      "init"
    ]
    
    arguments = [
      "-backend-config=backend.tfvars"
    ]
  }
}

inputs = {
  environment         = get_env("ENVIRONMENT", "dev")
  location            = "swedencentral"
  resource_group_name = "Solo-1"
}
