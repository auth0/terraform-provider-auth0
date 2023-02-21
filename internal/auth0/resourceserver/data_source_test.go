package resourceserver_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
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

const testAccDataConnectionConfigByIdentifier = testAccGivenAResourceServer + `
data "auth0_resource_server" "test" {
	depends_on = [ auth0_resource_server.my_api ]

	identifier = auth0_resource_server.my_api.identifier
}
`

const testAccDataConnectionConfigByID = testAccGivenAResourceServer + `
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
				Config: template.ParseTestName(testAccGivenAResourceServer, t.Name()),
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
				Config: template.ParseTestName(testAccDataConnectionConfigByIdentifier, t.Name()),
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
				Config: template.ParseTestName(testAccGivenAResourceServer, t.Name()),
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
				Config: template.ParseTestName(testAccDataConnectionConfigByID, t.Name()),
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
