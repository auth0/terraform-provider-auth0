package attackprotection_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
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

		pre_change_password {
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

		pre_change_password {
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
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccBreachedPasswordDetectionEnable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "6"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "true"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionUpdatePartial,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "6"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "admin_notification"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "daily"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.admin_notification_frequency.*", "monthly"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.method", "standard"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_user_registration.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_change_password.0.shields.*", "block"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionUpdateFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "6"),
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
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_change_password.0.shields.*", "block"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.pre_change_password.0.shields.*", "admin_notification"),
				),
			},
			{
				Config: testAccBreachedPasswordDetectionDisable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "brute_force_protection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "suspicious_ip_throttling.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "breached_password_detection.0.%", "6"),
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
	acctest.Test(t, resource.TestCase{
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
	acctest.Test(t, resource.TestCase{
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

// ============================================================================
// BOT DETECTION TESTS
// ============================================================================

const testAccBotDetectionEnable = `
resource "auth0_attack_protection" "my_protection" {
	bot_detection {
		bot_detection_level = "low"
		challenge_password_policy = "when_risky"
		challenge_passwordless_policy = "when_risky"
		challenge_password_reset_policy = "when_risky"
		allowlist = ["192.168.1.1", "10.0.0.0"]
		monitoring_mode_enabled = false
	}
}
`

const testAccBotDetectionUpdate = `
resource "auth0_attack_protection" "my_protection" {
	bot_detection {
		bot_detection_level = "medium"
		challenge_password_policy = "always"
		challenge_passwordless_policy = "never"
		challenge_password_reset_policy = "when_risky"
		allowlist = ["192.168.1.0"]
		monitoring_mode_enabled = true
	}
}
`

func TestAccAttackProtectionBotDetection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccBotDetectionEnable,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.bot_detection_level", "low"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_password_policy", "when_risky"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_passwordless_policy", "when_risky"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_password_reset_policy", "when_risky"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.monitoring_mode_enabled", "false"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "bot_detection.0.allowlist.*", "192.168.1.1"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "bot_detection.0.allowlist.*", "10.0.0.0"),
				),
			},
			{
				Config: testAccBotDetectionUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.bot_detection_level", "medium"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_password_policy", "always"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_passwordless_policy", "never"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.challenge_password_reset_policy", "when_risky"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "bot_detection.0.monitoring_mode_enabled", "true"),
					resource.TestCheckTypeSetElemAttr("auth0_attack_protection.my_protection", "bot_detection.0.allowlist.*", "192.168.1.0"),
				),
			},
		},
	})
}

// ============================================================================
// CAPTCHA TESTS - ALL PROVIDERS IN A SINGLE COMPREHENSIVE FUNCTION
// ============================================================================

// Terraform configurations for all CAPTCHA providers

// --- reCAPTCHA v2 Configurations ---
const testAccCaptchaRecaptchaV2 = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "recaptcha_v2"
		recaptcha_v2 {
			site_key = "test-site-key-v2"
			secret = "test-secret-v2"
		}
	}
}
`

const testAccCaptchaRecaptchaV2Update = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "recaptcha_v2"
		recaptcha_v2 {
			site_key = "updated-site-key-v2"
			secret = "updated-secret-v2"
		}
	}
}
`

// --- reCAPTCHA Enterprise Configurations ---
const testAccCaptchaRecaptchaEnterprise = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "recaptcha_enterprise"
		recaptcha_enterprise {
			site_key = "test-site-key-enterprise"
			api_key = "test-api-key-enterprise"
			project_id = "test-project-id-enterprise"
		}
	}
}
`

const testAccCaptchaRecaptchaEnterpriseUpdate = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "recaptcha_enterprise"
		recaptcha_enterprise {
			site_key = "updated-site-key-enterprise"
			api_key = "updated-api-key-enterprise"
			project_id = "updated-project-id-enterprise"
		}
	}
}
`

// --- hCaptcha Configurations ---
const testAccCaptchaHcaptcha = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "hcaptcha"
		hcaptcha {
			site_key = "test-site-key-hcaptcha"
			secret = "test-secret-hcaptcha"
		}
	}
}
`

const testAccCaptchaHcaptchaUpdate = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "hcaptcha"
		hcaptcha {
			site_key = "updated-site-key-hcaptcha"
			secret = "updated-secret-hcaptcha"
		}
	}
}
`

// --- Friendly Captcha Configurations ---
const testAccCaptchaFriendlyCaptcha = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "friendly_captcha"
		friendly_captcha {
			site_key = "test-site-key-friendly"
			secret = "test-secret-friendly"
		}
	}
}
`

const testAccCaptchaFriendlyCaptchaUpdate = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "friendly_captcha"
		friendly_captcha {
			site_key = "updated-site-key-friendly"
			secret = "updated-secret-friendly"
		}
	}
}
`

// --- Arkose Labs Configurations ---
const testAccCaptchaArkose = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "arkose"
		arkose {
			site_key = "test-site-key-arkose"
			secret = "test-secret-arkose"
			client_subdomain = "client-api"
			verify_subdomain = "verify-api"
			fail_open = false
		}
	}
}
`

const testAccCaptchaArkoseUpdate = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "arkose"
		arkose {
			site_key = "updated-site-key-arkose"
			secret = "updated-secret-arkose"
			client_subdomain = "updated-client-api"
			verify_subdomain = "updated-verify-api"
			fail_open = true
		}
	}
}
`

// --- Auth Challenge Configurations ---
const testAccCaptchaAuthChallenge = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "auth_challenge"
		auth_challenge {
			fail_open = false
		}
	}
}
`

const testAccCaptchaAuthChallengeUpdate = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "auth_challenge"
		auth_challenge {
			fail_open = true
		}
	}
}
`

// --- Simple Captcha (Auth0 v1) Configurations ---
const testAccCaptchaSimpleCaptcha = `
resource "auth0_attack_protection" "my_protection" {
	captcha {
		active_provider_id = "simple_captcha"
	}
}
`

// TestAccAttackProtectionCaptcha is a comprehensive test covering all CAPTCHA providers.
// Each provider is tested with its own set of steps including creation and update scenarios.
func TestAccAttackProtectionCaptcha(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// =================================================================
			// STEP 1: Test reCAPTCHA v2 - Basic Setup
			// =================================================================
			{
				Config: testAccCaptchaRecaptchaV2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "recaptcha_v2"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_v2.0.site_key", "test-site-key-v2"),
				),
			},
			// STEP 2: Test reCAPTCHA v2 - Update Credentials
			{
				Config: testAccCaptchaRecaptchaV2Update,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "recaptcha_v2"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_v2.0.site_key", "updated-site-key-v2"),
				),
			},

			// =================================================================
			// STEP 3: Test reCAPTCHA Enterprise - Basic Setup
			// =================================================================
			{
				Config: testAccCaptchaRecaptchaEnterprise,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "recaptcha_enterprise"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_enterprise.0.site_key", "test-site-key-enterprise"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_enterprise.0.project_id", "test-project-id-enterprise"),
				),
			},
			// STEP 4: Test reCAPTCHA Enterprise - Update All Credentials
			{
				Config: testAccCaptchaRecaptchaEnterpriseUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "recaptcha_enterprise"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_enterprise.0.site_key", "updated-site-key-enterprise"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.recaptcha_enterprise.0.project_id", "updated-project-id-enterprise"),
				),
			},

			// =================================================================
			// STEP 5: Test hCaptcha - Basic Setup
			// =================================================================
			{
				Config: testAccCaptchaHcaptcha,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "hcaptcha"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.hcaptcha.0.site_key", "test-site-key-hcaptcha"),
				),
			},
			// STEP 6: Test hCaptcha - Update Credentials
			{
				Config: testAccCaptchaHcaptchaUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "hcaptcha"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.hcaptcha.0.site_key", "updated-site-key-hcaptcha"),
				),
			},

			// =================================================================
			// STEP 7: Test Friendly Captcha - Basic Setup
			// =================================================================
			{
				Config: testAccCaptchaFriendlyCaptcha,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "friendly_captcha"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.friendly_captcha.0.site_key", "test-site-key-friendly"),
				),
			},
			// STEP 8: Test Friendly Captcha - Update Credentials
			{
				Config: testAccCaptchaFriendlyCaptchaUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "friendly_captcha"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.friendly_captcha.0.site_key", "updated-site-key-friendly"),
				),
			},

			// =================================================================
			// STEP 9: Test Arkose Labs - Basic Setup with All Fields
			// =================================================================
			{
				Config: testAccCaptchaArkose,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "arkose"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.site_key", "test-site-key-arkose"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.client_subdomain", "client-api"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.verify_subdomain", "verify-api"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.fail_open", "false"),
				),
			},
			// STEP 10: Test Arkose Labs - Update All Fields Including Boolean
			{
				Config: testAccCaptchaArkoseUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "arkose"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.site_key", "updated-site-key-arkose"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.client_subdomain", "updated-client-api"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.verify_subdomain", "updated-verify-api"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.arkose.0.fail_open", "true"),
				),
			},

			// =================================================================
			// STEP 11: Test Auth Challenge - Basic Setup
			// =================================================================
			{
				Config: testAccCaptchaAuthChallenge,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "auth_challenge"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.auth_challenge.0.fail_open", "false"),
				),
			},
			// STEP 12: Test Auth Challenge - Update Boolean Flag
			{
				Config: testAccCaptchaAuthChallengeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "auth_challenge"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.auth_challenge.0.fail_open", "true"),
				),
			},

			// =================================================================
			// STEP 13: Test Simple Captcha (Auth0 v1) - No Configuration Needed
			// =================================================================
			{
				Config: testAccCaptchaSimpleCaptcha,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.#", "1"),
					resource.TestCheckResourceAttr("auth0_attack_protection.my_protection", "captcha.0.active_provider_id", "simple_captcha"),
				),
			},
		},
	})
}
