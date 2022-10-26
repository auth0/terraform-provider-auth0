# This is an example of an OIDC connection.

resource "auth0_connection" "oidc" {
  name           = "oidc-connection"
  display_name   = "OIDC Connection"
  strategy       = "oidc"
  show_as_button = false

  options {
    client_id     = "1234567"
    client_secret = "1234567"
    domain_aliases = [
      "example.com"
    ]
    tenant_domain = ""
    icon_url                 = "http://example.com/assets/logo.png"
    type                     = "front_channel"
    issuer                   = "https://www.paypalobjects.com"
    jwks_uri                 = "https://api.paypal.com/v1/oauth2/certs"
    discovery_url            = "https://www.paypalobjects.com/.well-known/openid-configuration"
    token_endpoint           = "https://api.paypal.com/v1/oauth2/token"
    userinfo_endpoint        = "https://api.paypal.com/v1/oauth2/token/userinfo"
    authorization_endpoint   = "https://www.paypal.com/signin/authorize"
    scopes                   = ["openid", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs = ["ethnicity","gender"]
  }
}
