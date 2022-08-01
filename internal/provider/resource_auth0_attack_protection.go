package provider

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
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
			"application such as blocking traffic from certain IPs and displaying CAPTCHA.",
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
							Description: "Determines whether IP address is used when counting failed attempts. " +
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
	var allowList []interface{}
	for _, ip := range ipt.GetAllowList() {
		allowList = append(allowList, ip)
	}

	var shields []interface{}
	for _, shield := range ipt.GetShields() {
		shields = append(shields, shield)
	}

	return []interface{}{
		map[string]interface{}{
			"enabled":   ipt.GetEnabled(),
			"allowlist": allowList,
			"shields":   shields,
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
	var allowList []interface{}
	for _, ip := range bfp.GetAllowList() {
		allowList = append(allowList, ip)
	}

	var shields []interface{}
	for _, shield := range bfp.GetShields() {
		shields = append(shields, shield)
	}

	return []interface{}{
		map[string]interface{}{
			"enabled":      bfp.GetEnabled(),
			"mode":         bfp.GetMode(),
			"max_attempts": bfp.GetMaxAttempts(),
			"shields":      shields,
			"allowlist":    allowList,
		},
	}
}

func flattenBreachedPasswordProtection(bpd *management.BreachedPasswordDetection) []interface{} {
	var adminNotificationFrequency []interface{}
	for _, frequency := range bpd.GetAdminNotificationFrequency() {
		adminNotificationFrequency = append(adminNotificationFrequency, frequency)
	}

	var shields []interface{}
	for _, shield := range bpd.GetShields() {
		shields = append(shields, shield)
	}

	return []interface{}{
		map[string]interface{}{
			"enabled":                      bpd.GetEnabled(),
			"method":                       bpd.GetMethod(),
			"admin_notification_frequency": adminNotificationFrequency,
			"shields":                      shields,
		},
	}
}

func expandSuspiciousIPThrottling(d *schema.ResourceData) *management.SuspiciousIPThrottling {
	var ipt *management.SuspiciousIPThrottling

	List(d, "suspicious_ip_throttling", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		ipt = &management.SuspiciousIPThrottling{
			Enabled: Bool(d, "enabled"),
		}

		var shields []string
		for _, shield := range Set(d, "shields", IsNewResource(), HasChange()).List() {
			shields = append(shields, shield.(string))
		}
		if len(shields) > 0 {
			ipt.Shields = &shields
		}

		var allowList []string
		for _, ip := range Set(d, "allowlist", IsNewResource(), HasChange()).List() {
			allowList = append(allowList, ip.(string))
		}
		if len(allowList) > 0 {
			ipt.AllowList = &allowList
		}

		List(d, "pre_login", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
			ipt.Stage = &management.Stage{
				PreLogin: &management.PreLogin{
					MaxAttempts: Int(d, "max_attempts"),
					Rate:        Int(d, "rate"),
				},
			}
		})

		List(d, "pre_user_registration", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
			preUserRegistration := &management.PreUserRegistration{
				MaxAttempts: Int(d, "max_attempts"),
				Rate:        Int(d, "rate"),
			}

			if ipt.Stage != nil {
				ipt.Stage.PreUserRegistration = preUserRegistration
			} else {
				ipt.Stage = &management.Stage{
					PreUserRegistration: preUserRegistration,
				}
			}
		})
	})

	return ipt
}

func expandBruteForceProtection(d *schema.ResourceData) *management.BruteForceProtection {
	var bfp *management.BruteForceProtection

	List(d, "brute_force_protection", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		bfp = &management.BruteForceProtection{
			Enabled:     Bool(d, "enabled"),
			Mode:        String(d, "mode"),
			MaxAttempts: Int(d, "max_attempts"),
		}

		var shields []string
		for _, shield := range Set(d, "shields", IsNewResource(), HasChange()).List() {
			shields = append(shields, shield.(string))
		}
		if len(shields) > 0 {
			bfp.Shields = &shields
		}

		var allowList []string
		for _, ip := range Set(d, "allowlist", IsNewResource(), HasChange()).List() {
			allowList = append(allowList, ip.(string))
		}
		if len(allowList) > 0 {
			bfp.AllowList = &allowList
		}
	})

	return bfp
}

func expandBreachedPasswordDetection(d *schema.ResourceData) *management.BreachedPasswordDetection {
	var bpd *management.BreachedPasswordDetection

	List(d, "breached_password_detection", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		bpd = &management.BreachedPasswordDetection{
			Enabled: Bool(d, "enabled"),
			Method:  String(d, "method"),
		}

		var shields []string
		for _, shield := range Set(d, "shields", IsNewResource(), HasChange()).List() {
			shields = append(shields, shield.(string))
		}
		if len(shields) > 0 {
			bpd.Shields = &shields
		}

		var adminNotificationFrequency []string
		for _, frequency := range Set(d, "admin_notification_frequency", IsNewResource(), HasChange()).List() {
			adminNotificationFrequency = append(adminNotificationFrequency, frequency.(string))
		}
		if len(adminNotificationFrequency) > 0 {
			bpd.AdminNotificationFrequency = &adminNotificationFrequency
		}
	})

	return bpd
}
