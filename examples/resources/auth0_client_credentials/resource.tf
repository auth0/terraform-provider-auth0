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

# Configuring a write-only client secret.
# NOTE: Write-only arguments are supported in Terraform 1.11 and later.
# The secret is never stored in Terraform state. Increment
# client_secret_wo_version to rotate the secret.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method    = "client_secret_post"
  client_secret_wo         = "LUFqPx+sRLjbL7peYRPFmFu-bbvE7u7og4YUNe_C345=683341"
  client_secret_wo_version = 1
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

# Configuring tls_client_auth as an authentication method with a PEM certificate.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "tls_client_auth"

  tls_client_auth {
    credentials {
      name            = "Testing Credentials 1"
      credential_type = "cert_subject_dn"
      pem             = <<EOF
-----BEGIN CERTIFICATE-----
MIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }
  }
}

# Configuring tls_client_auth as an authentication method with a subject_dn.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "tls_client_auth"

  tls_client_auth {
    credentials {
      name            = "Testing Credentials 1"
      credential_type = "cert_subject_dn"
      subject_dn      = "C=es\nST=Madrid\nL=Madrid\nO=Okta\nOU=DX-CDT\nCN=Developer Experience"
    }
  }
}

# Configuring self_signed_tls_client_auth as an authentication method.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "self_signed_tls_client_auth"

  self_signed_tls_client_auth {
    credentials {
      name            = "Testing Credentials 1"
      credential_type = "x509_cert"
      pem             = <<EOF
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

# Configuring token_vault_privileged_access.
# NOTE: This is an Early Access feature and requires the corresponding tenant
# entitlement. The credentials, ip_allowlist and grants blocks are always sent
# together, so all three are required.
resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  token_vault_privileged_access {
    credentials {
      name            = "Token Vault Credentials 1"
      credential_type = "public_key"
      algorithm       = "RS256"
      pem             = <<EOF
-----BEGIN CERTIFICATE-----
MIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }

    ip_allowlist = ["10.0.0.1", "192.168.1.0/24"]

    grants {
      connection = "google-oauth2"
      scopes     = ["https://www.googleapis.com/auth/calendar.readonly"]
    }
  }
}
