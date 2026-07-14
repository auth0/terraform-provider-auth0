# This is an example of an OIDC connection configured as a Cross-App Access (XAA)
# requesting application authorization server.
#
# The cross_app_access_requesting_app block is an Early Access feature and requires
# the `token_vault_xaa` flag to be enabled on your tenant. It is only supported on
# `oidc` and `okta` strategy connections.
#
# Note: Once configured, removing the block from your configuration is a no-op and will
# not disable the purpose on the connection. Set `active = false` explicitly to deactivate it.

resource "auth0_connection" "oidc_xaa" {
  name         = "oidc-xaa-connection"
  display_name = "OIDC XAA Connection"
  strategy     = "oidc"

  options {
    client_id              = "1234567"
    client_secret          = "1234567"
    type                   = "back_channel"
    issuer                 = "https://api.login.yahoo.com"
    jwks_uri               = "https://api.login.yahoo.com/openid/v1/certs"
    discovery_url          = "https://api.login.yahoo.com/.well-known/openid-configuration"
    token_endpoint         = "https://api.login.yahoo.com/oauth2/get_token"
    userinfo_endpoint      = "https://api.login.yahoo.com/openid/v1/userinfo"
    authorization_endpoint = "https://api.login.yahoo.com/oauth2/request_auth"
    scopes                 = ["openid", "email", "profile"]
  }

  cross_app_access_requesting_app {
    active = true
  }
}
