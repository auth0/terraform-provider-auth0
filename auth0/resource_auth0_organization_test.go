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
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "1"),
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
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "2"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationUpdateAgain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "connections.#", "1"),
				),
			},
		},
	})
}

const testAccOrganizationAux = `
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

const testAccOrganizationCreate = testAccOrganizationAux + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"

	connections {
		connection_id = auth0_connection.acme.id
	}
}
`

const testAccOrganizationUpdate = testAccOrganizationAux + `
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
		connection_id = auth0_connection.acme.id
	}
	connections {
		connection_id = auth0_connection.acmeinc.id
		assign_membership_on_login = true
	}
}
`

const testAccOrganizationUpdateAgain = testAccOrganizationAux + `
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
		assign_membership_on_login = false
	}
}
`
