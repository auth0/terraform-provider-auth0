package provider

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func newAttackProtection() *schema.Resource {
	return &schema.Resource{
		CreateContext: createAttackProtection,
		ReadContext:   readAttackProtection,
		UpdateContext: updateAttackProtection,
		DeleteContext: deleteAttackProtection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Auth0 can detect attacks and stop malicious attempts to access your " +
			"application such as blocking traffic from certain IPs and displaying CAPTCHAs.",
		Schema: map[string]*schema.Schema{
			"breached_password_detection": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Breached password detection protects your applications " +
					"from bad actors logging in with stolen credentials.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether breached password detection is active.",
						},
						"shields": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"user_notification",
									"admin_notification",
								}, false),
							},
							Description: "Action to take when a breached password is detected.",
						},
						"admin_notification_frequency": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"immediately",
									"daily",
									"weekly",
									"monthly",
								}, false),
							},
							Description: "When \"admin_notification\" is enabled, " +
								"determines how often email notifications are sent. " +
								"Possible values: `immediately`, `daily`, `weekly`, `monthly`.",
						},
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"standard", "enhanced",
							}, false),
							Description: "The subscription level for breached password detection methods. " +
								"Use \"enhanced\" to enable Credential Guard. Possible values: `standard`, `enhanced`.",
						},
					},
				},
			},
			"brute_force_protection": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Brute-force protection safeguards against a " +
					"single IP address attacking a single user account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether brute force attack protections are active.",
						},
						"shields": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"user_notification",
								}, false),
							},
							Description: "Action to take when a brute force protection threshold is violated. " +
								"Possible values: `block`, `user_notification`",
						},
						"allowlist": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of trusted IP addresses that will not " +
								"have attack protection enforced against them.",
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"count_per_identifier_and_ip", "count_per_identifier",
							}, false),
							Description: "Determines whether the IP address is used when counting failed attempts. " +
								"Possible values: `count_per_identifier_and_ip` or `count_per_identifier`.",
						},
						"max_attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
							Description:  "Maximum number of unsuccessful attempts. Only available on public tenants.",
						},
					},
				},
			},
			"suspicious_ip_throttling": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Suspicious IP throttling blocks traffic from any " +
					"IP address that rapidly attempts too many logins or signups.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether suspicious IP throttling attack protections are active.",
						},
						"shields": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"admin_notification",
								}, false),
							},
							Description: "Action to take when a suspicious IP throttling threshold is violated. " +
								"Possible values: `block`, `admin_notification`",
						},
						"allowlist": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of trusted IP addresses that will not have " +
								"attack protection enforced against them.",
						},
						"pre_login": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Description: "Configuration options that apply before every login attempt. " +
								"Only available on public tenants.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Total number of attempts allowed per day.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description: "Interval of time, given in milliseconds, " +
											"at which new attempts are granted.",
									},
								},
							},
						},
						"pre_user_registration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Description: "Configuration options that apply before every user registration attempt. " +
								"Only available on public tenants.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Total number of attempts allowed.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description: "Interval of time, given in milliseconds, " +
											"at which new attempts are granted.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createAttackProtection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updateAttackProtection(ctx, d, m)
}

func readAttackProtection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	breachedPasswords, err := api.AttackProtection.GetBreachedPasswordDetection()
	if err != nil {
		return diag.FromErr(err)
	}

	bruteForce, err := api.AttackProtection.GetBruteForceProtection()
	if err != nil {
		return diag.FromErr(err)
	}

	ipThrottling, err := api.AttackProtection.GetSuspiciousIPThrottling()
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("breached_password_detection", flattenBreachedPasswordProtection(breachedPasswords)),
		d.Set("brute_force_protection", flattenBruteForceProtection(bruteForce)),
		d.Set("suspicious_ip_throttling", flattenSuspiciousIPThrottling(ipThrottling)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateAttackProtection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if ipt := expandSuspiciousIPThrottling(d); ipt != nil {
		if err := api.AttackProtection.UpdateSuspiciousIPThrottling(ipt); err != nil {
			return diag.FromErr(err)
		}
	}

	if bfp := expandBruteForceProtection(d); bfp != nil {
		if err := api.AttackProtection.UpdateBruteForceProtection(bfp); err != nil {
			return diag.FromErr(err)
		}
	}

	if bpd := expandBreachedPasswordDetection(d); bpd != nil {
		if err := api.AttackProtection.UpdateBreachedPasswordDetection(bpd); err != nil {
			return diag.FromErr(err)
		}
	}

	return readAttackProtection(ctx, d, m)
}

func deleteAttackProtection(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	result := multierror.Append(
		api.AttackProtection.UpdateBreachedPasswordDetection(
			&management.BreachedPasswordDetection{
				Enabled: auth0.Bool(false),
			},
		),
		api.AttackProtection.UpdateBruteForceProtection(
			&management.BruteForceProtection{
				Enabled: auth0.Bool(false),
			},
		),
		api.AttackProtection.UpdateSuspiciousIPThrottling(
			&management.SuspiciousIPThrottling{
				Enabled: auth0.Bool(false),
			},
		),
	)
	if err := result.ErrorOrNil(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

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
		},
	}
}

func expandSuspiciousIPThrottling(d *schema.ResourceData) *management.SuspiciousIPThrottling {
	if !d.HasChange("suspicious_ip_throttling") {
		return nil
	}

	var ipt *management.SuspiciousIPThrottling
	var iptConfig = d.GetRawConfig().GetAttr("suspicious_ip_throttling")

	if iptConfig.IsNull() {
		return nil
	}

	iptConfig.ForEachElement(
		func(_ cty.Value, ipThrottling cty.Value) (stop bool) {
			ipt = &management.SuspiciousIPThrottling{
				Enabled:   value.Bool(ipThrottling.GetAttr("enabled")),
				Shields:   value.Strings(ipThrottling.GetAttr("shields")),
				AllowList: value.Strings(ipThrottling.GetAttr("allowlist")),
			}

			pl := ipThrottling.GetAttr("pre_login")
			if !pl.IsNull() {
				pl.ForEachElement(
					func(_ cty.Value, preLogin cty.Value) (stop bool) {
						ipt.Stage = &management.Stage{
							PreLogin: &management.PreLogin{
								MaxAttempts: value.Int(preLogin.GetAttr("max_attempts")),
								Rate:        value.Int(preLogin.GetAttr("rate")),
							},
						}

						return stop
					},
				)
			}

			pur := ipThrottling.GetAttr("pre_user_registration")
			if !pur.IsNull() {
				pur.ForEachElement(
					func(_ cty.Value, preUserReg cty.Value) (stop bool) {
						preUserRegistration := &management.PreUserRegistration{
							MaxAttempts: value.Int(preUserReg.GetAttr("max_attempts")),
							Rate:        value.Int(preUserReg.GetAttr("rate")),
						}

						if ipt.Stage != nil {
							ipt.Stage.PreUserRegistration = preUserRegistration
						} else {
							ipt.Stage = &management.Stage{
								PreUserRegistration: preUserRegistration,
							}
						}

						return stop
					},
				)
			}

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

	bfpConfig := d.GetRawConfig().GetAttr("brute_force_protection")

	if bfpConfig.IsNull() {
		return nil
	}

	bfpConfig.ForEachElement(
		func(_ cty.Value, bruteForce cty.Value) (stop bool) {
			bfp = &management.BruteForceProtection{
				Enabled:     value.Bool(bruteForce.GetAttr("enabled")),
				Mode:        value.String(bruteForce.GetAttr("mode")),
				MaxAttempts: value.Int(bruteForce.GetAttr("max_attempts")),
				Shields:     value.Strings(bruteForce.GetAttr("shields")),
				AllowList:   value.Strings(bruteForce.GetAttr("allowlist")),
			}

			return stop
		},
	)

	return bfp
}

func expandBreachedPasswordDetection(d *schema.ResourceData) *management.BreachedPasswordDetection {
	if !d.HasChange("breached_password_detection") {
		return nil
	}

	var bpd *management.BreachedPasswordDetection
	bpdConfig := d.GetRawConfig().GetAttr("breached_password_detection")

	if bpdConfig.IsNull() {
		return nil
	}

	bpdConfig.ForEachElement(
		func(_ cty.Value, breach cty.Value) (stop bool) {
			bpd = &management.BreachedPasswordDetection{
				Enabled:                    value.Bool(breach.GetAttr("enabled")),
				Method:                     value.String(breach.GetAttr("method")),
				Shields:                    value.Strings(breach.GetAttr("shields")),
				AdminNotificationFrequency: value.Strings(breach.GetAttr("admin_notification_frequency")),
			}

			return stop
		},
	)

	return bpd
}
