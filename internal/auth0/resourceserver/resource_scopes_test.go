package resourceserver_test

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccResourceServerScopesPreventErasingScopesOnCreate(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - API - Prevent Erasing"
	identifier = "https://uat.api.terraform-provider-auth0.com/prevent-erasing-scopes"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}

# Pre-existing scopes
resource "auth0_resource_server_scope" "read_posts" {
	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts"
}

resource "auth0_resource_server_scopes" "my_scopes" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name = "read:appointments"
		description = "Ability to read appointments"
	}
}
`,
				ExpectError: regexp.MustCompile("Resource Server with non empty scopes"),
			},
		},
	})
}

const testAccResourceServerScopesImport = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name = "create:appointments"
		description = "Ability to create appointments"
	}

	scopes {
		name = "read:appointments"
		description = "Ability to read appointments"
	}
}
`

func TestAccResourceServerScopesImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	testName := strings.ToLower(t.Name())
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccResourceServerScopesImport, testName),
				ResourceName:       "auth0_resource_server.my_api",
				ImportState:        true,
				ImportStateId:      fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccResourceServerScopesImport, testName),
				ResourceName:       "auth0_resource_server_scopes.my_api_scopes",
				ImportState:        true,
				ImportStateId:      fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName),
				ImportStatePersist: true,
			},
			{
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Config: acctest.ParseTestName(testAccResourceServerScopesImport, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
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

const testAccCreateResourceServerScopesWithOneScope = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name = "create:appointments"
		description = "Ability to create appointments"
	}
}
`

const testAccCreateResourceServerScopesWithTwoScope = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}

resource "auth0_resource_server_scopes" "my_api_scopes" {
	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name = "create:appointments"
		description = "Ability to create appointments"
	}

	scopes {
		name = "read:appointments"
		description = "Ability to read appointments"
	}
}
`

const testAccDeleteResourceServerScopes = `
resource "auth0_resource_server" "my_api" {
	name       = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
}
`

func TestAccResourceServerScopes(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateResourceServerScopesWithOneScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "resource_server_identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.name", "create:appointments"),
					resource.TestCheckResourceAttr("auth0_resource_server_scopes.my_api_scopes", "scopes.0.description", "Ability to create appointments"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateResourceServerScopesWithTwoScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "name", fmt.Sprintf("Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)),
					resource.TestCheckResourceAttr("auth0_resource_server.my_api", "scopes.#", "0"),
				),
			},
		},
	})
}
