variable "domain" {
  description = "The name of the custom domain to provision"
  type        = string
}

variable "managed_zone_name" {
  description = "The name of the Cloud DNS managed zone to create DNS records in"
  type        = string
}
