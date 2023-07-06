package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// RuleConfigs will run a test sweeper to remove all Auth0 Rule Configs created through tests.
func RuleConfigs() {
	resource.AddTestSweepers("auth0_rule_config", &resource.Sweeper{
		Name: "auth0_rule_config",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			configurations, err := api.RuleConfig.List(ctx)
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, c := range configurations {
				log.Printf("[DEBUG] ➝ %s", c.GetKey())
				if strings.Contains(c.GetKey(), "test") {
					result = multierror.Append(
						result,
						api.RuleConfig.Delete(ctx, c.GetKey()),
					)
					log.Printf("[DEBUG] ✗ %s", c.GetKey())
				}
			}

			return result.ErrorOrNil()
		},
	})
}
