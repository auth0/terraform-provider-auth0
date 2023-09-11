resource "auth0_client" "my_client" {
  name                                = "Application - Acceptance Test"
  description                         = "Test Applications Long Description"
  app_type                            = "non_interactive"
  custom_login_page_on                = true
  is_first_party                      = true
  is_token_endpoint_ip_header_trusted = true
  oidc_conformant                     = false
  callbacks                           = ["https://example.com/callback"]
  allowed_origins                     = ["https://example.com"]
  allowed_logout_urls                 = ["https://example.com"]
  web_origins                         = ["https://example.com"]
  grant_types = [
    "authorization_code",
    "http://auth0.com/oauth/grant-type/password-realm",
    "implicit",
    "password",
    "refresh_token"
  ]
  client_metadata = {
    foo = "zoo"
  }

  jwt_configuration {
    lifetime_in_seconds = 300
    secret_encoded      = true
    alg                 = "RS256"
    scopes = {
      foo = "bar"
    }
  }

  refresh_token {
    leeway          = 0
    token_lifetime  = 2592000
    rotation_type   = "rotating"
    expiration_type = "expiring"
  }

  mobile {
    ios {
      team_id               = "9JA89QQLNQ"
      app_bundle_identifier = "com.my.bundle.id"
    }
  }

  addons {
    samlp {
      audience = "https://example.com/saml"
      issuer   = "https://example.com"
      mappings = {
        email = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
        name  = "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"
      }
      create_upn_claim                   = false
      passthrough_claims_with_no_mapping = false
      map_unknown_claims_as_is           = false
      map_identities                     = false
      name_identifier_format             = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"
      name_identifier_probes = [
        "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"
      ]
      signing_cert = "-----BEGIN PUBLIC KEY-----\nMIGf...bpP/t3\n+JGNGIRMj1hF1rnb6QIDAQAB\n-----END PUBLIC KEY-----\n"
    }
  }
}
