package resourceserver_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServer = `
resource "auth0_resource_server" "my_api" {
	name = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	signing_alg = "RS256"
	scopes {
		value = "create:foo"
		description = "Create foos"
	}
	scopes {
		value = "create:bar"
		description = "Create bars"
	}
	allow_offline_access = true
	token_lifetime = 7200
	token_lifetime_for_web = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies = true
}
`

const testAccDataResourceServerConfigByIdentifier = testAccGivenAResourceServer + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server.my_api ]

	identifier = auth0_resource_server.my_api.identifier
}
`

const testAccDataResourceServerConfigByID = testAccGivenAResourceServer + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_id = auth0_resource_server.my_api.id
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

func TestAccDataSourceResourceServerByIdentifier(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccGivenAResourceServer, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_api",
						"scopes.*",
						map[string]string{
							"value":       "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_api",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
							"description": "Create bars",
						},
					),
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
							"value":       "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
							"description": "Create bars",
						},
					),
				),
			},
		},
	})
}

func TestAccDataSourceResourceServerByID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccGivenAResourceServer, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_api",
						"scopes.*",
						map[string]string{
							"value":       "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_api",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
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
							"value":       "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_resource_server.test",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
							"description": "Create bars",
						},
					),
				),
			},
		},
	})
}

const testAccDataAuth0ManagementAPI = `
data "auth0_resource_server" "auth0" {
	resource_server_id = "112233445566777899011232"
}
`

func TestAccDataResourceServerAuth0APIManagement(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// Skip this test if we're running with a real domain as the Auth0 Management API
		// is a singleton resource always created on the tenant and each tenant
		// will have it created with different IDs and Identifiers.
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataAuth0ManagementAPI,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "name", "Auth0 Management API"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "identifier", "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_lifetime", "86400"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "skip_consent_for_verifiable_first_party_clients", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_lifetime_for_web", "7200"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "scopes.#", "136"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "verification_location", ""),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "enforce_policies", "false"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.auth0", "token_dialect", ""),
				),
			},
		},
	})
}

const testAccDataResourceServerNonexistentIdentifier = testAccGivenAResourceServer + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server.my_api ]

	identifier = "this-resource-server-does-not-exist"
}
`

func TestAccDataSourceResourceServerNonexistentIdentifier(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataResourceServerNonexistentIdentifier, t.Name()),
				ExpectError: regexp.MustCompile(
					"404 Not Found: The resource server does not exist",
				),
			},
		},
	})
}
