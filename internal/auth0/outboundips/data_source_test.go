package outboundips_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceOutboundIPs = `
data "auth0_outbound_ips" "test" {}
`

func TestAccOutboundIPsDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceOutboundIPs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "id"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "last_updated_at"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "regions.#"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "regions.0.region"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "regions.0.ipv4_cidrs.#"),

					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "changelog.#"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "changelog.0.region"),
					resource.TestCheckResourceAttrSet("data.auth0_outbound_ips.test", "changelog.0.ipv4_cidrs.#"),
				),
			},
		},
	})
}
