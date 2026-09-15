package client_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAClientAndAResourceServerWithScopes = `
resource "auth0_client" "my_client" {
	name                 = "Acceptance Test - Client Grant - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_resource_server" "my_resource_server" {
	name       = "Acceptance Test - Client Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/client-grant/{{.testName}}"
	authorization_details {
    	type = "payment"
  	}
  	authorization_details {
    	type = "shipping"
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

resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_resource_server ]

	resource_server_identifier = auth0_resource_server.my_resource_server.identifier

	scopes {
		name        = "create:foo"
		description = "Create foos"
	}

	scopes {
		name        = "create:bar"
		description = "Create bars"
	}
}
`

const testAccClientGrantConfigCreate = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = ["create:bar"]
	subject_type = "user"
	authorization_details_types = ["payment","shipping"]
}
`

const testAccClientGrantConfigUpdate = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ "create:foo" ]
	subject_type = "user"
	authorization_details_types = ["payment"]
}
`

const testAccClientGrantConfigUpdateAgain = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
	subject_type = "user"
	authorization_details_types = ["payment"]
}
`

const testAccClientGrantConfigUpdateAgainWithAllowAllScopes = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	subject_type = "user"
	authorization_details_types = ["payment"]
	allow_all_scopes = true
}
`

const testAccClientGrantConfigUpdateFromAllowAllScopesToSpecificScopes = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id = auth0_client.my_client.id
	audience  = auth0_resource_server.my_resource_server.identifier
	subject_type = "user"
	authorization_details_types = ["payment"]
	scopes    = [ "create:foo" ]
}
`

const testAccClientGrantConfigUpdateChangeClient = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client" "my_client_alt" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name                 = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_client.my_client_alt ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
	subject_type = "user"
	authorization_details_types = ["payment"]
}
`

const testAccAlreadyExistingGrantWillNotConflict = testAccGivenAClientAndAResourceServerWithScopes + `
resource "auth0_client" "my_client_alt" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	name                 = "Acceptance Test - Client Grant Alt - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_client.my_client_alt ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
	subject_type = "user"
	authorization_details_types = ["payment"]
}

resource "auth0_client_grant" "no_conflict_client_grant" {
	depends_on = [ auth0_client_grant.my_client_grant ]

	client_id = auth0_client.my_client_alt.id
	audience  = auth0_resource_server.my_resource_server.identifier
	scopes    = [ ]
	subject_type = "user"
	authorization_details_types = ["payment"]
}
`

func TestAccClientGrant(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "audience", fmt.Sprintf("https://uat.tf.terraform-provider-auth0.com/client-grant/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "subject_type", "user"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "authorization_details_types.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.0", "create:foo"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "subject_type", "user"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "authorization_details_types.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateAgainWithAllowAllScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "allow_all_scopes", "true"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateFromAllowAllScopesToSpecificScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.0", "create:foo"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "allow_all_scopes", "false"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigUpdateChangeClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccAlreadyExistingGrantWillNotConflict, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.no_conflict_client_grant", "scopes.#", "0"),
				),
			},
		},
	})
}

const testAccClientGrantDefaultForCreate = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - Default For Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/default-for-grant/{{.testName}}"
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [auth0_resource_server.my_api]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "read:data"
		description = "Read data"
	}
}

resource "auth0_client_grant" "default_for_grant" {
	depends_on = [auth0_resource_server_scopes.my_scopes]

	default_for = "third_party_clients"
	audience    = auth0_resource_server.my_api.identifier
	scopes      = ["read:data"]
}
`

const testAccClientGrantDefaultForUpdate = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - Default For Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/default-for-grant/{{.testName}}"
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [auth0_resource_server.my_api]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "read:data"
		description = "Read data"
	}

	scopes {
		name        = "write:data"
		description = "Write data"
	}
}

resource "auth0_client_grant" "default_for_grant" {
	depends_on = [auth0_resource_server_scopes.my_scopes]

	default_for = "third_party_clients"
	audience    = auth0_resource_server.my_api.identifier
	scopes      = ["read:data", "write:data"]
}
`

func TestAccClientGrant_DefaultFor(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientGrantDefaultForCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.default_for_grant", "default_for", "third_party_clients"),
					resource.TestCheckResourceAttr("auth0_client_grant.default_for_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.default_for_grant", "scopes.0", "read:data"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantDefaultForUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.default_for_grant", "default_for", "third_party_clients"),
					resource.TestCheckResourceAttr("auth0_client_grant.default_for_grant", "scopes.#", "2"),
				),
			},
		},
	})
}

const testAccGivenAClientAndAResourceServerForAnonymousUser = `
resource "auth0_client" "my_client" {
	name                 = "Acceptance Test - Anonymous Grant - {{.testName}}"
	custom_login_page_on = true
	is_first_party       = true
}

resource "auth0_resource_server" "my_resource_server" {
	name       = "Acceptance Test - Anonymous Grant - {{.testName}}"
	identifier = "https://uat.tf.terraform-provider-auth0.com/anon-grant/{{.testName}}"

	subject_type_authorization {
		anonymous_user {
			policy = "require_client_grant"
		}
	}
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_resource_server ]

	resource_server_identifier = auth0_resource_server.my_resource_server.identifier

	scopes {
		name        = "create:foo"
		description = "Create foos"
	}
}
`

const testAccClientGrantConfigAnonymousUser = testAccGivenAClientAndAResourceServerForAnonymousUser + `
resource "auth0_client_grant" "my_client_grant" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	client_id    = auth0_client.my_client.id
	audience     = auth0_resource_server.my_resource_server.identifier
	scopes       = ["create:foo"]
	subject_type = "anonymous_user"
}
`

func TestAccClientGrant_AnonymousUser(t *testing.T) {
	testAccPreCheckFeatureAnonymousSessions(t)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigAnonymousUser, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "subject_type", "anonymous_user"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_grant.my_client_grant", "scopes.0", "create:foo"),
				),
			},
		},
	})
}

// The following tests cover the CustomizeDiff mutual-exclusivity checks in
// validateClientGrant: organization_usage, allow_any_organization,
// authorization_details_types, and default_for cannot be set alongside
// subject_type = "anonymous_user". The errors are raised at plan time (before any
// API call), so each records only the provider-configure interaction. They are not
// gated behind the feature flag because the validation runs regardless of it.

const testAccClientGrantAnonymousUserWithOrganizationUsage = `
resource "auth0_client_grant" "my_client_grant" {
	client_id          = "test-client-id"
	audience           = "https://api.example.com/anonymous"
	scopes             = []
	subject_type       = "anonymous_user"
	organization_usage = "deny"
}
`

const testAccClientGrantAnonymousUserWithAllowAnyOrganization = `
resource "auth0_client_grant" "my_client_grant" {
	client_id              = "test-client-id"
	audience               = "https://api.example.com/anonymous"
	scopes                 = []
	subject_type           = "anonymous_user"
	allow_any_organization = true
}
`

const testAccClientGrantAnonymousUserWithAuthorizationDetailsTypes = `
resource "auth0_client_grant" "my_client_grant" {
	client_id                   = "test-client-id"
	audience                    = "https://api.example.com/anonymous"
	scopes                      = []
	subject_type                = "anonymous_user"
	authorization_details_types = ["payment"]
}
`

// default_for conflicts with client_id at the schema level, so client_id is omitted
// here to let the plan reach the anonymous_user CustomizeDiff check.
const testAccClientGrantAnonymousUserWithDefaultFor = `
resource "auth0_client_grant" "my_client_grant" {
	audience     = "https://api.example.com/anonymous"
	scopes       = []
	subject_type = "anonymous_user"
	default_for  = "third_party_clients"
}
`

func TestAccClientGrantAnonymousUserRejectsOrganizationUsage(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccClientGrantAnonymousUserWithOrganizationUsage,
				ExpectError: regexp.MustCompile("`organization_usage` cannot be set for client grants with `subject_type`: anonymous_user"),
			},
		},
	})
}

func TestAccClientGrantAnonymousUserRejectsAllowAnyOrganization(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccClientGrantAnonymousUserWithAllowAnyOrganization,
				ExpectError: regexp.MustCompile("`allow_any_organization` cannot be set for client grants with `subject_type`: anonymous_user"),
			},
		},
	})
}

func TestAccClientGrantAnonymousUserRejectsAuthorizationDetailsTypes(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccClientGrantAnonymousUserWithAuthorizationDetailsTypes,
				ExpectError: regexp.MustCompile("`authorization_details_types` cannot be set for client grants with `subject_type`: anonymous_user"),
			},
		},
	})
}

func TestAccClientGrantAnonymousUserRejectsDefaultFor(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccClientGrantAnonymousUserWithDefaultFor,
				ExpectError: regexp.MustCompile("`default_for` cannot be set for client grants with `subject_type`: anonymous_user"),
			},
		},
	})
}
