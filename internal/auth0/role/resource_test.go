package role_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRoleEmpty = `
resource "auth0_role" "the_one" {
	name = "The One - Acceptance Test - {{.testName}}"
}
`

const testAccRoleCreate = `
resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One - Acceptance Test"
}
`

const testAccRoleUpdate = `
resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One who will bring peace - Acceptance Test"
}
`

func TestAccRole(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRoleEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One - Acceptance Test"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One who will bring peace - Acceptance Test"),
				),
			},
		},
	})
}
