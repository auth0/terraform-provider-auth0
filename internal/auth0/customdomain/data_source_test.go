package customdomain_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceCustomDomainSingle = `
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
	depends_on = [resource.auth0_custom_domain.my_custom_domain]
}
`

const testAccDataSourceCustomDomainMultiple = `
resource "auth0_custom_domain" "my_custom_domain1" {
	domain     = "{{.testName}}-first.auth.tempdomain.com"
	type       = "auth0_managed_certs"
	tls_policy = "recommended"
	domain_metadata = {
        key1: "value1"
		key2: "value2"
    }
}

resource "auth0_custom_domain" "my_custom_domain2" {
	domain     = "{{.testName}}-second.auth.tempdomain.com"
	type       = "auth0_managed_certs"
	tls_policy = "recommended"
	domain_metadata = {
        key3: "value3"
    }
}
`

const testAccDataSourceCustomDomainMultipleWithCustomDomainId = testAccDataSourceCustomDomainMultiple + `
data "auth0_custom_domain" "test" {
	depends_on = [resource.auth0_custom_domain.my_custom_domain1,
					resource.auth0_custom_domain.my_custom_domain2]
	custom_domain_id = auth0_custom_domain.my_custom_domain1.id
}
`

const testAccDataSourceCustomDomainMultipleWithOutCustomDomainId = testAccDataSourceCustomDomainMultiple + `
data "auth0_custom_domain" "test" {
	depends_on = [resource.auth0_custom_domain.my_custom_domain1,
					resource.auth0_custom_domain.my_custom_domain2]
}
`

func TestAccDataSourceCustomDomain(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomainSingle, testName),
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
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomainSingle, testName),
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
			{
				Config: acctest.ParseTestName(testAccDataSourceCustomDomainMultipleWithCustomDomainId, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_custom_domain.test", "domain", fmt.Sprintf("%s-first.auth.tempdomain.com", testName)),
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
			{
				Config:      acctest.ParseTestName(testAccDataSourceCustomDomainMultipleWithOutCustomDomainId, testName),
				ExpectError: regexp.MustCompile("multiple custom domains found, please specify custom_domain_id"),
			},
		},
	})
}
