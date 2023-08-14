# This is an example of an Okta Workforce connection.

resource "auth0_connection" "okta" {
  name           = "okta-connection"
  display_name   = "Okta Workforce Connection"
  strategy       = "okta"
  show_as_button = false

  options {
    client_id                = "1234567"
    client_secret            = "1234567"
    domain                   = "example.okta.com"
    domain_aliases           = ["example.com"]
    issuer                   = "https://example.okta.com"
    jwks_uri                 = "https://example.okta.com/oauth2/v1/keys"
    token_endpoint           = "https://example.okta.com/oauth2/v1/token"
    userinfo_endpoint        = "https://example.okta.com/oauth2/v1/userinfo"
    authorization_endpoint   = "https://example.okta.com/oauth2/v1/authorize"
    scopes                   = ["openid", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]
    upstream_params = jsonencode({
      "screen_name" : {
        "alias" : "login_hint"
      }
    })
  }
}
