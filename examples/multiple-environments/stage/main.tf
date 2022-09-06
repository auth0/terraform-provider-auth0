provider "auth0" {
  domain = "example-stage.us.auth0.com"
}

module "admin_console" {
  source = "../modules/admin_console"
}
