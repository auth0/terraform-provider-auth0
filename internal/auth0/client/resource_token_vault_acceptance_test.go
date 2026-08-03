package client_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

// testAccClientWithoutTokenVaultPrivilegedAccess is used both as the starting
// point (to prove no perpetual diff exists for a client with no block) and to
// verify an unrelated field update does not clobber a previously-set block
// (analysis R1) once combined with a following step that re-adds it.
const testAccClientWithoutTokenVaultPrivilegedAccess = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name     = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	app_type = "non_interactive"
}
`

const testAccClientWithTokenVaultPrivilegedAccess = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name     = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	app_type = "non_interactive"

	token_vault_privileged_access {
		credentials {
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}

		ip_allowlist = ["10.0.0.1", "192.168.1.0/24"]

		grants {
			connection = auth0_connection.google.name
			scopes     = ["openid", "profile"]
		}
	}
}
`

// testAccClientWithTokenVaultPrivilegedAccessUnrelatedFieldChanged changes
// only description (unrelated to the block) to verify a v1 client update does
// not clobber a stored token_vault_privileged_access object (analysis R1).
const testAccClientWithTokenVaultPrivilegedAccessUnrelatedFieldChanged = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name        = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	description = "Now with a description"
	app_type    = "non_interactive"

	token_vault_privileged_access {
		credentials {
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}

		ip_allowlist = ["10.0.0.1", "192.168.1.0/24"]

		grants {
			connection = auth0_connection.google.name
			scopes     = ["openid", "profile"]
		}
	}
}
`

const testAccClientWithTokenVaultPrivilegedAccessUpdatedIPAllowlist = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name        = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	description = "Now with a description"
	app_type    = "non_interactive"

	token_vault_privileged_access {
		credentials {
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}

		ip_allowlist = ["10.0.0.1"]

		grants {
			connection = auth0_connection.google.name
			scopes     = ["openid", "profile"]
		}
	}
}
`

// testAccClientWithTokenVaultPrivilegedAccessGrantsCleared clears grants to an
// empty list, distinct from removing the block entirely (analysis §5 step 4).
const testAccClientWithTokenVaultPrivilegedAccessGrantsCleared = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name        = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	description = "Now with a description"
	app_type    = "non_interactive"

	token_vault_privileged_access {
		credentials {
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}

		ip_allowlist = ["10.0.0.1"]
	}
}
`

// testAccClientWithTokenVaultPrivilegedAccessCredentialReplaced swaps the
// credential material, exercising create-new + delete-stale orchestration
// (analysis §5 step 5).
const testAccClientWithTokenVaultPrivilegedAccessCredentialReplaced = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name        = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	description = "Now with a description"
	app_type    = "non_interactive"

	token_vault_privileged_access {
		credentials {
			credential_type = "public_key"
			pem             = <<EOF
%s
EOF
		}

		ip_allowlist = ["10.0.0.1"]

		grants {
			connection = auth0_connection.google.name
			scopes     = ["openid"]
		}
	}
}
`

// testAccClientWithoutTokenVaultPrivilegedAccessDescriptionKept removes only
// the block, keeping description set so the assertion isolates the block's
// removal rather than confounding it with unrelated optional-field-drop
// behavior (dropping description from config does not clear it server-side,
// which is pre-existing auth0_client behavior unrelated to this feature).
const testAccClientWithoutTokenVaultPrivilegedAccessDescriptionKept = `
resource "auth0_connection" "google" {
	name     = "Acceptance-Test-TVPA-Google-{{.testName}}"
	strategy = "google-oauth2"
}

resource "auth0_client" "worker" {
	name        = "Acceptance Test - Token Vault Privileged Access - {{.testName}}"
	description = "Now with a description"
	app_type    = "non_interactive"
}
`

func TestAccClientTokenVaultPrivilegedAccess(t *testing.T) {
	credsCert1, err := os.ReadFile("./../../../test/data/creds-cert-1.pem")
	require.NoError(t, err)

	credsCert2, err := os.ReadFile("./../../../test/data/creds-cert-2.pem")
	require.NoError(t, err)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// No perpetual diff for a client with no block configured.
				Config: acctest.ParseTestName(testAccClientWithoutTokenVaultPrivilegedAccess, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.#", "0"),
				),
			},
			{
				// Create with all three sub-fields populated.
				Config: fmt.Sprintf(acctest.ParseTestName(testAccClientWithTokenVaultPrivilegedAccess, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.credentials.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_client.worker", "token_vault_privileged_access.0.credentials.0.id"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.credentials.0.credential_type", "public_key"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.ip_allowlist.#", "2"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.0.connection", fmt.Sprintf("Acceptance-Test-TVPA-Google-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.0.scopes.#", "2"),
				),
			},
			{
				// R1: changing an unrelated auth0_client field must not clobber the
				// stored token_vault_privileged_access object.
				Config: fmt.Sprintf(acctest.ParseTestName(testAccClientWithTokenVaultPrivilegedAccessUnrelatedFieldChanged, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "description", "Now with a description"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.ip_allowlist.#", "2"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.#", "1"),
				),
			},
			{
				// Update ip_allowlist.
				Config: fmt.Sprintf(acctest.ParseTestName(testAccClientWithTokenVaultPrivilegedAccessUpdatedIPAllowlist, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.ip_allowlist.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.ip_allowlist.0", "10.0.0.1"),
				),
			},
			{
				// Clear grants to [] -- distinct from removing the block entirely.
				Config: fmt.Sprintf(acctest.ParseTestName(testAccClientWithTokenVaultPrivilegedAccessGrantsCleared, t.Name()), credsCert1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.#", "0"),
				),
			},
			{
				// Replace the credential material: create-new + delete-stale.
				Config: fmt.Sprintf(acctest.ParseTestName(testAccClientWithTokenVaultPrivilegedAccessCredentialReplaced, t.Name()), credsCert2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.0.grants.#", "1"),
				),
			},
			{
				// Remove the block entirely; assert the object is gone.
				Config: acctest.ParseTestName(testAccClientWithoutTokenVaultPrivilegedAccessDescriptionKept, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.worker", "token_vault_privileged_access.#", "0"),
				),
			},
		},
	})
}
