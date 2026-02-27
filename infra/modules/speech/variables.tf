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

variable "network_acls" {
  description = "Network ACLs configuration for the Speech service"
  type = object({
    default_action = string
    ip_rules       = list(string)
    subnet_id      = string
  })
  default = null
}

variable "public_network_access_enabled" {
  description = "Whether public network access is enabled"
  type        = bool
  default     = true
}

variable "tags" {
  description = "Tags to apply to the resource"
  type        = map(string)
  default     = {}
}
