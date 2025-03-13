package networkacl_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccNetworkACLCreateWithRequiredFields = `
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

const testAccNetworkACLUpdateWithAllFields = `
resource "auth0_network_acl" "my_acl" {
	description = "Updated Acceptance Test - {{.testName}}"
	active = false
	priority = 2
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

const testAccNetworkACLUpdateWithNotMatch = `
resource "auth0_network_acl" "my_acl" {
	description = "NotMatch Acceptance Test - {{.testName}}"
	active = true
	priority = 3
	rule {
		action {
			log = true
		}
		scope = "management"
		not_match {
			anonymous_proxy = true
			asns = [9876]
			geo_country_codes = ["UK"]
		}
	}
}
`

const testAccNetworkACLUpdateWithRedirect = `
resource "auth0_network_acl" "my_acl" {
	description = "Redirect Acceptance Test - {{.testName}}"
	active = true
	priority = 4
	rule {
		action {
			redirect = true
			redirect_uri = "https://example.com/blocked"
		}
		scope = "tenant"
		match {
			anonymous_proxy = true
		}
	}
}
`

// checkNetworkACLExists verifies the resource exists in Auth0
func checkNetworkACLExists(resourceName string) resource.TestCheckFunc {
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

func TestAccNetworkACL(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNetworkACLCreateWithRequiredFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLExists("auth0_network_acl.my_acl"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.block", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "tenant"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.anonymous_proxy", "true"),
					resource.TestCheckResourceAttrSet("auth0_network_acl.my_acl", "id"),
				),
			},
			{
				ResourceName:      "auth0_network_acl.my_acl",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLUpdateWithAllFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Updated Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "false"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.allow", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "authentication"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.anonymous_proxy", "false"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.asns.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.asns.0", "1234"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.asns.1", "5678"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.geo_country_codes.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.geo_country_codes.0", "US"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.geo_country_codes.1", "CA"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.geo_subdivision_codes.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.ipv4_cidrs.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.ipv6_cidrs.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.ja3_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.ja4_fingerprints.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.user_agents.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLUpdateWithNotMatch, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("NotMatch Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "3"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.log", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "management"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.anonymous_proxy", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.asns.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.asns.0", "9876"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.geo_country_codes.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.geo_country_codes.0", "UK"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLUpdateWithRedirect, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Redirect Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "4"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.redirect", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.redirect_uri", "https://example.com/blocked"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "tenant"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.anonymous_proxy", "true"),
				),
			},
		},
	})
}

// Test validation errors
const testAccNetworkACLInvalidAction = `
resource "auth0_network_acl" "my_acl" {
	description = "Invalid Action - {{.testName}}"
	active = true
	priority = 10
	rule {
		action {
			block = true
			allow = true
		}
		scope = "tenant"
		match {
			anonymous_proxy = true
		}
	}
}
`

const testAccNetworkACLMissingMatch = `
resource "auth0_network_acl" "my_acl" {
	description = "Missing Match - {{.testName}}"
	active = true
	priority = 10
	rule {
		action {
			block = true
		}
		scope = "tenant"
		# Missing both match and not_match
	}
}
`

const testAccNetworkACLMissingRedirectURI = `
resource "auth0_network_acl" "my_acl" {
	description = "Missing Redirect URI - {{.testName}}"
	active = true
	priority = 10
	rule {
		action {
			redirect = true
			# Missing redirect_uri
		}
		scope = "tenant"
		match {
			anonymous_proxy = true
		}
	}
}
`

const testAccNetworkACLInvalidScope = `
resource "auth0_network_acl" "my_acl" {
	description = "Invalid Scope - {{.testName}}"
	active = true
	priority = 10
	rule {
		action {
			block = true
		}
		scope = "invalid_scope"
		match {
			anonymous_proxy = true
		}
	}
}
`

const testAccNetworkACLMaxValues = `
resource "auth0_network_acl" "max_acl" {
	description = "Max Values Test - {{.testName}}"
	active = true
	priority = 9
	rule {
		action {
			block = true
		}
		scope = "tenant"
		match {
			anonymous_proxy = true
			asns = [1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010]
			geo_country_codes = ["US", "CA", "UK", "DE", "FR", "JP", "AU", "BR", "IN", "CN"]
			geo_subdivision_codes = ["US-NY", "US-CA", "US-TX", "CA-ON", "CA-QC", "UK-LND", "DE-BE", "FR-75", "JP-13", "AU-NSW"]
			ipv4_cidrs = ["192.168.1.0/24", "10.0.0.0/8", "172.16.0.0/12", "100.64.0.0/10", "169.254.0.0/16", "192.0.0.0/24", "192.0.2.0/24", "192.88.99.0/24", "198.18.0.0/15", "203.0.113.0/24"]
			ipv6_cidrs = ["2001:db8::/32", "2001:0db8:0000:0000:0000:0000:0000:0000/32", "2001:0db8:1234::/48", "2001:0db8:5678::/48", "2001:0db8:9abc::/48", "2001:0db8:def0::/48", "2001:0db8:1111::/48", "2001:0db8:2222::/48", "2001:0db8:3333::/48", "2001:0db8:4444::/48"]
			ja3_fingerprints = ["ja3_1", "ja3_2", "ja3_3", "ja3_4", "ja3_5", "ja3_6", "ja3_7", "ja3_8", "ja3_9", "ja3_10"]
			ja4_fingerprints = ["ja4_1", "ja4_2", "ja4_3", "ja4_4", "ja4_5", "ja4_6", "ja4_7", "ja4_8", "ja4_9", "ja4_10"]
			user_agents = ["Mozilla/5.0", "Chrome/91.0", "Safari/14.0", "Firefox/89.0", "Edge/91.0", "Opera/76.0", "IE/11.0", "Android/11.0", "iOS/14.0", "Bot/2.0"]
		}
	}
}
`

func TestAccNetworkACLValidation(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccNetworkACLInvalidAction, t.Name()),
				ExpectError: regexp.MustCompile("None of the valid schemas were met"),
			},
			{
				Config:      acctest.ParseTestName(testAccNetworkACLMissingMatch, t.Name()),
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config:      acctest.ParseTestName(testAccNetworkACLMissingRedirectURI, t.Name()),
				ExpectError: regexp.MustCompile("Missing required property: redirect_uri"),
			},
			{
				Config:      acctest.ParseTestName(testAccNetworkACLInvalidScope, t.Name()),
				ExpectError: regexp.MustCompile("got invalid_scope"),
			},
		},
	})
}

// Test for edge cases and maximum values
func TestAccNetworkACLEdgeCases(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNetworkACLMaxValues, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLExists("auth0_network_acl.max_acl"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "description", fmt.Sprintf("Max Values Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "priority", "9"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.asns.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.geo_country_codes.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.geo_subdivision_codes.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.ipv4_cidrs.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.ipv6_cidrs.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.ja3_fingerprints.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.ja4_fingerprints.#", "10"),
					resource.TestCheckResourceAttr("auth0_network_acl.max_acl", "rule.0.match.0.user_agents.#", "10"),
				),
			},
			{
				ResourceName:      "auth0_network_acl.max_acl",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
