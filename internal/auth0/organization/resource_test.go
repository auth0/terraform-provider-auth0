package organization_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
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

const testAccOrganizationWithTokenQuota = `
resource auth0_organization acme {
	name = "test-quota-{{.testName}}"
	display_name = "Acme Inc. Token Quota {{.testName}}"
	token_quota {
		client_credentials {
			enforce = true
			per_hour = 100
			per_day = 2000
		}
	}
}
`

const testAccOrganizationWithTokenQuotaUpdated = `
resource auth0_organization acme {
	name = "test-quota-{{.testName}}"
	display_name = "Acme Inc. Token Quota {{.testName}}"
	token_quota {
		client_credentials {
			enforce = false
			per_hour = 50
			per_day = 1000
		}
	}
}
`

const testAccOrganizationWithTokenQuotaRemoved = `
resource auth0_organization acme {
	name = "test-quota-{{.testName}}"
	display_name = "Acme Inc. Token Quota {{.testName}}"
	token_quota = null
}
`

func TestAccOrganization(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationEmpty, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "branding.#", "0"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationUpdate, strings.ToLower(t.Name())),
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
				Config: acctest.ParseTestName(testAccOrganizationUpdateAgain, strings.ToLower(t.Name())),
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
				Config: acctest.ParseTestName(testAccOrganizationUpdateAgainAndAgain, strings.ToLower(t.Name())),
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
				Config: acctest.ParseTestName(testAccOrganizationRemoveAllOptionalParams, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "metadata.%", "0"),
				),
			},
		},
	})
}

func TestAccOrganizationTokenQuota(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccOrganizationWithTokenQuota, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-quota-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. Token Quota %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.enforce", "true"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.per_hour", "100"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.per_day", "2000"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationWithTokenQuotaUpdated, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-quota-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. Token Quota %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.enforce", "false"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.per_hour", "50"),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.0.client_credentials.0.per_day", "1000"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccOrganizationWithTokenQuotaRemoved, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization.acme", "name", fmt.Sprintf("test-quota-%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "display_name", fmt.Sprintf("Acme Inc. Token Quota %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_organization.acme", "token_quota.#", "0"),
				),
			},
		},
	})
}
