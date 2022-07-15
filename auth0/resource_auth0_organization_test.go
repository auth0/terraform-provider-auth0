package auth0

import (
	"fmt"
	"log"
	"math/rand"
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

func TestAccOrganizationAssignUsers(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	testName := strings.ToLower(t.Name() + fmt.Sprintf("%d", rand.Intn(99))) // TODO: REMOVE! This is only here to get around annoying 409s

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{{
			Config: template.ParseTestName(testAccOrganizationAssignUsersCreateOrg, testName),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("auth0_organization.some_org", "name", fmt.Sprintf("some-org-%s", testName)),
				resource.TestCheckResourceAttr("auth0_organization.some_org", "members.#", "0"),
			),
		},
			{
				Config: template.ParseTestName(
					testAccOrganizationAssignUsersAux+`
					resource auth0_organization some_org{
						depends_on = [ auth0_user.user1, auth0_user.user2]
						name = "some-org-{{.testName}}"
						display_name = "{{.testName}}"
						members = [ auth0_user.user1.id ]
					}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.some_org", "name", fmt.Sprintf("some-org-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization.some_org", "members.#", "1"),
				),
			},
			{
				Config: template.ParseTestName(
					testAccOrganizationAssignUsersAux+`
					resource auth0_organization some_org{
						depends_on = [ auth0_user.user1, auth0_user.user2]
						name = "some-org-{{.testName}}"
						display_name = "{{.testName}}"
						members = [ auth0_user.user1.id, auth0_user.user2.id ]
					}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.some_org", "name", fmt.Sprintf("some-org-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization.some_org", "members.#", "2"),
				),
			},
			{
				Config: template.ParseTestName(
					testAccOrganizationAssignUsersAux+`
					resource auth0_organization some_org{
						depends_on = [ auth0_user.user1, auth0_user.user2]
						name = "some-org-{{.testName}}"
						display_name = "{{.testName}}"
						members = []
					}
				`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.some_org", "name", fmt.Sprintf("some-org-%s", testName)),
					resource.TestCheckResourceAttr("auth0_organization.some_org", "members.#", "0"),
				),
			},
		},
	})
}

const testAccOrganizationAssignUsersAux = `
resource auth0_user user1 {
	email = "will.vedder+1{{.testName}}@auth0.com"
	connection_name = "Username-Password-Authentication"
	email_verified = true
	password = "MyPass123$"
}
	
resource auth0_user user2 {
	email = "will.vedder+2{{.testName}}@auth0.com"
	connection_name = "Username-Password-Authentication"
	email_verified = true
	password = "MyPass123$"
}
`

const testAccOrganizationAssignUsersCreateOrg = testAccOrganizationAssignUsersAux +
	`
resource auth0_organization some_org{
	name = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`
