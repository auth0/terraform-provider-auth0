resource "auth0_client" "my_client" {
  name     = "Application - Acceptance Test"
  app_type = "non_interactive"

  jwt_configuration {
    alg = "RS256"
  }
}

# Configuring client_secret_post as an authentication method.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "client_secret_post"
}

# Configuring client_secret_basic as an authentication method.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "client_secret_basic"
}

# Configuring none as an authentication method.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "none"
}

# Configuring private_key_jwt as an authentication method.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "private_key_jwt"

  private_key_jwt {
    credentials {
      name                   = "Testing Credentials 1"
      credential_type        = "public_key"
      algorithm              = "RS256"
      parse_expiry_from_cert = true
      pem                    = <<EOF
-----BEGIN CERTIFICATE-----
MIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }
  }
}

# Configuring the client_secret.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "client_secret_basic"
  client_secret         = "LUFqPx+sRLjbL7peYRPFmFu-bbvE7u7og4YUNe_C345=683341"
}
