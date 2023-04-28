package sweep

import (
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Clients will run a test sweeper to remove all Auth0 Clients created through tests.
func Clients() {
	resource.AddTestSweepers("auth0_client", &resource.Sweeper{
		Name: "auth0_client",
		F: func(_ string) error {
			api, err := auth0API()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				clientList, err := api.Client.List(management.Page(page))
				if err != nil {
					return err
				}

				for _, client := range clientList.Clients {
					log.Printf("[DEBUG] ➝ %s", client.GetName())

					if strings.Contains(client.GetName(), "Test") {
						result = multierror.Append(
							result,
							api.Client.Delete(client.GetClientID()),
						)
						log.Printf("[DEBUG] ✗ %s", client.GetName())
					}
				}
				if !clientList.HasNext() {
					break
				}
				page++
			}

			return result.ErrorOrNil()
		},
	})
}
