package attackprotection

import (
	"github.com/auth0/go-auth0/management"
	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandSuspiciousIPThrottling(data *schema.ResourceData) *management.SuspiciousIPThrottling {
	if !data.HasChange("suspicious_ip_throttling") {
		return nil
	}

	var ipt *management.SuspiciousIPThrottling

	data.GetRawConfig().GetAttr("suspicious_ip_throttling").ForEachElement(
		func(_ cty.Value, iptCfg cty.Value) (stop bool) {
			ipt = &management.SuspiciousIPThrottling{
				Enabled:   value.Bool(iptCfg.GetAttr("enabled")),
				Shields:   value.Strings(iptCfg.GetAttr("shields")),
				AllowList: value.Strings(iptCfg.GetAttr("allowlist")),
			}

			iptCfg.GetAttr("pre_login").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
				ipt.Stage = &management.Stage{
					PreLogin: &management.PreLogin{
						MaxAttempts: value.Int(cfg.GetAttr("max_attempts")),
						Rate:        value.Int(cfg.GetAttr("rate")),
					},
				}

				return stop
			})

			iptCfg.GetAttr("pre_user_registration").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
				preUserRegistration := &management.PreUserRegistration{
					MaxAttempts: value.Int(cfg.GetAttr("max_attempts")),
					Rate:        value.Int(cfg.GetAttr("rate")),
				}

				if ipt.Stage != nil {
					ipt.Stage.PreUserRegistration = preUserRegistration
					return stop
				}

				ipt.Stage = &management.Stage{
					PreUserRegistration: preUserRegistration,
				}

				return stop
			})

			return stop
		},
	)

	return ipt
}

func expandBruteForceProtection(data *schema.ResourceData) *management.BruteForceProtection {
	if !data.HasChange("brute_force_protection") {
		return nil
	}

	var bfp *management.BruteForceProtection

	data.GetRawConfig().GetAttr("brute_force_protection").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		bfp = &management.BruteForceProtection{
			Enabled:     value.Bool(cfg.GetAttr("enabled")),
			Mode:        value.String(cfg.GetAttr("mode")),
			MaxAttempts: value.Int(cfg.GetAttr("max_attempts")),
			Shields:     value.Strings(cfg.GetAttr("shields")),
			AllowList:   value.Strings(cfg.GetAttr("allowlist")),
		}

		return stop
	})

	return bfp
}

func expandBreachedPasswordDetection(data *schema.ResourceData) *management.BreachedPasswordDetection {
	if !data.HasChange("breached_password_detection") {
		return nil
	}

	var bpd *management.BreachedPasswordDetection

	data.GetRawConfig().GetAttr("breached_password_detection").ForEachElement(
		func(_ cty.Value, breach cty.Value) (stop bool) {
			bpd = &management.BreachedPasswordDetection{
				Enabled:                    value.Bool(breach.GetAttr("enabled")),
				Method:                     value.String(breach.GetAttr("method")),
				Shields:                    value.Strings(breach.GetAttr("shields")),
				AdminNotificationFrequency: value.Strings(breach.GetAttr("admin_notification_frequency")),
			}

			preUserRegistration := &management.BreachedPasswordDetectionPreUserRegistration{}
			preChangePassword := &management.BreachedPasswordDetectionPreChangePassword{}
			breach.GetAttr("pre_user_registration").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
				preUserRegistration.Shields = value.Strings(cfg.GetAttr("shields"))
				return stop
			})

			breach.GetAttr("pre_change_password").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
				preChangePassword.Shields = value.Strings(cfg.GetAttr("shields"))
				return stop
			})

			if bpd.Stage != nil {
				bpd.Stage.PreUserRegistration = preUserRegistration
				bpd.Stage.PreChangePassword = preChangePassword
				return stop
			}

			bpd.Stage = &management.BreachedPasswordDetectionStage{
				PreUserRegistration: preUserRegistration,
				PreChangePassword:   preChangePassword,
			}

			return stop
		},
	)

	return bpd
}

func expandBotDetection(data *schema.ResourceData) *managementv2.UpdateBotDetectionSettingsRequestContent {
	if !data.HasChange("bot_detection") {
		return nil
	}

	// Check if the bot_detection block is present in the configuration
	botDetectionAttr := data.GetRawConfig().GetAttr("bot_detection")
	if botDetectionAttr.IsNull() {
		// Block is not present in configuration, skip update
		return nil
	}

	var request *managementv2.UpdateBotDetectionSettingsRequestContent

	botDetectionAttr.ForEachElement(
		func(_ cty.Value, cfg cty.Value) (stop bool) {
			request = &managementv2.UpdateBotDetectionSettingsRequestContent{}

			// BotDetectionLevel.
			if levelStr := value.String(cfg.GetAttr("bot_detection_level")); levelStr != nil && *levelStr != "" {
				level := managementv2.BotDetectionLevelEnum(*levelStr)
				request.BotDetectionLevel = &level
			}

			// ChallengePasswordPolicy.
			if policyStr := value.String(cfg.GetAttr("challenge_password_policy")); policyStr != nil && *policyStr != "" {
				policy := managementv2.BotDetectionChallengePolicyPasswordFlowEnum(*policyStr)
				request.ChallengePasswordPolicy = &policy
			}

			// ChallengePasswordlessPolicy.
			if policyStr := value.String(cfg.GetAttr("challenge_passwordless_policy")); policyStr != nil && *policyStr != "" {
				policy := managementv2.BotDetectionChallengePolicyPasswordlessFlowEnum(*policyStr)
				request.ChallengePasswordlessPolicy = &policy
			}

			// ChallengePasswordResetPolicy.
			if policyStr := value.String(cfg.GetAttr("challenge_password_reset_policy")); policyStr != nil && *policyStr != "" {
				policy := managementv2.BotDetectionChallengePolicyPasswordResetFlowEnum(*policyStr)
				request.ChallengePasswordResetPolicy = &policy
			}

			// Allowlist.
			if allowlist := value.Strings(cfg.GetAttr("allowlist")); allowlist != nil {
				request.Allowlist = allowlist
			}

			// MonitoringModeEnabled.
			if monitoringMode := value.Bool(cfg.GetAttr("monitoring_mode_enabled")); monitoringMode != nil {
				request.MonitoringModeEnabled = monitoringMode
			}

			return stop
		},
	)

	// Only return request if it has at least one field set
	// This prevents sending empty requests when the block exists but has no content
	if request != nil {
		hasContent := request.BotDetectionLevel != nil ||
			request.ChallengePasswordPolicy != nil ||
			request.ChallengePasswordlessPolicy != nil ||
			request.ChallengePasswordResetPolicy != nil ||
			request.Allowlist != nil ||
			request.MonitoringModeEnabled != nil

		if !hasContent {
			return nil
		}
	}

	return request
}

func expandCaptcha(data *schema.ResourceData) *managementv2.UpdateAttackProtectionCaptchaRequestContent {
	if !data.HasChange("captcha") {
		return nil
	}

	var request *managementv2.UpdateAttackProtectionCaptchaRequestContent

	data.GetRawConfig().GetAttr("captcha").ForEachElement(
		func(_ cty.Value, cfg cty.Value) (stop bool) {
			request = &managementv2.UpdateAttackProtectionCaptchaRequestContent{}

			// ActiveProviderID.
			if providerID := value.String(cfg.GetAttr("active_provider_id")); providerID != nil {
				if *providerID != "" {
					pid := managementv2.AttackProtectionCaptchaProviderID(*providerID)
					request.ActiveProviderID = &pid
				}
			}

			// RecaptchaV2.
			cfg.GetAttr("recaptcha_v2").ForEachElement(func(_ cty.Value, v2cfg cty.Value) (stop bool) {
				siteKey := value.String(v2cfg.GetAttr("site_key"))
				secret := value.String(v2cfg.GetAttr("secret"))
				if siteKey != nil && secret != nil {
					request.RecaptchaV2 = &managementv2.AttackProtectionUpdateCaptchaRecaptchaV2{
						SiteKey: *siteKey,
						Secret:  *secret,
					}
				}
				return stop
			})

			// RecaptchaEnterprise.
			cfg.GetAttr("recaptcha_enterprise").ForEachElement(func(_ cty.Value, entcfg cty.Value) (stop bool) {
				siteKey := value.String(entcfg.GetAttr("site_key"))
				apiKey := value.String(entcfg.GetAttr("api_key"))
				projectID := value.String(entcfg.GetAttr("project_id"))
				if siteKey != nil && apiKey != nil && projectID != nil {
					request.RecaptchaEnterprise = &managementv2.AttackProtectionUpdateCaptchaRecaptchaEnterprise{
						SiteKey:   *siteKey,
						APIKey:    *apiKey,
						ProjectID: *projectID,
					}
				}
				return stop
			})

			// Hcaptcha.
			cfg.GetAttr("hcaptcha").ForEachElement(func(_ cty.Value, hcfg cty.Value) (stop bool) {
				siteKey := value.String(hcfg.GetAttr("site_key"))
				secret := value.String(hcfg.GetAttr("secret"))
				if siteKey != nil && secret != nil {
					request.Hcaptcha = &managementv2.AttackProtectionUpdateCaptchaHcaptcha{
						SiteKey: *siteKey,
						Secret:  *secret,
					}
				}
				return stop
			})

			// FriendlyCaptcha.
			cfg.GetAttr("friendly_captcha").ForEachElement(func(_ cty.Value, fcfg cty.Value) (stop bool) {
				siteKey := value.String(fcfg.GetAttr("site_key"))
				secret := value.String(fcfg.GetAttr("secret"))
				if siteKey != nil && secret != nil {
					request.FriendlyCaptcha = &managementv2.AttackProtectionUpdateCaptchaFriendlyCaptcha{
						SiteKey: *siteKey,
						Secret:  *secret,
					}
				}
				return stop
			})

			// Arkose.
			cfg.GetAttr("arkose").ForEachElement(func(_ cty.Value, acfg cty.Value) (stop bool) {
				siteKey := value.String(acfg.GetAttr("site_key"))
				secret := value.String(acfg.GetAttr("secret"))
				if siteKey != nil && secret != nil {
					arkose := &managementv2.AttackProtectionUpdateCaptchaArkose{
						SiteKey: *siteKey,
						Secret:  *secret,
					}
					// Optional fields.
					if clientSubdomain := value.String(acfg.GetAttr("client_subdomain")); clientSubdomain != nil {
						arkose.ClientSubdomain = clientSubdomain
					}
					if verifySubdomain := value.String(acfg.GetAttr("verify_subdomain")); verifySubdomain != nil {
						arkose.VerifySubdomain = verifySubdomain
					}
					if failOpen := value.Bool(acfg.GetAttr("fail_open")); failOpen != nil {
						arkose.FailOpen = failOpen
					}
					request.Arkose = arkose
				}
				return stop
			})

			// AuthChallenge.
			cfg.GetAttr("auth_challenge").ForEachElement(func(_ cty.Value, ac cty.Value) (stop bool) {
				failOpen := value.Bool(ac.GetAttr("fail_open"))
				if failOpen != nil {
					request.AuthChallenge = &managementv2.AttackProtectionCaptchaAuthChallengeRequest{
						FailOpen: *failOpen,
					}
				}
				return stop
			})

			return stop
		},
	)

	return request
}
