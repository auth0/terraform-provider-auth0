package resourceserver_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAResourceServerWithNoScopes = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
}
`

const testAccResourceServerScopesPreventErasingScopesOnCreate = testAccGivenAResourceServerWithNoScopes + `
# Pre-existing scopes
resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts"
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "read:appointments"
		description = "Ability to read appointments"
	}
}
`

const testAccCreateResourceServerScopesWithOneScope = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "create:appointments"
		description = "Ability to create appointments"
	}
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccCreateResourceServerScopesWithTwoScope = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "create:appointments"
		description = "Ability to create appointments"
	}

	scopes {
		name        = "read:appointments"
		description = "Ability to read appointments"
	}
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccDeleteResourceServerScopes = testAccGivenAResourceServerWithNoScopes + `
data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccResourceServerScopesImportSetup = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scope" "create_appointments" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "create:appointments"
	description = "Ability to create appointments"
}

resource "auth0_resource_server_scope" "read_appointments" {
	depends_on = [ auth0_resource_server_scope.create_appointments ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:appointments"
	description = "Ability to read appointments"
}
`

const testAccResourceServerScopesImportCheck = testAccResourceServerScopesImportSetup + `
resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server_scope.read_appointments ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "create:appointments"
		description = "Ability to create appointments"
	}

	scopes {
		name        = "read:appointments"
		description = "Ability to read appointments"
	}
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

func TestAccResourceServerScopes(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccResourceServerScopesPreventErasingScopesOnCreate, testName),
				ExpectError: regexp.MustCompile("Resource Server with non empty scopes"),
			},
			{
				Config: acctest.ParseTestName(testAccDeleteResourceServerScopes, testName),
			},
			{
				Config: acctest.ParseTestName(testAccCreateResourceServerScopesWithOneScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.name", "create:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.description", "Ability to create appointments"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateResourceServerScopesWithTwoScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.name", "create:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.description", "Ability to create appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.1.name", "read:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.1.description", "Ability to read appointments"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeleteResourceServerScopes, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerScopesImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccResourceServerScopesImportCheck, testName),
				ResourceName: "auth0_resource_server_scopes.my_api_scopes",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.my_api", "id")
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerScopesImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.name", "create:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.description", "Ability to create appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.1.name", "read:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.1.description", "Ability to read appointments"),
				),
			},
		},
	})
}
