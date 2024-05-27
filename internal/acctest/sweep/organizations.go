package sweep

import (
	"context"
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
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var result *multierror.Error
			var from string
			options := []management.RequestOption{
				management.Take(100),
			}
			for {
				if from != "" {
					options = append(options, management.From(from))
				}

				organizationList, err := api.Organization.List(ctx, options...)
				if err != nil {
					return err
				}

				for _, organization := range organizationList.Organizations {
					log.Printf("[DEBUG] ➝ %s", organization.GetName())

					if strings.Contains(organization.GetName(), "test") {
						result = multierror.Append(
							result,
							api.Organization.Delete(ctx, organization.GetID()),
						)

						log.Printf("[DEBUG] ✗ %s", organization.GetName())
					}
				}

				if !organizationList.HasNext() {
					break
				}

				from = organizationList.Next
			}

			return result.ErrorOrNil()
		},
	})
}
