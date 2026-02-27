terraform {
  required_version = ">= 1.0"

  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }

  backend "azurerm" {
    # Backend configuration will be provided via backend config file or CLI
    # resource_group_name  = ""
    # storage_account_name = ""
    # container_name       = ""
    # key                  = ""
  }
}

provider "azurerm" {
  features {}
  subscription_id            = "61c53454-ceb0-49ba-bc5a-6178761ee50d"
  skip_provider_registration = true
}
