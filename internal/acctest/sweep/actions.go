package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Actions will run a test sweeper to remove all Auth0 Actions created through tests.
func Actions() {
	resource.AddTestSweepers("auth0_actions", &resource.Sweeper{
		Name: "auth0_actions",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				actionList, err := api.Action.List(ctx, management.Page(page))
				if err != nil {
					return err
				}

				for _, action := range actionList.Actions {
					log.Printf("[DEBUG] ➝ %s", action.GetName())

					if strings.Contains(action.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Action.Delete(ctx, action.GetID()),
						)
						log.Printf("[DEBUG] ✗ %s", action.GetName())
					}
				}
				if !actionList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
