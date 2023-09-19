package sweep

import (
	"context"
	"log"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Users will run a test sweeper to remove all Auth0 Users created through tests.
func Users() {
	resource.AddTestSweepers("auth0_user", &resource.Sweeper{
		Name: "auth0_user",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				userList, err := api.User.Search(
					ctx,
					management.Page(page),
					management.Query(`email.domain:"acceptance.test.com"`))
				if err != nil {
					return err
				}

				for _, user := range userList.Users {
					result = multierror.Append(
						result,
						api.User.Delete(ctx, user.GetID()),
					)
					log.Printf("[DEBUG] ✗ %s", user.GetName())
				}
				if !userList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
