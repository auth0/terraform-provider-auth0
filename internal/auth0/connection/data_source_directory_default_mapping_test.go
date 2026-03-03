package connection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDirectoryDefaultMappingDataSource = testAccDirectoryWithDefaults + `
data "auth0_connection_directory_default_mapping" "test" {
	depends_on    = [auth0_connection_directory.my_directory]
	connection_id = auth0_connection.my_connection.id
}
`

func TestAccDirectoryDefaultMappingDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDirectoryDefaultMappingDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory_default_mapping.test", "connection_id"),
					resource.TestCheckResourceAttrSet("data.auth0_connection_directory_default_mapping.test", "mapping.#"),
				),
			},
		},
	})
}
