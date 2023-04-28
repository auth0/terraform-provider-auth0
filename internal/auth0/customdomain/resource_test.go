package customdomain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateSelfManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
}
`

const testAccUpdateSelfManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
	custom_client_ip_header = "true-client-ip"
}
`

const testAccUpdateSelfManagedCustomDomainWithEmptyClientIPHeader = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "self_managed_certs"
	custom_client_ip_header = ""
}
`

const testAccCreateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
}
`

const testAccUpdateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
	tls_policy = "recommended"
}
`

func TestAccCustomDomain(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateSelfManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateSelfManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", "true-client-ip"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateSelfManagedCustomDomainWithEmptyClientIPHeader, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "primary", "true"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
				),
			},
		},
	})
}
