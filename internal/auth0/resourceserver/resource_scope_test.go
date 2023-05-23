package resourceserver_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAResourceServer = `
resource "auth0_resource_server" "resource_server" {
	name = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"

	lifecycle {
		ignore_changes = [ scopes ]
	}
}
`

const givenAScope = `
resource "auth0_resource_server_scope" "read_posts" { 
	scope = "read:posts"
	resource_server_identifier = auth0_resource_server.resource_server.identifier

	description = "Can read posts"
}
`

const givenAnUpdatedScope = `
resource "auth0_resource_server_scope" "read_posts" { 
	scope = "read:posts"
	resource_server_identifier = auth0_resource_server.resource_server.identifier

	description = "Can read posts from API"
}
`

const givenAnotherScope = `
resource "auth0_resource_server_scope" "write_posts" { 
	depends_on = [ auth0_resource_server_scope.read_posts ] 

	scope = "write:posts"
	resource_server_identifier = auth0_resource_server.resource_server.identifier
}
`

const testAccNoScopesAssigned = givenAResourceServer
const testAccOneScopeAssigned = givenAResourceServer + givenAScope + `data "auth0_resource_server" "resource_server" {
	depends_on = [ auth0_resource_server_scope.read_posts ]
	identifier = auth0_resource_server.resource_server.identifier
}`

const testAccOneScopeAssignedWithUpdate = givenAResourceServer + givenAnUpdatedScope + `data "auth0_resource_server" "resource_server" {
	depends_on = [ auth0_resource_server_scope.read_posts ]
	identifier = auth0_resource_server.resource_server.identifier
}`

const testAccTwoScopesAssigned = givenAResourceServer + givenAnUpdatedScope + givenAnotherScope + `data "auth0_resource_server" "resource_server" {
	depends_on = [ auth0_resource_server_scope.read_posts, auth0_resource_server_scope.write_posts]
	identifier = auth0_resource_server.resource_server.identifier
}`

const resourceServerIdentifier = "https://uat.api.terraform-provider-auth0.com/testaccresourceserverscope"

func TestAccResourceServerScope(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNoScopesAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.resource_server", "scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOneScopeAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "scope", "read:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "resource_server_identifier", resourceServerIdentifier),

					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.0.value", "read:posts"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.0.description", "Can read posts"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOneScopeAssignedWithUpdate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.#", "1"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.read_posts", "description", "Can read posts from API"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTwoScopesAssigned, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.#", "2"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "scope", "write:posts"),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "description", ""),
					resource.TestCheckResourceAttr("auth0_resource_server_scope.write_posts", "resource_server_identifier", resourceServerIdentifier),

					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.0.value", "write:posts"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.0.description", ""),
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.1.value", "read:posts"),
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.1.description", "Can read posts from API"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOneScopeAssignedWithUpdate, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_resource_server.resource_server", "scopes.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccNoScopesAssigned, strings.ToLower(t.Name())),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.resource_server", "scopes.#", "0"),
				),
			},
		},
	})
}
