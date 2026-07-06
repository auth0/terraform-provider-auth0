package ratelimitpolicy_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRateLimitPolicyBlock = `
resource "auth0_rate_limit_policy" "my_policy" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "default"

  configuration {
    action = "block"
    limit  = 100
  }
}
`

const testAccRateLimitPolicyRedirect = `
resource "auth0_rate_limit_policy" "my_policy" {
  resource          = "oauth_authentication_api"
  consumer          = "client"
  consumer_selector = "default"

  configuration {
    action       = "redirect"
    limit        = 500
    redirect_uri = "https://example.com/rate-limited"
  }
}
`

func TestAccRateLimitPolicy(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitPolicyBlock,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "resource", "oauth_authentication_api"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "consumer", "client"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "consumer_selector", "default"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "configuration.0.action", "block"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "configuration.0.limit", "100"),
					resource.TestCheckResourceAttrSet("auth0_rate_limit_policy.my_policy", "created_at"),
				),
			},
			{
				Config: testAccRateLimitPolicyRedirect,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "configuration.0.action", "redirect"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "configuration.0.limit", "500"),
					resource.TestCheckResourceAttr("auth0_rate_limit_policy.my_policy", "configuration.0.redirect_uri", "https://example.com/rate-limited"),
				),
			},
			{
				ResourceName:      "auth0_rate_limit_policy.my_policy",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
