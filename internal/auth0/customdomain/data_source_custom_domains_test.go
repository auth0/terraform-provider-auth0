package customdomain_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCustomDomainFirst = `
resource "auth0_custom_domain" "my_custom_domain1" {
  domain = "authninja1.auth.tempdomain.com"
  type   = "self_managed_certs"
}
`

const testAccCustomDomainSecond = testAccCustomDomainFirst + `
resource "auth0_custom_domain" "my_custom_domain2" {
  domain = "authninja2.auth.tempdomain.com"
  type   = "self_managed_certs"
}
`

const testAccCustomDomainThird = testAccCustomDomainSecond + `
resource "auth0_custom_domain" "my_custom_domain3" {
  domain = "beacon.auth.tempdomain.com"
  type   = "self_managed_certs"
}
`

const testAccDataSourceCustomDomainsFilter1 = `
data "auth0_custom_domains" "filtered" {
  query = "domain:authninja*"
}
`

const testAccDataSourceCustomDomainsFilter2 = `
data "auth0_custom_domains" "filtered" {
  query = "domain:beacon*"
}
`

func TestAccDataSourceCustomDomains(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// We had to split this into two separate posts to work around an issue
				// in the test recording library. We need to add X-Request-Id header to the POST requests
				// to fix this, and make sure that go-vcr uses that to match requests.
				Config: acctest.ParseTestName(testAccCustomDomainFirst, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccCustomDomainSecond, t.Name()),
			},
			{
				Config: testAccCustomDomainThird + testAccDataSourceCustomDomainsFilter1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "query", "domain:authninja*"),
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "custom_domains.#", "2"),

					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_custom_domains.filtered", "custom_domains.*", map[string]string{
						"type":   "self_managed_certs",
						"status": "pending_verification",
					}),
				),
			},
			{
				Config: testAccCustomDomainThird + testAccDataSourceCustomDomainsFilter2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domains.filtered", "query", "domain:beacon*"),
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
