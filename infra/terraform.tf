terraform {
  backend "azurerm" {
    resource_group_name  = "tfstate-rg"
    storage_account_name = "tfstate[unique-id]"
    container_name       = "tfstate"
    key                  = "terraform.tfstate"
  }
}
