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
			asns = [9453]
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
			asns = [9453]
		}
	}
}
`

const testAccNetworkACLWithNewMatchFields = `
resource "auth0_network_acl" "my_acl" {
	description = "New Fields Match - {{.testName}}"
	active = true
	priority = 2
	rule {
		action {
			block = true
		}
		scope = "tenant"
		match {
			hostnames = ["testA-dev.us.auth0.com", "api.example.com"]
			connecting_ipv4_cidrs = ["192.168.1.0/24", "10.0.0.1"]
			connecting_ipv6_cidrs = ["2001:db8::/32", "::1"]
		}
	}
}
`

const testAccNetworkACLWithNewNotMatchFields = `
resource "auth0_network_acl" "my_acl" {
	description = "New Fields NotMatch - {{.testName}}"
	active = true
	priority = 3
	rule {
		action {
			block = true
		}
		scope = "tenant"
		not_match {
			hostnames = ["testB-dev.us.auth0.com"]
			connecting_ipv4_cidrs = ["203.0.113.0/24"]
			connecting_ipv6_cidrs = ["1001:d08::/32", "::1"]
		}
	}
}
`

const testAccNetworkACLWithBothMatchAndNotMatch = `
resource "auth0_network_acl" "my_acl" {
	description = "Both Match and NotMatch - {{.testName}}"
	active = true
	priority = 4
	rule {
		action {
			block = true
		}
		scope = "tenant"
		match {
			hostnames = ["testC-dev.us.auth0.com"]
			connecting_ipv4_cidrs = ["10.0.0.0/8"]
		}
		not_match {
			hostnames = ["testD-dev.us.auth0.com"]
			connecting_ipv4_cidrs = ["20.0.0.0/8"]
		}
	}
}
`

// checkNetworkACLExists verifies the resource exists in Auth0.
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
				),
			},
		},
	})
}

// Test validation errors.
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
			asns = [9453]
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
			asns = [9453]
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
			asns = [9453]
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
				ExpectError: regexp.MustCompile("only one action type"),
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

func TestAccNetworkACLNewFields(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNetworkACLWithNewMatchFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLExists("auth0_network_acl.my_acl"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("New Fields Match - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.block", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "tenant"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.hostnames.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.hostnames.0", "testA-dev.us.auth0.com"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.hostnames.1", "api.example.com"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv4_cidrs.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv4_cidrs.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv4_cidrs.1", "10.0.0.1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv6_cidrs.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv6_cidrs.0", "2001:db8::/32"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv6_cidrs.1", "::1"),
				),
			},
			{
				ResourceName:      "auth0_network_acl.my_acl",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLWithNewNotMatchFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("New Fields NotMatch - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.hostnames.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.hostnames.0", "testB-dev.us.auth0.com"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv4_cidrs.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv4_cidrs.0", "203.0.113.0/24"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv6_cidrs.#", "2"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv6_cidrs.0", "1001:d08::/32"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv6_cidrs.1", "::1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLWithBothMatchAndNotMatch, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Both Match and NotMatch - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.hostnames.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.hostnames.0", "testC-dev.us.auth0.com"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv4_cidrs.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.connecting_ipv4_cidrs.0", "10.0.0.0/8"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.hostnames.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.hostnames.0", "testD-dev.us.auth0.com"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv4_cidrs.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.connecting_ipv4_cidrs.0", "20.0.0.0/8"),
				),
			},
		},
	})
}

const testAccNetworkACLWithAuth0Managed = `
resource "auth0_network_acl" "my_acl" {
	description = "Auth0 Managed - {{.testName}}"
	active = true
	priority = 7
	rule {
		action {
			block = true
		}
		scope = "authentication"
		match {
			auth0_managed = ["auth0.icloud_relay_proxy"]
		}
	}
}
`

const testAccNetworkACLWithAuth0ManagedNotMatch = `
resource "auth0_network_acl" "my_acl" {
	description = "Auth0 Managed NotMatch - {{.testName}}"
	active = true
	priority = 7
	rule {
		action {
			allow = true
		}
		scope = "authentication"
		not_match {
			auth0_managed = ["auth0.low_reputation"]
		}
	}
}
`

// TestAccNetworkACLAuth0Managed exercises a create/update round-trip of the
// Early Access auth0_managed curated blocklists field on both match and
// not_match.
func TestAccNetworkACLAuth0Managed(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccNetworkACLWithAuth0Managed, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					checkNetworkACLExists("auth0_network_acl.my_acl"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Auth0 Managed - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "active", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "priority", "7"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.block", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.scope", "authentication"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.auth0_managed.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.match.0.auth0_managed.0", "auth0.icloud_relay_proxy"),
				),
			},
			{
				ResourceName:      "auth0_network_acl.my_acl",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testAccNetworkACLWithAuth0ManagedNotMatch, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "description", fmt.Sprintf("Auth0 Managed NotMatch - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.action.0.allow", "true"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.auth0_managed.#", "1"),
					resource.TestCheckResourceAttr("auth0_network_acl.my_acl", "rule.0.not_match.0.auth0_managed.0", "auth0.low_reputation"),
				),
			},
		},
	})
}

const testAccNetworkACLAuth0ManagedInvalidPattern = `
resource "auth0_network_acl" "my_acl" {
	description = "Auth0 Managed Invalid - {{.testName}}"
	active = true
	priority = 7
	rule {
		action {
			block = true
		}
		scope = "authentication"
		match {
			auth0_managed = ["not_a_valid_identifier"]
		}
	}
}
`

const testAccNetworkACLAuth0ManagedMutualExclusivity = `
resource "auth0_network_acl" "my_acl" {
	description = "Auth0 Managed Mutual Exclusivity - {{.testName}}"
	active = true
	priority = 7
	rule {
		action {
			block = true
		}
		scope = "authentication"
		match {
			auth0_managed = ["auth0.icloud_relay_proxy"]
		}
		not_match {
			auth0_managed = ["auth0.low_reputation"]
		}
	}
}
`

// TestAccNetworkACLAuth0ManagedValidation covers the client-side validation for
// the auth0_managed field: the identifier pattern and the mutual-exclusivity
// invariant between match and not_match. Neither case reaches the API.
func TestAccNetworkACLAuth0ManagedValidation(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccNetworkACLAuth0ManagedInvalidPattern, t.Name()),
				ExpectError: regexp.MustCompile("must be an Auth0-curated blocklist identifier"),
			},
			{
				Config:      acctest.ParseTestName(testAccNetworkACLAuth0ManagedMutualExclusivity, t.Name()),
				ExpectError: regexp.MustCompile("'auth0_managed' can only be set on one of 'match' or 'not_match'"),
			},
		},
	})
}

// Test for edge cases and maximum values.
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
