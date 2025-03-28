package attackprotection

import (
	"github.com/auth0/go-auth0/management"
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
