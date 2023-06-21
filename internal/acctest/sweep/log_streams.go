package sweep

import (
	"context"
	"log"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// LogStreams will run a test sweeper to remove all Auth0 Log Streams created through tests.
func LogStreams() {
	resource.AddTestSweepers("auth0_log_stream", &resource.Sweeper{
		Name: "auth0_log_stream",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}

			logStreams, err := api.LogStream.List(ctx)
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, logStream := range logStreams {
				log.Printf("[DEBUG] ➝ %s", logStream.GetName())

				if strings.Contains(logStream.GetName(), "Test") {
					result = multierror.Append(
						result,
						api.LogStream.Delete(ctx, logStream.GetID()),
					)

					log.Printf("[DEBUG] ✗ %v\n", logStream.GetName())
				}
			}

			return result.ErrorOrNil()
		},
	})
}
