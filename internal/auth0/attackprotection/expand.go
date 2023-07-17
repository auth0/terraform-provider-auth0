package attackprotection

import (
	"github.com/auth0/go-auth0/management"
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

func expandBruteForceProtection(d *schema.ResourceData) *management.BruteForceProtection {
	if !d.HasChange("brute_force_protection") {
		return nil
	}

	var bfp *management.BruteForceProtection

	d.GetRawConfig().GetAttr("brute_force_protection").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
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

func expandBreachedPasswordDetection(d *schema.ResourceData) *management.BreachedPasswordDetection {
	if !d.HasChange("breached_password_detection") {
		return nil
	}

	var bpd *management.BreachedPasswordDetection

	d.GetRawConfig().GetAttr("breached_password_detection").ForEachElement(
		func(_ cty.Value, breach cty.Value) (stop bool) {
			bpd = &management.BreachedPasswordDetection{
				Enabled:                    value.Bool(breach.GetAttr("enabled")),
				Method:                     value.String(breach.GetAttr("method")),
				Shields:                    value.Strings(breach.GetAttr("shields")),
				AdminNotificationFrequency: value.Strings(breach.GetAttr("admin_notification_frequency")),
			}

			breach.GetAttr("pre_user_registration").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
				preUserRegistration := &management.BreachedPasswordDetectionPreUserRegistration{
					Shields: value.Strings(cfg.GetAttr("shields")),
				}

				if bpd.Stage != nil {
					bpd.Stage.PreUserRegistration = preUserRegistration
					return stop
				}

				bpd.Stage = &management.BreachedPasswordDetectionStage{
					PreUserRegistration: preUserRegistration,
				}

				return stop
			})

			return stop
		},
	)

	return bpd
}
