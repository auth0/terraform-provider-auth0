package customdomain_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
	domain     = "{{.testName}}.auth.tempdomain.com"
	type       = "auth0_managed_certs"
	tls_policy = "recommended"
	domain_metadata = {
        key1: "value1"
		key2: "value2"
    }
}

data "auth0_custom_domain" "test" {
	custom_domain_id = auth0_custom_domain.my_custom_domain.id
}
`

func TestAccDataSourceCustomDomain(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_custom_domain" "test" {}`,
				ExpectError: regexp.MustCompile("Missing required argument"),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomain, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain", fmt.Sprintf("%s.auth.tempdomain.com", testName)),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "status", "pending_verification"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "origin_domain_name", ""),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "custom_client_ip_header", ""),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "tls_policy", "recommended"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "verification.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain_metadata.%", "2"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain_metadata.key1", "value1"),
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain_metadata.key2", "value2"),
				),
			},
		},
	})
}
