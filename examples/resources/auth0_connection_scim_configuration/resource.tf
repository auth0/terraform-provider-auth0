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

resource "auth0_connection" "my_enterprise_connection_2" {
  name         = "my-enterprise-connection-2"
  display_name = "My Enterprise Connection 2"
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

# A resource for configuring an Auth0 Connection SCIM Configuration, using default values.
# Only one can be specified for a connection.
resource "auth0_connection_scim_configuration" "my_conn_scim_configuration_default" {
  connection_id = auth0_connection.my_enterprise_connection.id
}

# A resource for configuring an Auth0 Connection SCIM Configuration, specifying `user_id_attribute` and `mapping`.
# Only one can be specified for a connection.
resource "auth0_connection_scim_configuration" "my_conn_scim_configuration" {
  connection_id     = auth0_connection.my_enterprise_connection_2.id
  user_id_attribute = "attribute1"
  mapping {
    auth0 = "auth0_attribute1"
    scim  = "sacim_attribute1"
  }
  mapping {
    auth0 = "auth0_attribute2"
    scim  = "sacim_attribute2"
  }
}
