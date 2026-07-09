package ratelimitpolicy_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRateLimitPolicyDataSourceSingular = `
resource "auth0_rate_limit_policy" "my_policy" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "default"

  configuration {
    action = "block"
    limit  = 100
  }
}

data "auth0_rate_limit_policy" "test" {
  policy_id = auth0_rate_limit_policy.my_policy.id
}
`

const testAccRateLimitPolicyDataSourcePlural = `
resource "auth0_rate_limit_policy" "my_policy" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "default"

  configuration {
    action = "block"
    limit  = 100
  }
}

data "auth0_rate_limit_policies" "test" {
  resource   = "oauth_authentication_api"
  depends_on = [auth0_rate_limit_policy.my_policy]
}
`

func TestAccRateLimitPolicyDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitPolicyDataSourceSingular,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_rate_limit_policy.test", "consumer_selector", "default"),
					resource.TestCheckResourceAttr("data.auth0_rate_limit_policy.test", "configuration.0.action", "block"),
					resource.TestCheckResourceAttr("data.auth0_rate_limit_policy.test", "configuration.0.limit", "100"),
				),
			},
			{
				Config: testAccRateLimitPolicyDataSourcePlural,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_rate_limit_policies.test", "rate_limit_policies.#"),
				),
			},
		},
	})
}
