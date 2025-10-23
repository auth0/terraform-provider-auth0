package client_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataClientGrantsByClientID = `
data "auth0_client_grants" "test" {
	depends_on = [auth0_client_grant.my_client_grant]
	client_id = auth0_client.my_client.id
}
`

const testAccDataClientGrantsByAudience = `
data "auth0_client_grants" "test" {
	depends_on = [auth0_client_grant.my_client_grant]
	audience = auth0_resource_server.my_resource_server.identifier
}
`

func TestAccDataSourceClientGrants(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigCreate, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigCreate+testAccDataClientGrantsByClientID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client_grants.test", "client_grants.0.id"),
					resource.TestCheckResourceAttrSet("data.auth0_client_grants.test", "client_grants.0.client_id"),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.audience", fmt.Sprintf("https://uat.tf.terraform-provider-auth0.com/client-grant/%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.scope.0", "create:bar"),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.subject_type", "user"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientGrantConfigCreate+testAccDataClientGrantsByAudience, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_client_grants.test", "client_grants.0.id"),
					resource.TestCheckResourceAttrSet("data.auth0_client_grants.test", "client_grants.0.client_id"),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.audience", fmt.Sprintf("https://uat.tf.terraform-provider-auth0.com/client-grant/%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.scope.0", "create:bar"),
					resource.TestCheckResourceAttr("data.auth0_client_grants.test", "client_grants.0.subject_type", "user"),
				),
			},
		},
	})
}
