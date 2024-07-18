resource "auth0_connection" "my_enterprise_connection" {
  name         = "my-enterprise-connection"
  display_name = "My Enterprise Connection"
  strategy     = "okta"

  options {
    client_id              = "1234567"
    client_secret          = "1234567"
    issuer                 = "https://example.okta.com"
    jwks_uri               = "https://example.okta.com/oauth2/v1/keys"
    token_endpoint         = "https://example.okta.com/oauth2/v1/token"
    authorization_endpoint = "https://example.okta.com/oauth2/v1/authorize"
  }
}

resource "auth0_connection_scim_configuration" "my_conn_scim_configuration" {
  connection_id = auth0_connection.my_enterprise_connection.id
}

# A data source for an Auth0 Connection SCIM Configuration.
data "auth0_connection_scim_configuration" "my_conn_scim_configuration_data" {
  connection_id = auth0_connection_scim_configuration.my_conn_scim_configuration.id
}

