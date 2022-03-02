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
				Config: testAccAttackProtectionCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "suspicious_ip_throttling.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "brute_force_protection.0.enabled", "false"),
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "breached_password_detection.0.enabled", "true"),
				),
			},
			{
				Config: testAccAttackProtectionConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "brute_force_protection.0.enabled", "true"),
					resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "breached_password_detection.0.enabled", "true"),
				),
			},
			// {
			// 	Config: random.Template(testAccAttackProtectionConfigUpdateAgain, rand),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		random.TestCheckResourceAttr("auth0_attack_protection.acc_test", "name", "Test Action {{.random}}", rand),
			// 		resource.TestCheckResourceAttrSet("auth0_attack_protection.acc_test", "version_id"),
			// 		resource.TestCheckResourceAttr("auth0_attack_protection.acc_test", "secrets.#", "0"),
			// 	),
			// },
		},
	})
}

const testAccAttackProtectionCreate = `

resource "auth0_attack_protection" "acc_test" {
	suspicious_ip_throttling {
	  enabled   = false
	}
	brute_force_protection {
	  enabled      = false
	}
	breached_password_detection {
	  admin_notification_frequency = ["daily"]
	  enabled                      = true
	  shields                      = ["admin_notification"]
	}
  }
`

const testAccAttackProtectionConfigUpdate = `

resource "auth0_attack_protection" "attack_protection" {
	suspicious_ip_throttling {
	  enabled   = true
	  shields   = ["admin_notification", "block"]
	  allowlist = ["192.168.1.1"]
	  pre_login {
		max_attempts = 100
		rate         = 864000
	  }
	  pre_user_registration {
		max_attempts = 50
		rate         = 1200
	  }
	}
	brute_force_protection {
	  allowlist    = ["127.0.0.1"]
	  enabled      = true
	  max_attempts = 5
	  mode         = "count_per_identifier_and_ip"
	  shields      = ["block", "user_notification"]
	}
	breached_password_detection {
	  admin_notification_frequency = ["daily"]
	  enabled                      = true
	  method                       = "standard"
	  shields                      = ["admin_notification", "block"]
	}
  }
`
