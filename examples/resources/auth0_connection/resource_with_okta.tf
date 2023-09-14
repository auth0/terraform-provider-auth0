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

    connection_settings {
      pkce = "auto"
    }

    attribute_map {
      mapping_mode   = "basic_profile"
      userinfo_scope = "openid email profile groups"
      attributes = jsonencode({
        "name" : "$${context.tokenset.name}",
        "email" : "$${context.tokenset.email}",
        "email_verified" : "$${context.tokenset.email_verified}",
        "nickname" : "$${context.tokenset.nickname}",
        "picture" : "$${context.tokenset.picture}",
        "given_name" : "$${context.tokenset.given_name}",
        "family_name" : "$${context.tokenset.family_name}"
      })
    }
  }
}
