package organization_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationDiscoveryDomainsPreventErasingDomainsOnCreate = testAccGivenTwoDomainsAndAnOrganization + `
resource "auth0_organization_discovery_domains" "existing" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	discovery_domains {
		domain = "{{.testName}}-existing.example.com"
		status = "pending"
	}
}

resource "auth0_organization_discovery_domains" "one_to_many" {
	depends_on = [ auth0_organization_discovery_domains.existing ]

	organization_id = auth0_organization.my_org.id

	discovery_domains {
		domain = "{{.testName}}-new1.example.com"
		status = "pending"
	}

	discovery_domains {
		domain = "{{.testName}}-new2.example.com"
		status = "pending"
	}
}
`

const testAccOrganizationDiscoveryDomainsDelete = testAccGivenTwoDomainsAndAnOrganization + `
data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationDiscoveryDomainsWithOneDomain = testAccGivenTwoDomainsAndAnOrganization + `
resource "auth0_organization_discovery_domains" "one_to_many" {
	organization_id = auth0_organization.my_org.id

	discovery_domains {
		domain = "{{.testName}}-domain1.example.com"
		status = "pending"
	}
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_discovery_domains.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccOrganizationDiscoveryDomainsWithTwoDomains = testAccGivenTwoDomainsAndAnOrganization + `
resource "auth0_organization_discovery_domains" "one_to_many" {
	organization_id = auth0_organization.my_org.id

	discovery_domains {
		domain = "{{.testName}}-domain1.example.com"
		status = "pending"
	}

	discovery_domains {
		domain = "{{.testName}}-domain2.example.com"
		status = "verified"
	}
}

data "auth0_organization" "my_org" {
	depends_on = [ auth0_organization_discovery_domains.one_to_many ]

	organization_id = auth0_organization.my_org.id
}
`

const testAccGivenTwoDomainsAndAnOrganization = `
resource "auth0_organization" "my_org" {
	name         = "test-dd-{{.testName}}"
	display_name = "Test Discovery Domains {{.testName}}"
}
`

const testAccGivenTwoDomainsAndAnOrganizationForImport = `
resource "auth0_organization" "my_org" {
	name         = "test-dd-import-{{.testName}}"
	display_name = "Test Discovery Domains Import {{.testName}}"
}
`

const testAccOrganizationDiscoveryDomainsImportSetup = testAccGivenTwoDomainsAndAnOrganizationForImport + `
resource "auth0_organization_discovery_domain" "domain1" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id
	domain         = "{{.testName}}-domain1.example.com"
	status         = "pending"
}

resource "auth0_organization_discovery_domain" "domain2" {
	depends_on = [ auth0_organization_discovery_domain.domain1 ]

	organization_id = auth0_organization.my_org.id
	domain         = "{{.testName}}-domain2.example.com"
	status         = "verified"
}
`

const testAccOrganizationDiscoveryDomainsImportCheck = testAccOrganizationDiscoveryDomainsImportSetup + `
resource "auth0_organization_discovery_domains" "import_target" {
	depends_on = [ auth0_organization.my_org ]

	organization_id = auth0_organization.my_org.id

	discovery_domains {
		domain = "{{.testName}}-domain1.example.com"
		status = "pending"
	}

	discovery_domains {
		domain = "{{.testName}}-domain2.example.com"
		status = "verified"
	}
}
`

func TestAccOrgDiscoveryDomains(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test 1: Create resource with one domain first.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsWithOneDomain, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domains.one_to_many", "discovery_domains.#", "1"),
				),
			},
			// Test 2: Verify guard function prevents creating domains when org already has domains.
			{
				Config:      acctest.ParseTestName(testAccOrganizationDiscoveryDomainsPreventErasingDomainsOnCreate, testName),
				ExpectError: regexp.MustCompile("Organization with non empty enabled discovery domains"),
			},
			// Test 3: Clean slate - delete all domains.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsDelete, testName),
			},
			{
				RefreshState: true,
				Check:        resource.ComposeTestCheckFunc(
				// No discovery domains resource exists, so nothing to check.
				),
			},
			// Test 4: Create resource with one domain again.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsWithOneDomain, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domains.one_to_many", "discovery_domains.#", "1"),
				),
			},
			// Test 5: Update to two domains.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsWithTwoDomains, testName),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domains.one_to_many", "discovery_domains.#", "2"),
				),
			},
		},
	})
}

func TestAccOrgDiscoveryDomainsImport(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test 1: Create individual discovery domain resources.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsImportSetup, testName),
			},
			// Test 2: Import the domains into the plural resource.
			{
				Config:       acctest.ParseTestName(testAccOrganizationDiscoveryDomainsImportCheck, testName),
				ResourceName: "auth0_organization_discovery_domains.import_target",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return acctest.ExtractResourceAttributeFromState(state, "auth0_organization.my_org", "id")
				},
				ImportStatePersist: true,
			},
			// Test 3: Verify import worked correctly with no plan changes.
			{
				Config: acctest.ParseTestName(testAccOrganizationDiscoveryDomainsImportCheck, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domains.import_target", "discovery_domains.#", "2"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domains.import_target", "organization_id"),
				),
			},
		},
	})
}
