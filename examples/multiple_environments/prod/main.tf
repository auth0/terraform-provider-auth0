provider "auth0" {
  domain = "example-prod.us.auth0.com"
}

module "admin_console" {
  source = "../modules/admin_console"
}
