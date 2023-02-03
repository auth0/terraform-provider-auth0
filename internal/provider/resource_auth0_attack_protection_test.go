package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

const testAccBreachedPasswordDetectionEnable = `
resource "auth0_attack_protection" "my_protection" {
	breached_password_detection {
		enabled = true
	}
}
`

const testAccBreachedPasswordDetectionUpdatePartial = `
resource "auth0_attack_protection" "my_protection" {
	breached_password_detection {
		enabled = true
		shields = ["admin_notification","block"]
		admin_notification_frequency = ["daily", "monthly"]
		method = "standard"

		pre_user_registration {
			shields = ["block"]
		}
	}
}
`

const testAccBreachedPasswordDetectionUpdateFull = `
resource "auth0_attack_protection" "my_protection" {
	breached_password_detection {
		enabled = true
		shields = ["user_notification", "block", "admin_notification"]
		admin_notification_frequency = ["daily", "monthly", "immediately", "weekly"]
		method = "standard"

		pre_user_registration {
			shields = ["block", "admin_notification"]
		}
	}
}
`

const testAccBreachedPasswordDetectionDisable = `
resource "auth0_attack_protection" "my_protection" {
	breached_password_detection {
		enabled = false
	}
}
`

func TestAccAttackProtectionBreachedPasswordDetection(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccBreachedPasswordDetectionEnable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "true"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionUpdatePartial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "daily"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "monthly"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.method", "standard"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_user_registration.0.shields.*", "block"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionUpdateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "user_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "daily"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "monthly"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "immediately"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "weekly"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.method", "standard"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_user_registration.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_user_registration.0.shields.*", "admin_notification"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionDisable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "false"),
				),
			},
		},
	})
}

const testAccBruteForceProtectionEnable = `
resource "auth0_attack_protection" "my_protection" {
	brute_force_protection {
		enabled = true
	}
}
`

const testAccBruteForceProtectionUpdatePartial = `
resource "auth0_attack_protection" "my_protection" {
	brute_force_protection {
		enabled = true
		shields = ["block"]
		mode = "count_per_identifier"
	}
}
`

const testAccBruteForceProtectionUpdateFull = `
resource "auth0_attack_protection" "my_protection" {
	brute_force_protection {
		enabled = true
		shields = ["user_notification","block"]
		allowlist = ["127.0.0.1"]
		max_attempts = 5
		mode = "count_per_identifier"
	}
}
`

const testAccBruteForceProtectionDisable = `
resource "auth0_attack_protection" "my_protection" {
	brute_force_protection {
		enabled = false
	}
}
`

func TestAccAttackProtectionBruteForceProtection(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccBruteForceProtectionEnable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.enabled", "true"),
				),
			},
			{
				Config: testAccBruteForceProtectionUpdatePartial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.shields.*", "block"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.mode", "count_per_identifier"),
				),
			},
			{
				Config: testAccBruteForceProtectionUpdateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.shields.*", "user_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.allowlist.*", "127.0.0.1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.mode", "count_per_identifier"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.max_attempts", "5"),
				),
			},
			{
				Config: testAccBruteForceProtectionDisable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.0.enabled", "false"),
				),
			},
		},
	})
}

const testAccSuspiciousIPThrottlingEnable = `
resource "auth0_attack_protection" "my_protection" {
	suspicious_ip_throttling {
		enabled = true
	}
}
`

const testAccSuspiciousIPThrottlingUpdatePartial = `
resource "auth0_attack_protection" "my_protection" {
	suspicious_ip_throttling {
		enabled = true
		shields = ["block"]
		allowlist = ["127.0.0.1"]
		pre_login {
			max_attempts = 5
		}
	}
}
`

const testAccSuspiciousIPThrottlingUpdateFull = `
resource "auth0_attack_protection" "my_protection" {
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
`

const testAccSuspiciousIPThrottlingDisable = `
resource "auth0_attack_protection" "my_protection" {
	suspicious_ip_throttling {
		enabled = false
	}
}
`

func TestAccAttackProtectionSuspiciousIPThrottling(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: testAccSuspiciousIPThrottlingEnable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.enabled", "true"),
				),
			},
			{
				Config: testAccSuspiciousIPThrottlingUpdatePartial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.allowlist.*", "127.0.0.1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.pre_login.0.max_attempts", "5"),
				),
			},
			{
				Config: testAccSuspiciousIPThrottlingUpdateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.allowlist.*", "127.0.0.1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.pre_login.0.max_attempts", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.pre_login.0.rate", "34560"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.pre_user_registration.0.max_attempts", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.pre_user_registration.0.rate", "34561"),
				),
			},
			{
				Config: testAccSuspiciousIPThrottlingDisable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.%", "5"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.0.enabled", "false"),
				),
			},
		},
	})
}
