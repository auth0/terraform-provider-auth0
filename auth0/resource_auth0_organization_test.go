package auth0

import (
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_organization", &resource.Sweeper{
		Name: "auth0_organization",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				organizationList, err := api.Organization.List(management.Page(page))
				if err != nil {
					return err
				}

				for _, organization := range organizationList.Organizations {
					log.Printf("[DEBUG] ➝ %s", organization.GetName())

					if strings.Contains(organization.GetName(), "test") {
						result = multierror.Append(
							result,
							api.Organization.Delete(organization.GetID()),
						)

						log.Printf("[DEBUG] ✗ %s", organization.GetName())
					}
				}
				if !organizationList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}

const testAccOrganizationGiven2Connections = `
resource auth0_connection acme {
	name = "Acceptance-Test-Connection-Acme-{{.testName}}"
	strategy = "auth0"
}

resource auth0_connection acmeinc {
	depends_on = [auth0_connection.acme]
	name = "Acceptance-Test-Connection-Acme-Inc-{{.testName}}"
	strategy = "auth0"
}
`

const testAccOrganizationCreate = testAccOrganizationGiven2Connections + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"

	metadata = {
		some_key = "some_value"
	}

	connections {
		connection_id = auth0_connection.acme.id
	}
}
`

const testAccOrganizationUpdate = testAccOrganizationGiven2Connections + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"

	metadata = {
		some_key = "some_value"
		another_key = "another_value"
	}

	branding {
		logo_url = "https://acme.com/logo.svg"
		colors = {
			primary = "#e3e2f0"
			page_background = "#e3e2ff"
		}
	}

	connections {
		connection_id = auth0_connection.acme.id
		assign_membership_on_login = false
	}

	connections {
		connection_id = auth0_connection.acmeinc.id
		assign_membership_on_login = true
	}
}
`

const testAccOrganizationUpdateAgain = testAccOrganizationGiven2Connections + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"

	metadata = {
		some_key = "some_value"
	}

	branding {
		logo_url = "https://acme.com/logo.svg"
		colors = {
			primary = "#e3e2f0"
			page_background = "#e3e2ff"
		}
	}

	connections {
		connection_id = auth0_connection.acmeinc.id
		assign_membership_on_login = false
	}
}
`

const testAccOrganizationUpdateAgainAndAgain = testAccOrganizationGiven2Connections + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"

	branding {
		logo_url = "https://acme.com/logo.svg"
		colors = {
			primary = "#e3e2f0"
			page_background = "#e3e2ff"
		}
	}

	connections {
		connection_id = auth0_connection.acmeinc.id
		assign_membership_on_login = true
	}
}
`

const testAccOrganizationRemoveAllConnections = `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}
`

func TestAccOrganization(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccOrganizationCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.some_key", "some_value"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_organization.acme", "connections.0.connection_id"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.0.assign_membership_on_login", "false"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationUpdate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.logo_url", "https://acme.com/logo.svg"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.%", "2"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.primary", "#e3e2f0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.page_background", "#e3e2ff"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "2"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.some_key", "some_value"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.another_key", "another_value"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "2"),
					resource.TestCheckResourceAttrSet("auth0_organization.acme", "connections.0.connection_id"),
					resource.TestCheckResourceAttrSet("auth0_organization.acme", "connections.1.connection_id"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.0.assign_membership_on_login", "false"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.1.assign_membership_on_login", "true"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationUpdateAgain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.logo_url", "https://acme.com/logo.svg"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.%", "2"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.primary", "#e3e2f0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.page_background", "#e3e2ff"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.some_key", "some_value"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_organization.acme", "connections.0.connection_id"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.0.assign_membership_on_login", "false"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationUpdateAgainAndAgain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.logo_url", "https://acme.com/logo.svg"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.%", "2"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.primary", "#e3e2f0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.0.colors.page_background", "#e3e2ff"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "1"),
					resource.TestCheckResourceAttrSet("auth0_organization.acme", "connections.0.connection_id"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.0.assign_membership_on_login", "true"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationRemoveAllConnections, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "0"),
				),
			},
		},
	})
}
