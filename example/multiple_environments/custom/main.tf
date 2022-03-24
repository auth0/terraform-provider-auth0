variable "auth0_domain" {
  type    = string
  default = null # Leave empty to use AUTH0_DOMAIN from the environment
}

provider "auth0" {
  domain = var.auth0_domain
}

module "admin_console" {
  source = "../modules/admin_console"
}
