package resourceserver_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServerWithScopes = `
resource "auth0_resource_server" "my_api" {
	name                                            = "Acceptance Test - {{.testName}}"
	identifier                                      = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg                                     = "RS256"
	allow_offline_access                            = true
	token_lifetime                                  = 7200
	token_lifetime_for_web                          = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies                                = true
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

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

const testAccDataResourceServerNonExistentIdentifier = testAccGivenAResourceServerWithScopes + `
data "auth0_resource_server" "test" {
	identifier = "this-resource-server-does-not-exist"
}
`

const testAccDataResourceServerConfigByIdentifier = testAccGivenAResourceServerWithScopes + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server_scopes.my_scopes ]

	identifier = auth0_resource_server.my_api.identifier
}
`

const testAccDataResourceServerConfigByID = testAccGivenAResourceServerWithScopes + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server_scopes.my_scopes ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccDataAuth0ManagementAPI = `
data "auth0_resource_server" "auth0" {
	identifier = %q
}
`

func TestAccDataSourceResourceServerRequiredArguments(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_resource_server" "test" { }`,
				ExpectError: regexp.MustCompile("one of `identifier,resource_server_id` must be specified"),
			},
		},
	})
}

func TestAccDataSourceResourceServer(t *testing.T) {
	managementAPIIdentifier := "https://" + os.Getenv("AUTH0_DOMAIN") + "/api/v2/"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataResourceServerNonExistentIdentifier, t.Name()),
				ExpectError: regexp.MustCompile(
					`404 Not Found: The resource server does not exist`,
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataResourceServerConfigByIdentifier, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"name":        "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"name":        "create:bar",
							"description": "Create bars",
						},
					),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataResourceServerConfigByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_resource_server.test", "resource_server_id"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.test", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"name":        "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"name":        "create:bar",
							"description": "Create bars",
						},
					),
				),
			},
			{
				Config: fmt.Sprintf(testAccDataAuth0ManagementAPI, managementAPIIdentifier),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "name", "Auth0 Management API"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "identifier", managementAPIIdentifier),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_lifetime", "86400"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "skip_consent_for_verifiable_first_party_clients", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_lifetime_for_web", "7200"),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_resource_server.auth0", "scopes.*", map[string]string{
						"name":        "read:users",
						"description": "Read Users",
					}), // Checking just one to ensure that scopes are not empty, as they get expanded periodically.
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "verification_location", ""),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_dialect", ""),
				),
			},
		},
	})
}
