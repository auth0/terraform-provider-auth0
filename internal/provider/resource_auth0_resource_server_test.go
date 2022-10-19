package provider

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
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
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccResourceServerConfigEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", ""),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "scopes.#", "0"),
					resource.TestCheckResourceAttrSet("auth0_resource_server.my_resource_server", "signing_alg"),
					resource.TestCheckResourceAttrSet("auth0_resource_server.my_resource_server", "token_lifetime_for_web"),
				),
			},
			{
				Config: template.ParseTestName(testAccResourceServerConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
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
				Config: template.ParseTestName(testAccResourceServerConfigUpdate, t.Name()),
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
			{
				Config: template.ParseTestName(testAccResourceServerConfigEmptyAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "identifier", fmt.Sprintf("https://uat.api.terraform-provider-auth0.com/%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "name", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "scopes.#", "0"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "token_lifetime_for_web", "3600"),
					resource.TestCheckResourceAttr("auth0_resource_server.my_resource_server", "skip_consent_for_verifiable_first_party_clients", "true"),
				),
			},
		},
	})
}

const testAccResourceServerConfigEmpty = `
resource "auth0_resource_server" "my_resource_server" {
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
}
`

const testAccResourceServerConfigCreate = `
resource "auth0_resource_server" "my_resource_server" {
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

const testAccResourceServerConfigUpdate = `
resource "auth0_resource_server" "my_resource_server" {
	name = "Acceptance Test - {{.testName}}"
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
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

const testAccResourceServerConfigEmptyAgain = `
resource "auth0_resource_server" "my_resource_server" {
	identifier = "https://uat.api.terraform-provider-auth0.com/{{.testName}}"
	name = "Acceptance Test - {{.testName}}"
}
`

func TestAccResourceServerAuth0APIManagement(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != recorder.RecordingsDomain {
		t.Skip()
	}

	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: `
resource "auth0_resource_server" "auth0" {
	name = "Auth0 Management API"
	identifier = "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"
	token_lifetime = 86400
	skip_consent_for_verifiable_first_party_clients = false
}
`,
				ResourceName:       "auth0_resource_server.auth0",
				ImportState:        true,
				ImportStateId:      "xxxxxxxxxxxxxxxxxxxx",
				ImportStatePersist: true,
			},
			{
				Config: `
resource "auth0_resource_server" "auth0" {
	name = "Auth0 Management API"
	identifier = "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"
	token_lifetime = 86400
	skip_consent_for_verifiable_first_party_clients = false
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "name", "Auth0 Management API"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "identifier", "https://terraform-provider-auth0-dev.eu.auth0.com/api/v2/"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "token_lifetime", "86400"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "skip_consent_for_verifiable_first_party_clients", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "allow_offline_access", "false"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "signing_alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "token_lifetime_for_web", "7200"),
					resource.TestCheckResourceAttr("auth0_resource_server.auth0", "scopes.#", "0"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "verification_location"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "options"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "enforce_policies"),
					resource.TestCheckNoResourceAttr("auth0_resource_server.auth0", "token_dialect"),
				),
			},
		},
	})
}
