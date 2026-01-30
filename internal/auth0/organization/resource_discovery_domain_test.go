package organization_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccOrganizationDiscoveryDomainCreate = `
resource "auth0_organization" "test_org" {
	name         = "test-{{.testName}}"
	display_name = "Test Organization {{.testName}}"
}

resource "auth0_organization_discovery_domain" "test_domain" {
	organization_id = auth0_organization.test_org.id
	domain         = "{{.testName}}.example.com"
	status         = "pending"
	use_for_organization_discovery = false
}
`

const testAccOrganizationDiscoveryDomainUpdate = `
resource "auth0_organization" "test_org" {
	name         = "test-{{.testName}}"
	display_name = "Test Organization {{.testName}}"
}

resource "auth0_organization_discovery_domain" "test_domain" {
	organization_id = auth0_organization.test_org.id
	domain         = "{{.testName}}.example.com"
	status         = "verified"
	use_for_organization_discovery = true
}
`

func TestAccOrganizationDiscoveryDomain(t *testing.T) {
	testName := strings.ToLower(t.Name())
	domainName := strings.ReplaceAll(testName, "_", "-")

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: strings.ReplaceAll(acctest.ParseTestName(testAccOrganizationDiscoveryDomainCreate, testName), testName+".example.com", domainName+".example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "domain", fmt.Sprintf("%s.example.com", domainName)),
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "status", "pending"),
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "use_for_organization_discovery", "false"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "verification_txt"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "verification_host"),
				),
			},
			{
				Config: strings.ReplaceAll(acctest.ParseTestName(testAccOrganizationDiscoveryDomainUpdate, testName), testName+".example.com", domainName+".example.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "domain", fmt.Sprintf("%s.example.com", domainName)),
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "status", "verified"),
					resource.TestCheckResourceAttr("auth0_organization_discovery_domain.test_domain", "use_for_organization_discovery", "true"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "verification_txt"),
					resource.TestCheckResourceAttrSet("auth0_organization_discovery_domain.test_domain", "verification_host"),
				),
			},
		},
	})
}

const testAccOrganizationDiscoveryDomainImport = `
resource "auth0_organization" "test_org" {
	name         = "test-{{.testName}}"
	display_name = "Test Organization {{.testName}}"
}

resource "auth0_organization_discovery_domain" "test_domain" {
	organization_id = auth0_organization.test_org.id
	domain         = "{{.testName}}.example.com"
	status         = "pending"
}
`

func TestAccOrganizationDiscoveryDomain_Import(t *testing.T) {
	testName := strings.ToLower(t.Name())
	domainName := strings.ReplaceAll(testName, "_", "-")

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: strings.ReplaceAll(acctest.ParseTestName(testAccOrganizationDiscoveryDomainImport, testName), testName+".example.com", domainName+".example.com"),
			},
			{
				ResourceName:      "auth0_organization_discovery_domain.test_domain",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					resourceState := state.RootModule().Resources["auth0_organization_discovery_domain.test_domain"]
					return resourceState.Primary.ID, nil
				},
			},
		},
	})
}
