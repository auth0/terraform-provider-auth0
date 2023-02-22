package organization_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

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

const testAccOrganizationEmpty = `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}
`

const testAccOrganizationCreate = testAccOrganizationGiven2Connections + `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}
`

const testAccOrganizationUpdate = testAccOrganizationGiven2Connections + `
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
}
`

const testAccOrganizationUpdateAgain = testAccOrganizationGiven2Connections + `
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
}
`

const testAccOrganizationUpdateAgainAndAgain = testAccOrganizationGiven2Connections + `
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
}
`

const testAccOrganizationRemoveAllOptionalParams = `
resource auth0_organization acme {
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
	metadata = {}
}
`

func TestAccOrganization(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccOrganizationEmpty, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
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
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "1"),
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
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "2"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.some_key", "some_value"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.another_key", "another_value"),
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
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.some_key", "some_value"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationRemoveAllOptionalParams, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
				),
			},
		},
	})
}
