package networkacl_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

// checkNetworkACLDataSourceExists verifies the data source exists in state
func checkNetworkACLDataSourceExists(resourceName string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		rs, ok := state.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set")
		}
		return nil
	}
}

const testAccGivenANetworkACL = `
resource "auth0_network_acl" "my_acl" {
	description = "Acceptance Test - {{.testName}}"
	active = true
	priority = 10
	rule {
		action {
			block = true
		}
		scope = "tenant"
		match {
			anonymous_proxy = true
		}
	}
}
`

const testAccGivenAComplexNetworkACL = `
resource "auth0_network_acl" "complex_acl" {
	description = "Complex ACL - {{.testName}}"
	active = true
	priority = 5
	rule {
		action {
			allow = true
		}
		scope = "authentication"
		match {
			anonymous_proxy = false
			asns = [1234, 5678]
			geo_country_codes = ["US", "CA"]
			geo_subdivision_codes = ["US-NY", "CA-ON"]
			ipv4_cidrs = ["192.168.1.0/24", "10.0.0.0/8"]
			ipv6_cidrs = ["2001:db8::/32"]
			ja3_fingerprints = ["ja3fingerprint1", "ja3fingerprint2"]
			ja4_fingerprints = ["ja4fingerprint1", "ja4fingerprint2"]
			user_agents = ["Mozilla/5.0", "Chrome/91.0"]
		}
	}
}
`

const testAccDataNetworkACLConfigWithoutID = testAccGivenANetworkACL + `
data "auth0_network_acl" "my_acl" {
	depends_on = [resource.auth0_network_acl.my_acl]
}`

const testAccDataNetworkACLConfigWithInvalidID = testAccGivenANetworkACL + `
data "auth0_network_acl" "my_acl" {
	depends_on = [resource.auth0_network_acl.my_acl]
	id = "acl_invalid_id"
}`

const testAccDataNetworkACLConfigWithValidID = testAccGivenANetworkACL + `
data "auth0_network_acl" "my_acl" {
	depends_on = [resource.auth0_network_acl.my_acl]
	id = resource.auth0_network_acl.my_acl.id
}`

const testAccDataNetworkACLConfigWithComplexACL = testAccGivenAComplexNetworkACL + `
data "auth0_network_acl" "complex_acl" {
	depends_on = [resource.auth0_network_acl.complex_acl]
	id = resource.auth0_network_acl.complex_acl.id
}
`

func TestAccNetworkACLDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccDataNetworkACLConfigWithoutID, t.Name()),
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
			{
				Config: acctest.ParseTestName(testAccDataNetworkACLConfigWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: (404 Not Found|400 Bad Request: Network ACL id is not a valid UUID)`,
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataNetworkACLConfigWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLDataSourceExists("data.auth0_network_acl.my_acl"),
					resource.TestCheckResourceAttrSet("data.auth0_network_acl.my_acl", "description"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.my_acl", "priority", "10"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.my_acl", "rule.0.action.0.block", "true"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.my_acl", "rule.0.scope", "tenant"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.my_acl", "rule.0.match.0.anonymous_proxy", "true"),
				),
			},
		},
	})
}

func TestAccNetworkACLDataSourceComplex(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataNetworkACLConfigWithComplexACL, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLDataSourceExists("data.auth0_network_acl.complex_acl"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "description", fmt.Sprintf("Complex ACL - %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "active", "true"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "priority", "5"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.action.0.allow", "true"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.scope", "authentication"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.anonymous_proxy", "false"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.asns.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.asns.0", "1234"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.asns.1", "5678"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.geo_country_codes.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.geo_country_codes.0", "US"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.geo_country_codes.1", "CA"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.geo_subdivision_codes.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.ipv4_cidrs.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.ipv6_cidrs.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.ja3_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.ja4_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_network_acl.complex_acl", "rule.0.match.0.user_agents.#", "2"),
				),
			},
		},
	})
}
