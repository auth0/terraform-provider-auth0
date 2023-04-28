package customdomain_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain = "{{.testName}}.auth.terraform-provider-auth0.com"
	type = "auth0_managed_certs"
	tls_policy = "recommended"
}

data "auth0_custom_domain" "test" {
	depends_on = [ auth0_custom_domain.my_custom_domain ]
}
`

func TestAccDataSourceCustomDomain(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomain, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain", fmt.Sprintf("%s.auth.terraform-provider-auth0.com", testName)),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "status", "pending_verification"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "primary", "true"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "verification.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "tls_policy", "recommended"),
				),
			},
		},
	})
}
