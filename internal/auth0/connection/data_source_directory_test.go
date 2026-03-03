package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDirectoryDataSourceConfigCreate = testAccDirectoryGivenAConnection + `
resource "auth0_connection_directory" "my_directory" {
	connection_id = auth0_connection.my_connection.id
	mapping {
		auth0 = "email"
		idp   = "primaryEmail"
	}
	mapping {
		auth0 = "external_id"
		idp   = "id"
	}
}

data "auth0_connection_directory" "test" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
}
`

func TestAccDirectoryDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectoryDataSourceConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory.test", "connection_id"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory.test", "connection_name"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory.test", "strategy"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory.test", "synchronize_automatically", "false"),
					resource.TestCheckResourceAttr("data.auth0_connection_directory.test", "mapping.#", "2"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory.test", "updated_at"),
				),
			},
		},
	})
}
