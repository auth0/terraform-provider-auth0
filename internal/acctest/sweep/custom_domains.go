package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// CustomDomains will run a test sweeper to remove all Auth0 Custom Domains created through tests.
func CustomDomains() {
	resource.AddTestSweepers("auth0_custom_domain", &resource.Sweeper{
		Name: "auth0_custom_domain",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			domains, err := api.CustomDomain.List(ctx)
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, domain := range domains {
				log.Printf("[DEBUG] ➝ %s", domain.GetDomain())

				if strings.Contains(domain.GetDomain(), "auth.terraform-provider-auth0.com") {
					result = multierror.Append(
						result,
						api.CustomDomain.Delete(ctx, domain.GetID()),
					)

					log.Printf("[DEBUG] ✗ %s", domain.GetDomain())
				}
			}

			return result.ErrorOrNil()
		},
	})
}
