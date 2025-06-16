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
	domain = "{{.testName}}.auth.tempdomain.com"
	type   = "self_managed_certs"
}
`

const testAccUpdateSelfManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain                  = "{{.testName}}.auth.tempdomain.com"
	type                    = "self_managed_certs"
	custom_client_ip_header = "true-client-ip"
}
`

const testAccCreateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.tempdomain.com"
	type   = "auth0_managed_certs"
	domain_metadata = {
        key1: "value1"
		key2: "value2"
    }
}
`

const testAccUpdateAuth0ManagedCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain     = "{{.testName}}.auth.tempdomain.com"
	type       = "auth0_managed_certs"
	tls_policy = "recommended"
	domain_metadata = {
        key1: "value3"
    }
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
						fmt.Sprintf("%s.auth.tempdomain.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
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
						fmt.Sprintf("%s.auth.tempdomain.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "self_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", "true-client-ip"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.tempdomain.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.%", "2"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.key1", "value1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.key2", "value2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateAuth0ManagedCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"auth0_custom_domain.my_custom_domain",
						"domain",
						fmt.Sprintf("%s.auth.tempdomain.com", strings.ToLower(t.Name())),
					),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "verification.#", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "tls_policy", "recommended"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.%", "1"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.key1", "value3"),
					resource.TestCheckNoResourceAttr("auth0_custom_domain.my_custom_domain", "domain_metadata.key2"),
				),
			},
		},
	})
}
