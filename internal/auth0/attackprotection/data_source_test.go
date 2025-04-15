package attackprotection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceAttackProtection = `
resource "auth0_attack_protection" "my_protection" {
	breached_password_detection {
		enabled = true
		shields = ["admin_notification","block"]
		admin_notification_frequency = ["daily", "monthly"]
		method = "standard"

		pre_user_registration {
			shields = ["block"]
		}

		pre_change_password {
			shields = ["block"]
		}
	}

	brute_force_protection {
		enabled = true
		shields = ["user_notification","block"]
		allowlist = ["127.0.0.1"]
		max_attempts = 5
		mode = "count_per_identifier"
	}

	suspicious_ip_throttling {
		enabled = true
		shields = ["block", "admin_notification"]
		allowlist = ["127.0.0.1"]
		pre_login {
			max_attempts = 5
			rate = 34560
		}
		pre_user_registration {
			max_attempts = 5
			rate = 34561
		}
	}
}

data "auth0_attack_protection" "test" {
	depends_on = [ auth0_attack_protection.my_protection ]
}
`

func TestAccDataSourceAttackProtection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAttackProtection,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "breached_password_detection.0.%", "6"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.admin_notification_frequency.*", "daily"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.admin_notification_frequency.*", "monthly"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "breached_password_detection.0.method", "standard"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.pre_user_registration.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "breached_password_detection.0.pre_change_password.0.shields.*", "block"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "brute_force_protection.0.%", "5"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "brute_force_protection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "brute_force_protection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "brute_force_protection.0.shields.*", "user_notification"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "brute_force_protection.0.allowlist.*", "127.0.0.1"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "brute_force_protection.0.mode", "count_per_identifier"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "brute_force_protection.0.max_attempts", "5"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.%", "5"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.allowlist.*", "127.0.0.1"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.pre_login.0.max_attempts", "5"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.pre_login.0.rate", "34560"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.pre_user_registration.0.max_attempts", "5"),
					resource.TestCheckResourceAttr("data.auth0_attack_protection.test", "suspicious_ip_throttling.0.pre_user_registration.0.rate", "34561"),
				),
			},
		},
	})
}
