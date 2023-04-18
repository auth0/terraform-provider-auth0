package sweep

import (
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Organizations will run a test sweeper to remove all Auth0 Organizations created through tests.
func Organizations() {
	resource.AddTestSweepers("auth0_organization", &resource.Sweeper{
		Name: "auth0_organization",
		F: func(_ string) error {
			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				organizationList, err := api.Organization.List(management.Page(page))
				if err != nil {
					return err
				}

				for _, organization := range organizationList.Organizations {
					log.Printf("[DEBUG] ➝ %s", organization.GetName())

					if strings.Contains(organization.GetName(), "test") {
						result = multierror.Append(
							result,
							api.Organization.Delete(organization.GetID()),
						)

						log.Printf("[DEBUG] ✗ %s", organization.GetName())
					}
				}
				if !organizationList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
