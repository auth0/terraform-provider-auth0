package sweep

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Email will run a test sweeper to remove the Auth0 Email Provider created through tests.
func Email() {
	resource.AddTestSweepers("auth0_email", &resource.Sweeper{
		Name: "auth0_email",
		F: func(_ string) error {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return err
			}
			return api.EmailProvider.Delete(ctx)
		},
	})
}

// EmailTemplates will run a test sweeper to remove all Auth0 Email Templates created through tests.
func EmailTemplates() {
	resource.AddTestSweepers("auth0_email_template", &resource.Sweeper{
		Name: "auth0_email_template",
		F: func(_ string) (err error) {
			ctx := context.Background()

			api, err := auth0API()
			if err != nil {
				return
			}
			err = api.EmailTemplate.Update(ctx, "welcome_email", &management.EmailTemplate{
				Enabled: auth0.Bool(false),
			})
			return
		},
	})
}
