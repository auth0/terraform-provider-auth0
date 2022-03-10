package auth0

import (
	"fmt"
	"os"
	"testing"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const testAccDataTenantConfig = `
data auth0_tenant current {
}

resource auth0_client my_client {
	name = "Acceptance Test - Tenant Data Source - {{.random}}"
	app_type = "non_interactive"
}

resource auth0_client_grant management_api {
	client_id = auth0_client.my_client.id
	audience = data.auth0_tenant.current.management_api_identifier
	scope = [ "read:insights" ]
}
`

func TestAccDataTenant(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccDataTenantConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_tenant.current", "domain", os.Getenv("AUTH0_DOMAIN")),
					resource.TestCheckResourceAttr("data.auth0_tenant.current", "management_api_identifier", fmt.Sprintf("https://%s/api/v2/", os.Getenv("AUTH0_DOMAIN"))),
				),
			},
		},
	})
}
