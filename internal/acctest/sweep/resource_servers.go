package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ResourceServers will run a test sweeper to remove all Auth0 Resource Servers created through tests.
func ResourceServers() {
	resource.AddTestSweepers("auth0_resource_server", &resource.Sweeper{
		Name: "auth0_resource_server",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				resourceServerList, err := api.ResourceServer.List(ctx, management.Page(page))
				if err != nil {
					return err
				}

				for _, server := range resourceServerList.ResourceServers {
					log.Printf("[DEBUG] ➝ %s", server.GetName())

					if strings.Contains(server.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.ResourceServer.Delete(ctx, server.GetID()),
						)

						log.Printf("[DEBUG] ✗ %s", server.GetName())
					}
				}
				if !resourceServerList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
