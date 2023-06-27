package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Connections will run a test sweeper to remove all Auth0 Connections created through tests.
func Connections() {
	resource.AddTestSweepers("auth0_connection", &resource.Sweeper{
		Name: "auth0_connection",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				connectionList, err := api.Connection.List(
					ctx,
					management.IncludeFields("id", "name"),
					management.Page(page),
				)
				if err != nil {
					return err
				}

				for _, connection := range connectionList.Connections {
					log.Printf("[DEBUG] ➝ %s", connection.GetName())

					if strings.Contains(connection.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Connection.Delete(ctx, connection.GetID()),
						)
						log.Printf("[DEBUG] ✗ %s", connection.GetName())
					}
				}
				if !connectionList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
