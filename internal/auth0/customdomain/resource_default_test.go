package customdomain_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateCustomDomainDefault = `
resource "auth0_custom_domain_default" "default" {
	domain = "aranusrii.acmetest.org"
}
`

const testAccUpdateCustomDomainDefaultNonVerified = `
resource "auth0_custom_domain_default" "default" {
	domain = "aranusrii2.acmetest.org"
}
`

func TestAccCustomDomainDefault(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateCustomDomainDefault, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_custom_domain_default.default", "domain", "aranusrii.acmetest.org"),
				),
			},
			{
				Config:      acctest.ParseTestName(testAccUpdateCustomDomainDefaultNonVerified, t.Name()),
				ExpectError: regexp.MustCompile("The domain must be verified before it can be set as default."),
			},
		},
	})
}
