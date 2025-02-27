# Example:
resource "auth0_flow_vault_connection" "my_connection" {
  app_id = "AUTH0"
  name   = "Auth0 M2M Connection"
  setup = {
    client_id     = "******************"
    client_secret = "*********************************"
    domain        = "*****************************"
    type          = "OAUTH_APP"
  }
}
