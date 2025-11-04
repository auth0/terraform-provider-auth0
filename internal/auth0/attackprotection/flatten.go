package attackprotection

import (
	"github.com/auth0/go-auth0/management"
	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenSuspiciousIPThrottling(ipt *management.SuspiciousIPThrottling) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"enabled":   ipt.GetEnabled(),
			"allowlist": ipt.GetAllowList(),
			"shields":   ipt.GetShields(),
			"pre_login": []interface{}{
				map[string]int{
					"max_attempts": ipt.GetStage().GetPreLogin().GetMaxAttempts(),
					"rate":         ipt.GetStage().GetPreLogin().GetRate(),
				},
			},
			"pre_user_registration": []interface{}{
				map[string]int{
					"max_attempts": ipt.GetStage().GetPreUserRegistration().GetMaxAttempts(),
					"rate":         ipt.GetStage().GetPreUserRegistration().GetRate(),
				},
			},
		},
	}
}

func flattenBruteForceProtection(bfp *management.BruteForceProtection) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"enabled":      bfp.GetEnabled(),
			"mode":         bfp.GetMode(),
			"max_attempts": bfp.GetMaxAttempts(),
			"shields":      bfp.GetShields(),
			"allowlist":    bfp.GetAllowList(),
		},
	}
}

func flattenBreachedPasswordProtection(bpd *management.BreachedPasswordDetection) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"enabled":                      bpd.GetEnabled(),
			"method":                       bpd.GetMethod(),
			"admin_notification_frequency": bpd.GetAdminNotificationFrequency(),
			"shields":                      bpd.GetShields(),
			"pre_user_registration": []interface{}{
				map[string][]string{
					"shields": bpd.GetStage().GetPreUserRegistration().GetShields(),
				},
			},
			"pre_change_password": []interface{}{
				map[string][]string{
					"shields": bpd.GetStage().GetPreChangePassword().GetShields(),
				},
			},
		},
	}
}

func flattenBotDetection(botDetection *managementv2.GetBotDetectionSettingsResponseContent) []interface{} {
	if botDetection == nil {
		return nil
	}

	m := make(map[string]interface{})

	// BotDetectionLevel
	m["bot_detection_level"] = string(botDetection.GetBotDetectionLevel())

	// ChallengePasswordPolicy
	m["challenge_password_policy"] = string(botDetection.GetChallengePasswordPolicy())

	// ChallengePasswordlessPolicy
	m["challenge_passwordless_policy"] = string(botDetection.GetChallengePasswordlessPolicy())

	// ChallengePasswordResetPolicy
	m["challenge_password_reset_policy"] = string(botDetection.GetChallengePasswordResetPolicy())

	// Allowlist (it's a slice, not a pointer)
	m["allowlist"] = botDetection.Allowlist

	// MonitoringModeEnabled (it's a bool, not a pointer)
	m["monitoring_mode_enabled"] = botDetection.MonitoringModeEnabled

	return []interface{}{m}
}

func flattenCaptcha(data *schema.ResourceData, captcha *managementv2.GetAttackProtectionCaptchaResponseContent) []interface{} {
	if captcha == nil {
		return nil
	}

	m := make(map[string]interface{})

	// ActiveProviderID
	m["active_provider_id"] = captcha.GetActiveProviderID()

	// RecaptchaV2
	if recaptchaV2 := captcha.RecaptchaV2; recaptchaV2 != nil {
		// Preserve the user's configured secret (not returned by API)
		recaptchaV2Secret := ""
		if v, ok := data.GetOk("captcha.0.recaptcha_v2.0.secret"); ok {
			recaptchaV2Secret = v.(string)
		}

		m["recaptcha_v2"] = []interface{}{
			map[string]interface{}{
				"site_key": recaptchaV2.GetSiteKey(),
				"secret":   recaptchaV2Secret,
			},
		}
	}

	// RecaptchaEnterprise
	if recaptchaEnterprise := captcha.RecaptchaEnterprise; recaptchaEnterprise != nil {
		// Preserve the user's configured api_key (not returned by API)
		recaptchaEnterpriseAPIKey := ""
		if v, ok := data.GetOk("captcha.0.recaptcha_enterprise.0.api_key"); ok {
			recaptchaEnterpriseAPIKey = v.(string)
		}

		m["recaptcha_enterprise"] = []interface{}{
			map[string]interface{}{
				"site_key":   recaptchaEnterprise.GetSiteKey(),
				"api_key":    recaptchaEnterpriseAPIKey,
				"project_id": recaptchaEnterprise.GetProjectID(),
			},
		}
	}

	// Hcaptcha
	if hcaptcha := captcha.Hcaptcha; hcaptcha != nil {
		// Preserve the user's configured secret (not returned by API)
		hcaptchaSecret := ""
		if v, ok := data.GetOk("captcha.0.hcaptcha.0.secret"); ok {
			hcaptchaSecret = v.(string)
		}

		m["hcaptcha"] = []interface{}{
			map[string]interface{}{
				"site_key": hcaptcha.GetSiteKey(),
				"secret":   hcaptchaSecret,
			},
		}
	}

	// FriendlyCaptcha
	if friendlyCaptcha := captcha.FriendlyCaptcha; friendlyCaptcha != nil {
		// Preserve the user's configured secret (not returned by API)
		friendlyCaptchaSecret := ""
		if v, ok := data.GetOk("captcha.0.friendly_captcha.0.secret"); ok {
			friendlyCaptchaSecret = v.(string)
		}

		m["friendly_captcha"] = []interface{}{
			map[string]interface{}{
				"site_key": friendlyCaptcha.GetSiteKey(),
				"secret":   friendlyCaptchaSecret,
			},
		}
	}

	// Arkose
	if arkose := captcha.Arkose; arkose != nil {
		// Preserve the user's configured secret (not returned by API)
		arkoseSecret := ""
		if v, ok := data.GetOk("captcha.0.arkose.0.secret"); ok {
			arkoseSecret = v.(string)
		}

		arkoseMap := map[string]interface{}{
			"site_key": arkose.GetSiteKey(),
			"secret":   arkoseSecret,
			"fail_open": arkose.GetFailOpen(),
		}
		if clientSubdomain := arkose.GetClientSubdomain(); clientSubdomain != "" {
			arkoseMap["client_subdomain"] = clientSubdomain
		}
		if verifySubdomain := arkose.GetVerifySubdomain(); verifySubdomain != "" {
			arkoseMap["verify_subdomain"] = verifySubdomain
		}
		m["arkose"] = []interface{}{arkoseMap}
	}

	// AuthChallenge
	if authChallenge := captcha.AuthChallenge; authChallenge != nil {
		m["auth_challenge"] = []interface{}{
			map[string]interface{}{
				"fail_open": authChallenge.GetFailOpen(),
			},
		}
	}

	return []interface{}{m}
}
