# This is an example of an Okta Workforce connection.

resource "auth0_connection_okta" "okta" {
  name           = "okta-connection"
  display_name   = "Okta Workforce Connection"
  strategy       = "okta"
  show_as_button = false

  client_id                = "1234567"
  client_secret            = "1234567"
  domain                   = "example.okta.com"
  domain_aliases           = ["example.com"]
  issuer                   = "https://example.okta.com"
  jwks_uri                 = "https://example.okta.com/v1/oauth2/certs"
  token_endpoint           = "https://example.okta.com/v1/oauth2/token"
  userinfo_endpoint        = "https://example.okta.com/v1/oauth2/token/userinfo"
  authorization_endpoint   = "https://example.okta.com/signin/authorize"
  scopes                   = ["openid", "email"]
  set_user_root_attributes = "on_first_login"
  non_persistent_attrs     = ["ethnicity", "gender"]
  upstream_params = jsonencode({
    "screen_name" : {
      "alias" : "login_hint"
    }
  })
}
