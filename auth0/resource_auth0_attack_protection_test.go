package auth0

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAttackProtection(t *testing.T) {

	resource.Test(t, resource.TestCase{
		Providers: map[string]terraform.ResourceProvider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: testAttackProtectionCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.method", "standard"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.method", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.max_attempts", "10"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.max_attempts", "10"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.mode", "count_per_identifier_and_ip"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.shields.#", "2"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.pre_login.0.rate", "864000"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.pre_login.0.max_attempts", "100"),
				),
			},
			{
				Config: testAttackProtectionUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "breached_password_detection.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "suspicious_ip_throttling.0.shields.0", "admin_notification"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection_tests", "brute_force_protection.0.max_attempts", "11"),
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
    shields   = ["block", "admin_notification"]
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
}
`
