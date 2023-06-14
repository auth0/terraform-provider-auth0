package resourceserver_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccResourceWillNotFailOnCreateIfScopeAlreadyExisting = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scope" "read_posts_copy" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts"
}

resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server_scope.read_posts_copy ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts"
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scope.read_posts_copy ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccResourceServerWithOneScope = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts"
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccResourceServerWithUpdatedScope = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts from API"
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccResourceServerWithTwoScopes = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts from API"
}

resource "auth0_resource_server_scope" "write_posts" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope = "write:posts"
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scope.write_posts ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

const testAccResourceServerScopeImportSetup = testAccGivenAResourceServerWithNoScopes + `
resource "auth0_resource_server_scopes" "my_api_scopes" {
	depends_on = [ auth0_resource_server.my_api ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scopes {
		name        = "read:posts"
		description = "Can read posts from API"
	}

	scopes {
		name = "write:posts"
	}
}
`

const testAccResourceServerScopeImportCheck = testAccResourceServerScopeImportSetup + `
resource "auth0_resource_server_scope" "read_posts" {
	depends_on = [ auth0_resource_server_scopes.my_api_scopes ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope       = "read:posts"
	description = "Can read posts from API"
}

resource "auth0_resource_server_scope" "write_posts" {
	depends_on = [ auth0_resource_server_scope.read_posts ]

	resource_server_identifier = auth0_resource_server.my_api.identifier

	scope = "write:posts"
}

data "auth0_resource_server" "my_api" {
	depends_on = [ auth0_resource_server_scope.write_posts ]

	resource_server_id = auth0_resource_server.my_api.id
}
`

func TestAccResourceServerScope(t *testing.T) {
	testName := strings.ToLower(t.Name())
	resourceServerIdentifier := fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", testName)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccResourceWillNotFailOnCreateIfScopeAlreadyExisting, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "resource_server_identifier", resourceServerIdentifier),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts_copy", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts_copy", "description", "Can read posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts_copy", "resource_server_identifier", resourceServerIdentifier),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDeleteResourceServerScopes, testName),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerWithOneScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "resource_server_identifier", resourceServerIdentifier),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerWithUpdatedScope, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts from API"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "resource_server_identifier", resourceServerIdentifier),
				),
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerWithTwoScopes, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts from API"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "resource_server_identifier", resourceServerIdentifier),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "scope", "write:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "description", ""),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "resource_server_identifier", resourceServerIdentifier),
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
				Config: acctest.ParseTestName(testAccResourceServerScopeImportSetup, testName),
			},
			{
				Config:       acctest.ParseTestName(testAccResourceServerScopeImportCheck, testName),
				ResourceName: "auth0_resource_server_scope.read_posts",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.my_api", "identifier")
					assert.NoError(t, err)

					return apiID + "::read:posts", nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccResourceServerScopeImportCheck, testName),
				ResourceName: "auth0_resource_server_scope.write_posts",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					apiID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_resource_server.my_api", "identifier")
					assert.NoError(t, err)

					return apiID + "::write:posts", nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccResourceServerScopeImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.my_api", "scopes.#", "2"),
				),
			},
		},
	})
}
