package supplementalsignals_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccSupplementalSignalsCreate = `
resource "auth0_supplemental_signals" "test" {
	akamai_enabled = true
}
`

const testAccSupplementalSignalsUpdate = `
resource "auth0_supplemental_signals" "test" {
	akamai_enabled = false
}
`

func TestAccSupplementalSignals(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccSupplementalSignalsCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_supplemental_signals.test", "akamai_enabled", "true"),
				),
			},
			{
				Config: testAccSupplementalSignalsUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_supplemental_signals.test", "akamai_enabled", "false"),
				),
			},
		},
	})
}
