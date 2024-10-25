package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceProvider = `
provider "auth0" {
	api_token = "dummy_token"
}

data "auth0_provider" "my_provider" {
}
`

func TestFrameworkDataSourceProvider(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		// This is primarily here to allow us to test the regular provider instantiation flow.
		ProtoV6ProviderFactories: acctest.TestProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceProvider, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_provider.my_provider", "provider_version"),
				),
			},
		},
	})
}
