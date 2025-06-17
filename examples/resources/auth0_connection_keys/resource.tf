resource "auth0_client" "my_client" {
  name = "My-Auth0-Client"
}

resource "auth0_connection" "oidc" {
    name     = "OIDC-Connection"
    strategy = "oidc"
    options {
        client_id                     = auth0_client.my_client.id
        scopes                        = ["ext_nested_groups","openid"]
        issuer                        = "https://example.com"
        authorization_endpoint        = "https://example.com"
        jwks_uri                      = "https://example.com/jwks"
        type                          = "front_channel"
        discovery_url                 = "https://www.paypalobjects.com/.well-known/openid-configuration"
        token_endpoint_auth_method    = "private_key_jwt"
        token_endpoint_auth_signing_alg = "RS256"
    }
}

# Resource used to rotate the keys for above OIDC connection
resource "auth0_connection_keys" "my_keys"{
    connection_id = auth0_connection.oidc.id
}
