package sweep

import (
	"log"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ResourceServers will run a test sweeper to remove all Auth0 Resource Servers created through tests.
func ResourceServers() {
	resource.AddTestSweepers("auth0_resource_server", &resource.Sweeper{
		Name: "auth0_resource_server",
		F: func(_ string) error {
			api, err := auth0API()
			if err != nil {
				return err
			}

			fn := func(rs *management.ResourceServer) {
				log.Printf("[DEBUG] ➝ %s", rs.GetName())
				if strings.Contains(rs.GetName(), "Test") {
					if err := api.ResourceServer.Delete(rs.GetID()); err != nil {
						log.Printf("[DEBUG] Failed to delete resource server with ID: %s", rs.GetID())
					}
					log.Printf("[DEBUG] ✗ %s", rs.GetName())
				}
			}

			return api.ResourceServer.Stream(fn, management.IncludeFields("id", "name"))
		},
	})
}
