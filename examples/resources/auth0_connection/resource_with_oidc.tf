# This is an example of an OIDC connection.

resource "auth0_connection" "oidc" {
  name           = "oidc-connection"
  display_name   = "OIDC Connection"
  strategy       = "oidc"
  show_as_button = false

  options {
    client_id                = "1234567"
    client_secret            = "1234567"
    domain_aliases           = ["example.com"]
    tenant_domain            = ""
    icon_url                 = "https://example.com/assets/logo.png"
    type                     = "back_channel"
    send_back_channel_nonce  = true
    issuer                   = "https://www.paypalobjects.com"
    jwks_uri                 = "https://api.paypal.com/v1/oauth2/certs"
    discovery_url            = "https://www.paypalobjects.com/.well-known/openid-configuration"
    token_endpoint           = "https://api.paypal.com/v1/oauth2/token"
    userinfo_endpoint        = "https://api.paypal.com/v1/oauth2/token/userinfo"
    authorization_endpoint   = "https://www.paypal.com/signin/authorize"
    scopes                   = ["openid", "email"]
    set_user_root_attributes = "on_first_login"
    non_persistent_attrs     = ["ethnicity", "gender"]

    connection_settings {
      pkce = "auto"
    }

    attribute_map {
      mapping_mode   = "use_map"
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
