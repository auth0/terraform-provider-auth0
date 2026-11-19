# This is an example of a SAML connection configured as a Cross-App Access (XAA)
# requesting application authorization server.
#
# The cross_app_access_requesting_app block is an Early Access (EA only) feature.
#
# Note: Once configured, removing the block from your configuration is a no-op and will
# not disable the purpose on the connection. Set `active = false` explicitly to deactivate it.

resource "auth0_connection" "saml_xaa" {
  name     = "saml-xaa-connection"
  strategy = "samlp"

  options {
    signing_cert     = "<signing-certificate>"
    sign_in_endpoint = "https://saml.provider/sign_in"

    # Credentials Auth0 uses to authenticate to the identity provider's token endpoint.
    client_id                       = "1234567"
    client_secret                   = "1234567"
    token_endpoint_auth_method      = "client_secret_post"
    token_endpoint_auth_signing_alg = "RS256"
    token_endpoint_jwtca_aud_format = "token_endpoint"

    # The metadata must contain a token_endpoint for the requesting app to be activated.
    oidc_metadata = jsonencode({
      issuer                                = "https://idp.example.com"
      authorization_endpoint                = "https://idp.example.com/authorize"
      token_endpoint                        = "https://idp.example.com/token"
      jwks_uri                              = "https://idp.example.com/jwks"
      id_token_signing_alg_values_supported = ["RS256"]
    })
  }

  cross_app_access_requesting_app {
    active = true
  }
}
