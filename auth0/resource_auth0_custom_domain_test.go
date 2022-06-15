package auth0

import (
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
)

func init() {
	resource.AddTestSweepers("auth0_custom_domain", &resource.Sweeper{
		Name: "auth0_custom_domain",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			domains, err := api.CustomDomain.List()
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, domain := range domains {
				log.Printf("[DEBUG] ➝ %s", domain.GetDomain())

				if strings.Contains(domain.GetDomain(), "auth.uat.terraform-provider-auth0.com") {
					result = multierror.Append(
						result,
						api.CustomDomain.Delete(domain.GetID()),
					)

					log.Printf("[DEBUG] ✗ %s", domain.GetDomain())
				}
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccCustomDomain(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccCustomDomain, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "domain", "{{.random}}.auth.uat.terraform-provider-auth0.com", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "type", "auth0_managed_certs"),
					resource.TestCheckResourceAttr("auth0_custom_domain.my_custom_domain", "status", "pending_verification"),
				),
			},
		},
	})
}

const testAccCustomDomain = `
resource "auth0_custom_domain" "my_custom_domain" {
  domain = "{{.random}}.auth.uat.terraform-provider-auth0.com"
  type = "auth0_managed_certs"
}
`
