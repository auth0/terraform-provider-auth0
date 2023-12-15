---
page_title: Zero downtime client credentials rotation
description: |-
  Achieve zero downtime client credentials rotation with the Auth0 Terraform provider.
---

# Achieving zero downtime client credentials rotation

In this guide we'll show how to rotate a client's credentials to eliminate downtime for the impacted system when using
Private Key JWT credentials.

## Rotating Private Key JWT credentials

1. Generate a new Private Key JWT credential on behalf of the system associated with the client application record.
2. Add the newly generated credential to the system configuration as the next credential in the list or in the
respective entry if separate configuration entries are used.
3. Attach the Private Key JWT credential to the client application record using Terraform and run `terraform apply`:

```terraform
resource "auth0_client" "my_client" {
	name     = "My client that needs the credentials rotated"
	app_type = "non_interactive"

  jwt_configuration {
    alg = "RS256"
  }
}

resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "private_key_jwt"
  
  private_key_jwt {
    credentials {
      name                   = "Current Credential"
      credential_type        = "public_key"
      algorithm              = "RS256"
      pem                    = <<EOF
-----BEGIN CERTIFICATE-----
MIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }

    credentials {
      name                   = "Next Credential"
      credential_type        = "public_key"
      algorithm              = "RS256"
      pem                    = <<EOF
-----BEGIN CERTIFICATE-----
BBBIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }
  }
}
```

4. Remove the old Private Key JWT credential on the client application record using Terraform and run `terraform apply`:

```terraform
resource "auth0_client" "my_client" {
	name     = "My client that needs the credentials rotated"
	app_type = "non_interactive"

  jwt_configuration {
    alg = "RS256"
  }
}

resource "auth0_client_credentials" "test" {
  client_id = auth0_client.my_client.id

  authentication_method = "private_key_jwt"
  
  private_key_jwt {
    credentials {
      name                   = "Current Credential" # Next becomes current.
      credential_type        = "public_key"
      algorithm              = "RS256"
      pem                    = <<EOF
-----BEGIN CERTIFICATE-----
BBBIIFWDCCA0ACCQDXqpBo3R...G9w0BAQsFADBuMQswCQYDVQQGEwJl
-----END CERTIFICATE-----
EOF
    }
  }
}
```
