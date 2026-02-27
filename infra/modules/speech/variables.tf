variable "name" {
  description = "Name of the Cognitive Services Speech account"
  type        = string
}

variable "location" {
  description = "Azure region where the resource will be created"
  type        = string
}

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

variable "sku_name" {
  description = "SKU name for the Speech service (F0 for free tier, S0 for standard)"
  type        = string
  default     = "S0"

  validation {
    condition     = contains(["F0", "S0"], var.sku_name)
    error_message = "SKU must be either F0 (free) or S0 (standard)."
  }
}

variable "custom_subdomain_name" {
  description = "Custom subdomain name for the Speech service endpoint"
  type        = string
  default     = null
}

variable "allowed_subnet_ids" {
  description = "List of subnet IDs allowed to access the Speech service"
  type        = list(string)
  default     = null
}

variable "public_network_access_enabled" {
  description = "Whether public network access is enabled"
  type        = bool
  default     = false
}

variable "enable_private_endpoint" {
  description = "Enable private endpoint for the Speech service"
  type        = bool
  default     = true
}

variable "private_endpoint_subnet_id" {
  description = "Subnet ID for the private endpoint"
  type        = string
  default     = null
}

variable "vnet_id" {
  description = "Virtual Network ID for private DNS zone link"
  type        = string
  default     = null
}

variable "tags" {
  description = "Tags to apply to the resource"
  type        = map(string)
  default     = {}
}
