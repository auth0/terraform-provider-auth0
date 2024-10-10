package resourceserver_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccResourceServerConfigEmpty = `
resource "auth0_resource_server" "my_resource_server" {
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
}
`

const testAccResourceServerConfigCreate = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = true
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
	consent_policy                                  = "null"
	authorization_details = [
	]
	token_encryption = {
	}
	proof_of_possession = {
	}
}
`

const testAccResourceServerConfigUpdate = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false # <--- set to false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
	consent_policy                                  = "transactional-authorization-with-mfa"
	authorization_details = [
		{
			type = "payment"
		}, {
			type = "not-payment"
		}
	]
	token_encryption = {
		format = "compact-nested-jwe"
		encryption_key = {
			name      = "encryptkey"
			algorithm = "RSA-OAEP-256"
			pem       = <<EOF
%s
EOF
		}
	}
	proof_of_possession = {
		mechanism = "mtls"
		required = true
	}
}
`

const testAccResourceServerInvalidTokenEncryption = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false # <--- set to false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
	consent_policy                                  = "transactional-authorization-with-mfa"
	token_encryption = {
		encryption_key = {
			name      = "encryptkey"
			algorithm = "RSA-OAEP-256"
			pem       = <<EOF
%s
EOF
		}
	}
}
`

const testAccResourceServerInvalidProofOfPossession = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false # <--- set to false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
	consent_policy                                  = "transactional-authorization-with-mfa"
	proof_of_possession = {
		mechanism = "mtls"
	}
}
`

const testAccResourceServerConfigUpdateWithMissingAttributes = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false # <--- set to false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
}
`

const testAccResourceServerConfigUpdateWithAccessToken = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false
	token_dialect                                   = "access_token" # <--- set to access_token
}
`

const testAccResourceServerConfigUpdateWithAccessTokenAuthz = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = true # <--- set to true
	token_dialect                                   = "access_token_authz" # <--- set to access_token_authz
}
`

const testAccResourceServerConfigUpdateWithRFC9068Profile = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = false # <--- set to false
	token_dialect                                   = "rfc9068_profile" # <--- set to rfc9068_profile
}
`

const testAccResourceServerConfigUpdateWithRFC9068ProfileAuthz = `
resource "auth0_resource_server" "my_resource_server" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = false
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = true # <--- set to true
	token_dialect                                   = "rfc9068_profile_authz" # <--- set to rfc9068_profile_authz
}
`

const testAccResourceServerConfigEmptyAgain = `
resource "auth0_resource_server" "my_resource_server" {
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	name       = "Acceptance Test - {{.testName}}"
}
`

func TestAccFrameworkResourceServer(t *testing.T) {
	credsCert1, err := os.ReadFile("../../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "name"),
					resource.TestCheckResourceAttrSet("auth0_resource_server.my_resource_server", "signing_alg"),
					resource.TestCheckResourceAttrSet("auth0_resource_server.my_resource_server", "token_lifetime_for_web"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "consent_policy", "null"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "authorization_details.#", "0"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "token_encryption.format"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "token_encryption.encryption_key"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "proof_of_possession.mechanism"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "proof_of_possession.required"),
				),
			},
			{
				Config:      fmt.Sprintf(acctest.ParseTestName(testAccResourceServerInvalidTokenEncryption, t.Name()), credsCert1),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			{
				Config:      acctest.ParseTestName(testAccResourceServerInvalidProofOfPossession, t.Name()),
				ExpectError: regexp.MustCompile(`Invalid Attribute Combination`),
			},
			{
				Config: fmt.Sprintf(acctest.ParseTestName(testAccResourceServerConfigUpdate, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "consent_policy", "transactional-authorization-with-mfa"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "authorization_details.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "authorization_details.0.type", "payment"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "authorization_details.1.type", "not-payment"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_encryption.format", "compact-nested-jwe"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_encryption.encryption_key.name", "encryptkey"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_encryption.encryption_key.algorithm", "RSA-OAEP-256"),
					resource.TestCheckResourceAttrSet("auth0_resource_server.my_resource_server", "token_encryption.encryption_key.pem"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "proof_of_possession.mechanism", "mtls"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "proof_of_possession.required", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigUpdateWithMissingAttributes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "false"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "consent_policy"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "authorization_details"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "token_encryption"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.my_resource_server", "proof_of_possession"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigUpdateWithAccessToken, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_dialect", "access_token"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigUpdateWithAccessTokenAuthz, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_dialect", "access_token_authz"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigUpdateWithRFC9068Profile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_dialect", "rfc9068_profile"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigUpdateWithRFC9068ProfileAuthz, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_dialect", "rfc9068_profile_authz"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerConfigEmptyAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_dialect", "rfc9068_profile_authz"),
				),
			},
		},
	})
}

const testAccAuth0ManagementAPIResourceImport = `
resource "auth0_resource_server" "auth0" {
	name                                            = "Auth0 Management API"
	identifier                                      = "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"
	token_lifetime                                  = 86400
	skip_consent_for_verifiable_first_party_clients = false
}
`

func TestAccFrameworkResourceServerAuth0APIManagementImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             testAccAuth0ManagementAPIResourceImport,
				ResourceName:       "auth0_resource_server.auth0",
				ImportState:        true,
				ImportStateId:      "xxxxxxxxxxxxxxxxxxxx",
				ImportStatePersist: true,
			},
			{
				Config: testAccAuth0ManagementAPIResourceImport,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "name", "Auth0 Management API"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "identifier", "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "token_lifetime", "86400"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "skip_consent_for_verifiable_first_party_clients", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "token_lifetime_for_web", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "scopes.#", "0"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "verification_location"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "enforce_policies"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "token_dialect"),
				),
			},
		},
	})
}
