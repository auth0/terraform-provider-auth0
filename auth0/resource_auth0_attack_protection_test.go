package auth0

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAttackProtection(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAttackProtectionCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.method", "standard"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.max_attempts", "10"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.mode", "count_per_identifier_and_ip"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.shields.#", "2"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.pre_login.0.rate", "864000"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.pre_login.0.max_attempts", "100"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.policy", "always_on"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.selected", "auth0"),
				),
			},
			{
				Config: testAttackProtectionUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.shields.0", "admin_notification"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.max_attempts", "11"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.policy", "high_risk"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.selected", "recaptcha_v2"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.providers.0.recaptcha_v2.0.secret", "someSecret"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "bot_detection.0.response.0.providers.0.recaptcha_v2.0.site_key", "someSiteKey"),
				),
			},
		},
	})
}

const testAttackProtectionCreate = `
resource "auth0_attack_protection" "my_protection_tests" {
  breached_password_detection {
    enabled = true
    method  = "standard"
  }
  brute_force_protection {
    enabled      = true
    max_attempts = 10
    mode         = "count_per_identifier_and_ip"
    shields      = ["block", "user_notification"]
  }
  suspicious_ip_throttling {
    enabled   = true
    shields   = ["block","admin_notification"]
    allowlist = ["127.0.0.1"]
    pre_login {
      max_attempts = 100
      rate         = 864000
    }
    pre_user_registration {
      max_attempts = 50
      rate         = 1200
    }
  }
  bot_detection {
    response {
      policy   = "always_on"
      selected = "auth0"
    }
  }
}
`

const testAttackProtectionUpdate = `
resource "auth0_attack_protection" "my_protection_tests" {
  breached_password_detection {
    enabled = false
    method  = "standard"
  }
  brute_force_protection {
    enabled      = true
    max_attempts = 11
    mode         = "count_per_identifier_and_ip"
    shields      = ["block", "user_notification"]
  }
  suspicious_ip_throttling {
    enabled   = true
    shields   = ["admin_notification"]
    allowlist = ["127.0.0.1"]
    pre_login {
      max_attempts = 100
      rate         = 864000
    }
    pre_user_registration {
      max_attempts = 50
      rate         = 1200
    }
  }
  bot_detection {
    response {
      policy   = "high_risk"
      selected = "recaptcha_v2"
      providers {
        recaptcha_v2 {
          secret = "someSecret"
          site_key = "someSiteKey"
        }
      }
    }
  }
}
`
