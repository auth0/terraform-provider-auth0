package auth0

import (
	"log"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
)

func init() {
	resource.AddTestSweepers("auth0_resource_server", &resource.Sweeper{
		Name: "auth0_resource_server",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			fn := func(rs *management.ResourceServer) {
				log.Printf("[DEBUG] ➝ %s", rs.GetName())
				if strings.Contains(rs.GetName(), "Test") {
					if err := api.ResourceServer.Delete(rs.GetID()); err != nil {
						log.Printf("[DEBUG] Failed to delete resource server with ID: %s", rs.GetID())
					}
					log.Printf("[DEBUG] ✗ %s", rs.GetName())
				}
			}

			return api.ResourceServer.Stream(fn, management.IncludeFields("id", "name"))
		},
	})
}

func TestAccResourceServer(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccResourceServerConfigCreate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", "Acceptance Test - {{.random}}", rand),
					random.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", "https://uat.api.alexkappa.com/{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "enforce_policies", "true"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_resource_server",
						"scopes.*",
						map[string]string{
							"value":       "create:foo",
							"description": "Create foos",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_resource_server",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
							"description": "Create bars",
						},
					),
				),
			},
			{
				Config: random.Template(testAccResourceServerConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "scopes.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"auth0_resource_server.my_resource_server",
						"scopes.*",
						map[string]string{
							"value":       "create:bar",
							"description": "Create bars for bar reasons",
						},
					),
				),
			},
		},
	})
}

const testAccResourceServerConfigCreate = `
resource "auth0_resource_server" "my_resource_server" {
	name = "Acceptance Test - {{.random}}"
	identifier = "https://uat.api.alexkappa.com/{{.random}}"
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

const testAccResourceServerConfigUpdate = `
resource "auth0_resource_server" "my_resource_server" {
	name = "Acceptance Test - {{.random}}"
	identifier = "https://uat.api.alexkappa.com/{{.random}}"
	signing_alg = "RS256"
	scopes {
		value = "create:foo"
		description = "Create foos"
	}
	scopes {
		value = "create:bar"
		description = "Create bars for bar reasons"
	}
	allow_offline_access = false # <--- set to false
	token_lifetime = 7200
	token_lifetime_for_web = 3600
	skip_consent_for_verifiable_first_party_clients = true
	enforce_policies = true
}
`
