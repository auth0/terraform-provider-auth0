package customdomain_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccMultipleCustomDomains = `
resource "auth0_custom_domain" "my_custom_domain1" {
  domain = "authninja1.auth.tempdomain.com"
  type   = "self_managed_certs"
}

resource "auth0_custom_domain" "my_custom_domain2" {
  domain = "authninja2.auth.tempdomain.com"
  type   = "self_managed_certs"
}

resource "auth0_custom_domain" "my_custom_domain3" {
  domain = "beacon.auth.tempdomain.com"
  type   = "self_managed_certs"
}
`

const testAccDataSourceCustomDomainsFilter1 = testAccMultipleCustomDomains + `
data "auth0_custom_domains" "filtered" {
  q = "domain:authninja*"
}
`

const testAccDataSourceCustomDomainsFilter2 = testAccMultipleCustomDomains + `
data "auth0_custom_domains" "filtered" {
  q = "domain:beacon*"
}
`

func TestAccDataSourceCustomDomains(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccMultipleCustomDomains, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain1", "domain", "authninja1.auth.tempdomain.com"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain2", "domain", "authninja2.auth.tempdomain.com"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain3", "domain", "beacon.auth.tempdomain.com"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomainsFilter1, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "q", "domain:authninja*"),
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "custom_domains.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_custom_domains.filtered", "custom_domains.*", map[string]string{
						"type":   "self_managed_certs",
						"status": "pending_verification",
					}),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomainsFilter2, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "q", "domain:beacon*"),
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "custom_domains.#", "1"),

					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_custom_domains.filtered", "custom_domains.*", map[string]string{
						"type":   "self_managed_certs",
						"status": "pending_verification",
					}),

					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_custom_domains.filtered", "custom_domains.*", map[string]string{
						"domain": "beacon.auth.tempdomain.com",
					}),
				),
			},
		},
	})
}
