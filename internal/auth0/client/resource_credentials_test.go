package client_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}
`

const testAccThrowErrorWhenPrivateKeyJWTNoCredentials = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"
}
`

const testAccThrowErrorWhenPrivateKeyJWTWrongCredentialType = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccCreateOneClientCredentialUsingPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem                    = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddAnotherClientCredentialUsingPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem                    = <<EOF
%s
EOF
		}
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			parse_expiry_from_cert = false
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddUpdateClientCredentialsPrivateKeyJWTExpiresAt = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			expires_at             = "2050-05-13T09:33:13.000Z" # This takes precedence.
			pem                    = <<EOF
%s
EOF
		}
		credentials {
			name                   = "Testing Credentials 2"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = false
			expires_at             = "2025-05-13T09:33:13.000Z"
			pem                    = <<EOF
%s
EOF
		}
	}
}
`

const testAccRemoveOneClientCredentialPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccSwitchToClientSecretBasicFromPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "client_secret_basic"
}
`

const testAccSwitchBackToUsePrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

func TestAccClientAuthenticationMethodsPrivateKeyJWT(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	credsCert2, err := os.ReadFile("./../../../test/data/creds-cert-2.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenPrivateKeyJWTNoCredentials, t.Name()),
				ExpectError: regexp.MustCompile("Client Credentials Missing"),
			},
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenPrivateKeyJWTWrongCredentialType, t.Name()),
				ExpectError: regexp.MustCompile("expected .*credential_type to be one of .*public_key.* got cert_subject_dn"),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccCreateOneClientCredentialUsingPrivateKeyJWT, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddAnotherClientCredentialUsingPrivateKeyJWT, t.Name()), credsCert1, credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddUpdateClientCredentialsPrivateKeyJWTExpiresAt, t.Name()), credsCert1, credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2050-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.1.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccRemoveOneClientCredentialPrivateKeyJWT, t.Name()), credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSwitchToClientSecretBasicFromPrivateKeyJWT, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "client_secret_basic"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccSwitchBackToUsePrivateKeyJWT, t.Name()), credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "private_key_jwt.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient, t.Name()),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
				),
			},
		},
	})
}

const testAccThrowErrorWhenTLSClientAuthNoCredentials = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"
}
`

const testAccThrowErrorWhenTLSClientAuthWrongCredentialType = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"

	tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccThrowErrorWhenTLSClientAuthPEMSubjectDN = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"

	tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			subject_dn      = "C=es\nST=Madrid\nL=Madrid\nO=Okta\nOU=DX-CDT\nCN=Developer Experience"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccCreateClientCredentialUsingTLSClientAuthPEM = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"

	tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccCreateClientCredentialUsingTLSClientAuthSubjectDN = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"

	tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			subject_dn      = "C=es\nST=Madrid\nL=Madrid\nO=Okta\nOU=DX-CDT\nCN=Developer Experience"
		}
	}
}
`

const testAccSwitchToClientSecretBasicFromTLSClientAuth = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "client_secret_basic"
}
`

const testAccSwitchBackToUseTLSClientAuthSubjectDN = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "tls_client_auth"

	tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			subject_dn      = "C=es\nST=Madrid\nL=Madrid\nO=Okta\nOU=DX-CDT\nCN=Developer Experience"
		}
	}
}
`

func TestAccClientAuthenticationMethodsTLSClientAuth(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenTLSClientAuthNoCredentials, t.Name()),
				ExpectError: regexp.MustCompile("Client Credentials Missing"),
			},
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenTLSClientAuthWrongCredentialType, t.Name()),
				ExpectError: regexp.MustCompile("expected .*credential_type to be one of .*cert_subject_dn.* got public_key"),
			},
			{
				Config:      fmt.Sprintf(acctest.ParseTestName(testAccThrowErrorWhenTLSClientAuthPEMSubjectDN, t.Name()), credsCert1),
				ExpectError: regexp.MustCompile("Client Credentials Invalid"),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccCreateClientCredentialUsingTLSClientAuthPEM, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "tls_client_auth"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.credential_type", "cert_subject_dn"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.subject_dn"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateClientCredentialUsingTLSClientAuthSubjectDN, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "tls_client_auth"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.credential_type", "cert_subject_dn"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.subject_dn"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSwitchToClientSecretBasicFromTLSClientAuth, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "client_secret_basic"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSwitchBackToUseTLSClientAuthSubjectDN, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "tls_client_auth"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.credential_type", "cert_subject_dn"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.subject_dn"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "tls_client_auth.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient, t.Name()),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
				),
			},
		},
	})
}

const testAccThrowErrorWhenSelfSignedTLSClientAuthNoCredentials = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "self_signed_tls_client_auth"
}
`

const testAccThrowErrorWhenSelfSignedTLSClientAuthWrongCredentialType = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	authentication_method = "self_signed_tls_client_auth"

	self_signed_tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccCreateClientCredentialUsingSelfSignedTLSClientAuth = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "self_signed_tls_client_auth"

	self_signed_tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "x509_cert"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccSwitchToClientSecretBasicFromSelfSignedTLSClientAuth = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "client_secret_basic"
}
`

const testAccSwitchBackToUseSelfSignedTLSClientAuth = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "self_signed_tls_client_auth"

	self_signed_tls_client_auth {
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "x509_cert"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

func TestAccClientAuthenticationMethodsSelfSignedTLSClientAuth(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenSelfSignedTLSClientAuthNoCredentials, t.Name()),
				ExpectError: regexp.MustCompile("Client Credentials Missing"),
			},
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenSelfSignedTLSClientAuthWrongCredentialType, t.Name()),
				ExpectError: regexp.MustCompile("expected .*credential_type to be one of .*x509_cert.* got public_key"),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccCreateClientCredentialUsingSelfSignedTLSClientAuth, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "self_signed_tls_client_auth"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.credential_type", "x509_cert"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.expires_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSwitchToClientSecretBasicFromSelfSignedTLSClientAuth, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "client_secret_basic"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccSwitchBackToUseSelfSignedTLSClientAuth, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "self_signed_tls_client_auth"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.credential_type", "x509_cert"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "self_signed_tls_client_auth.0.credentials.0.expires_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient, t.Name()),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
				),
			},
		},
	})
}

const testAccAllowUpdatingTheClientSecret = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials With Secret Rotation"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "client_secret_post"
	client_secret         = "LUFqPx+sRLjbL7peYRPFmFu-bbvE7u7og4YUNe_C345=683341"
}
`

func TestAccAllowUpdatingTheClientSecret(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// Only run with recorded HTTP requests, because
		// the http recorder redacts the client secret.
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccAllowUpdatingTheClientSecret,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", "Acceptance Test - Client Credentials With Secret Rotation"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "client_secret_post"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "client_secret", "LUFqPx+sRLjbL7peYRPFmFu-bbvE7u7og4YUNe_C345=683341"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "private_key_jwt.#", "0"),
				),
			},
		},
	})
}

const testAccThrowErrorWhenSignedRequestObjectNoCredentials = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id
	signed_request_object {
		required = true
	}
}
`

const testAccThrowErrorWhenSignedRequestObjectWrongCredentialType = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = true
		credentials {
			name            = "Testing Credentials 1"
			credential_type = "cert_subject_dn"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccCreateOneClientCredentialUsingSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = false
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem                    = <<EOF
%s
EOF
		}
	}
}
`

const testAccChangeRequiredClientCredentialUsingSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = true
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem                    = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddAnotherClientCredentialUsingSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = true
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem                    = <<EOF
%s
EOF
		}
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			parse_expiry_from_cert = false
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddUpdateClientCredentialsSignedRequestObjectExpiresAt = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = true
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			expires_at             = "2050-05-13T09:33:13.000Z" # This takes precedence.
			pem                    = <<EOF
%s
EOF
		}
		credentials {
			name                   = "Testing Credentials 2"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = false
			expires_at             = "2025-05-13T09:33:13.000Z"
			pem                    = <<EOF
%s
EOF
		}
	}
}
`

const testAccRemoveOneClientCredentialSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		required = true
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

const testAccSwitchToClientSecretBasicFromSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id
	authentication_method = "client_secret_basic"
}
`

const testAccSwitchBackToUseSignedRequestObject = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test" {
	client_id             = auth0_client.my_client.id

	signed_request_object {
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem             = <<EOF
%s
EOF
		}
	}
}
`

func TestAccClientAuthenticationMethodsSignedRequestObject(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	credsCert2, err := os.ReadFile("./../../../test/data/creds-cert-2.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenSignedRequestObjectNoCredentials, t.Name()),
				ExpectError: regexp.MustCompile("Insufficient credentials blocks"),
			},
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenSignedRequestObjectWrongCredentialType, t.Name()),
				ExpectError: regexp.MustCompile("expected signed_request_object.0.credentials.0.credential_type to be one of .*public_key.* got cert_subject_dn"),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccCreateOneClientCredentialUsingSignedRequestObject, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.required", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccChangeRequiredClientCredentialUsingSignedRequestObject, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.required", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddAnotherClientCredentialUsingSignedRequestObject, t.Name()), credsCert1, credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2033-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddUpdateClientCredentialsSignedRequestObjectExpiresAt, t.Name()), credsCert1, credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "true"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2050-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.1.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.1.updated_at"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccRemoveOneClientCredentialSignedRequestObject, t.Name()), credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccSwitchToClientSecretBasicFromSignedRequestObject, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "authentication_method", "client_secret_basic"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "tls_client_auth.#", "0"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "self_signed_tls_client_auth.#", "0"),
				),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccSwitchBackToUseSignedRequestObject, t.Name()), credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.name", "Testing Credentials 2"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.parse_expiry_from_cert", "false"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test", "signed_request_object.0.credentials.0.expires_at", "2025-05-13T09:33:13.000Z"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test", "signed_request_object.0.credentials.0.updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient, t.Name()),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance Test - Client Credentials - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.my_client", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "app_type", "non_interactive"),
				),
			},
		},
	})
}

const testAccImportClientWithSecretPost = `
resource "auth0_client" "my_test_client_secret" {
	name     = "Acceptance Test - Client Credentials Import"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "test_simple_client" {
	client_id             = auth0_client.my_test_client_secret.id
	authentication_method = "client_secret_post"
}
`

const testAccImportClientWithPrivateKeyJWT = `
resource "auth0_client" "my_test_client_jwt_ca" {
	name 	 = "Acceptance Test - Client Credentials Import"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}

resource "auth0_client_credentials" "test_jwt_ca_client" {
	client_id             = auth0_client.my_test_client_jwt_ca.id
	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			credential_type = "public_key"
			algorithm       = "RS256"
			pem = <<EOF
%s
EOF
		}
	}
}
`

func TestAccClientCredentialsImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	credsCert, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             testAccImportClientWithSecretPost,
				ResourceName:       "auth0_client.my_test_client_secret",
				ImportState:        true,
				ImportStateId:      "Bjnm4jQ66Kb5Ug33eSBDxHsW6teU7SE1",
				ImportStatePersist: true,
			},
			{
				Config:             testAccImportClientWithSecretPost,
				ResourceName:       "auth0_client_credentials.test_simple_client",
				ImportState:        true,
				ImportStateId:      "Bjnm4jQ66Kb5Ug33eSBDxHsW6teU7SE1",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: testAccImportClientWithSecretPost,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_credentials.test_simple_client", "authentication_method", "client_secret_post"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test_simple_client", "client_secret"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_simple_client", "private_key_jwt.#", "0"),
				),
			},
			{
				Config:             testAccImportClientWithPrivateKeyJWT,
				ResourceName:       "auth0_client.my_test_client_jwt_ca",
				ImportState:        true,
				ImportStateId:      "zm5DPtaaSqenbpEX36nNLbmUQ2rW81Mu",
				ImportStatePersist: true,
			},
			{
				Config:             fmt.Sprintf(testAccImportClientWithPrivateKeyJWT, credsCert),
				ResourceName:       "auth0_client_credentials.test_jwt_ca_client",
				ImportState:        true,
				ImportStateId:      "zm5DPtaaSqenbpEX36nNLbmUQ2rW81Mu",
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_client_credentials.test_jwt_ca_client", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Config: fmt.Sprintf(testAccImportClientWithPrivateKeyJWT, credsCert),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "authentication_method", "private_key_jwt"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "client_secret", ""),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.algorithm", "RS256"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.pem"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.key_id"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.created_at"),
					resource.TestCheckResourceAttrSet("auth0_client_credentials.test_jwt_ca_client", "private_key_jwt.0.credentials.0.updated_at"),
				),
			},
		},
	})
}
