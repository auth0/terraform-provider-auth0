package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Roles will run a test sweeper to remove all Auth0 Roles created through tests.
func Roles() {
	resource.AddTestSweepers("auth0_role", &resource.Sweeper{
		Name: "auth0_role",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				roleList, err := api.Role.List(ctx, management.Page(page))
				if err != nil {
					return err
				}

				for _, role := range roleList.Roles {
					log.Printf("[DEBUG] ➝ %s", role.GetName())
					if strings.Contains(role.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Role.Delete(ctx, role.GetID()),
						)
						log.Printf("[DEBUG] ✗ %s", role.GetName())
					}
				}
				if !roleList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
