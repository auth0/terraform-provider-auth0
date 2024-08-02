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

const testAccThrowErrorWhenPrivateKeyJWT = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"
}

resource "auth0_client_credentials" "my_client_credentials" {
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"
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
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem = <<EOF
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
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			pem = <<EOF
%s
EOF
		}
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			parse_expiry_from_cert = false
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem = <<EOF
%s
EOF
		}
	}
}
`

const testAccAddUpdateClientCredentialsExpiresAt = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
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
			name                   = "Testing Credentials 1"
			credential_type        = "public_key"
			algorithm              = "RS256"
			parse_expiry_from_cert = true
			expires_at             = "2050-05-13T09:33:13.000Z" # This takes precedence.
			pem = <<EOF
%s
EOF
		}
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			parse_expiry_from_cert = false
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem = <<EOF
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
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem = <<EOF
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
	client_id = auth0_client.my_client.id

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
	client_id = auth0_client.my_client.id

	authentication_method = "private_key_jwt"

	private_key_jwt {
		credentials {
			name            = "Testing Credentials 2"
			credential_type = "public_key"
			algorithm       = "RS256"
			expires_at      = "2025-05-13T09:33:13.000Z"
			pem = <<EOF
%s
EOF
		}
	}
}
`

const testAccDeletingTheResourceSetsTheTokenEndpointAuthMethodToADefaultOnTheClient = `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Client Credentials - {{.testName}}"
	app_type = "non_interactive"

	jwt_configuration {
		alg = "RS256"
	}
}
`

func TestAccClientAuthenticationMethods(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	credsCert2, err := os.ReadFile("./../../../test/data/creds-cert-2.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccThrowErrorWhenPrivateKeyJWT, t.Name()),
				ExpectError: regexp.MustCompile("Client Credentials Missing"),
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
				Config: fmt.Sprintf(acctest.ParseTestName(testAccAddUpdateClientCredentialsExpiresAt, t.Name()), credsCert1, credsCert2),
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
	client_id = auth0_client.my_client.id

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

const testAccImportClientWithSecretPost = `
resource "auth0_client" "my_test_client_secret" {
	name 	                   = "Acceptance Test - Client Credentials Import"
	app_type                   = "non_interactive"
}

resource "auth0_client_credentials" "test_simple_client" {
	client_id = auth0_client.my_test_client_secret.id

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
	client_id = auth0_client.my_test_client_jwt_ca.id

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
