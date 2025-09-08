resource "auth0_resource_server" "my_resource_server" {
  name        = "Example Resource Server (Managed by Terraform)"
  identifier  = "https://api.example.com"
  signing_alg = "RS256"

  allow_offline_access                            = true
  token_lifetime                                  = 8600
  skip_consent_for_verifiable_first_party_clients = true
  consent_policy                                  = "transactional-authorization-with-mfa"
  token_encryption {
    format = "compact-nested-jwe"
    encryption_key {
      name      = "keyname"
      algorithm = "RSA-OAEP-256"
      pem       = <<EOF
-----BEGIN CERTIFICATE-----
MIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }
  }
  authorization_details {
    type = "payment"
  }
  authorization_details {
    type = "non-payment"
  }
  proof_of_possession {
    mechanism = "mtls"
    required  = true
  }
  subject_type_authorization {
    user {
      policy = "allow_all"
    }
    client {
      policy = "require_client_grant"
    }
  }
}
