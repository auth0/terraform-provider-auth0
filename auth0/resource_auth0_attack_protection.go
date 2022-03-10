package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func newAttackProtection() *schema.Resource {
	return &schema.Resource{
		Create: createAttackProtection,
		Read:   readAttackProtection,
		Update: updateAttackProtection,
		Delete: deleteAttackProtection,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"breached_password_detection": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Computed:    true,
				Description: "Breached password detection protects your applications from bad actors logging in with stolen credentials.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not breached password detection is active.",
						},
						"shields": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"user_notification",
									"admin_notification",
								}, false),
							},
							Optional:    true,
							Description: "Action to take when a breached password is detected.",
						},
						"admin_notification_frequency": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"immediately",
									"daily",
									"weekly",
									"monthly",
								}, false),
							},
							Optional:    true,
							Description: "When \"admin_notification\" is enabled, determines how often email notifications are sent.",
						},
						"method": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"standard", "enhanced",
							}, false),
							Description: "The subscription level for breached password detection methods. Use \"enhanced\" to enable Credential Guard.",
						},
					},
				},
			},
			"brute_force_protection": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Computed:    true,
				Description: "Brute-force protection safeguards against a single IP address attacking a single user account.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not brute force attack protections are active.",
						},
						"shields": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"user_notification",
								}, false),
							},
							Optional:    true,
							Description: "Action to take when a brute force protection threshold is violated.",
						},
						"allowlist": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "List of trusted IP addresses that will not have attack protection enforced against them.",
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"count_per_identifier_and_ip", "count_per_identifier",
							}, false),
							Description: "Account Lockout: Determines whether or not IP address is used when counting failed attempts.",
						},
						"max_attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(0),
							Description:  "Maximum number of unsuccessful attempts.",
						},
					},
				},
			},
			"suspicious_ip_throttling": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Suspicious IP throttling blocks traffic from any IP address that rapidly attempts too many logins or signups.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether or not suspicious IP throttling attack protections are active.",
						},
						"shields": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								ValidateFunc: validation.StringInSlice([]string{
									"block",
									"admin_notification",
								}, false),
							},
							Description: "Action to take when a suspicious IP throttling threshold is violated.",
						},
						"allowlist": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							Description: "List of trusted IP addresses that will not have attack protection enforced against them.",
						},
						"pre_login": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Configuration options that apply before every login attempt.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Total number of attempts allowed per day.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Interval of time, given in milliseconds, at which new attempts are granted.",
									},
								},
							},
						},
						"pre_user_registration": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Configuration options that apply before every user registration attempt.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"max_attempts": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Total number of attempts allowed.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description:  "Interval of time, given in milliseconds, at which new attempts are granted.",
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

func createAttackProtection(d *schema.ResourceData, m interface{}) error {
	d.SetId(resource.UniqueId())
	return updateAttackProtection(d, m)
}

func readAttackProtection(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	ipThrottling, err := api.AttackProtection.GetSuspiciousIPThrottling()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("suspicious_ip_throttling", flattenSuspiciousIPThrottling(ipThrottling)); err != nil {
		return err
	}

	bruteForce, err := api.AttackProtection.GetBruteForceProtection()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("brute_force_protection", flattenBruteForceProtection(bruteForce)); err != nil {
		return err
	}

	breachedPasswords, err := api.AttackProtection.GetBreachedPasswordDetection()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if err = d.Set("breached_password_detection", flattenBreachedPasswordProtection(breachedPasswords)); err != nil {
		return err
	}

	return nil
}

func updateAttackProtection(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)

	if ipt := expandSuspiciousIPThrottling(d); ipt != nil {
		if err := api.AttackProtection.UpdateSuspiciousIPThrottling(ipt); err != nil {
			return err
		}
	}

	if bfp := expandBruteForceProtection(d); bfp != nil {
		if err := api.AttackProtection.UpdateBruteForceProtection(bfp); err != nil {
			return err
		}
	}

	if bpd := expandBreachedPasswordDetection(d); bpd != nil {
		if err := api.AttackProtection.UpdateBreachedPasswordDetection(bpd); err != nil {
			return err
		}
	}

	return readAttackProtection(d, m)
}

func deleteAttackProtection(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

func flattenSuspiciousIPThrottling(ipt *management.SuspiciousIPThrottling) []interface{} {
	m := make(map[string]interface{})
	if ipt != nil {
		m["enabled"] = ipt.Enabled
		m["allowlist"] = ipt.AllowList
		m["shields"] = ipt.Shields

		if ipt.Stage != nil {
			if ipt.Stage.PreLogin != nil {
				m["pre_login"] = []interface{}{
					map[string]int{
						"max_attempts": ipt.Stage.PreLogin.GetMaxAttempts(),
						"rate":         ipt.Stage.PreLogin.GetRate(),
					},
				}
			}
			if ipt.Stage.PreUserRegistration != nil {
				m["pre_user_registration"] = []interface{}{
					map[string]int{
						"max_attempts": ipt.Stage.PreUserRegistration.GetMaxAttempts(),
						"rate":         ipt.Stage.PreUserRegistration.GetRate(),
					},
				}
			}
		}

	}
	return []interface{}{m}
}

func flattenBruteForceProtection(bfp *management.BruteForceProtection) []interface{} {
	m := make(map[string]interface{})
	if bfp != nil {
		m["enabled"] = bfp.Enabled
		m["max_attempts"] = bfp.MaxAttempts
		m["mode"] = bfp.Mode
		m["allowlist"] = bfp.AllowList
		m["shields"] = bfp.Shields
	}
	return []interface{}{m}
}

func flattenBreachedPasswordProtection(bpd *management.BreachedPasswordDetection) []interface{} {
	m := make(map[string]interface{})
	if bpd != nil {
		m["enabled"] = bpd.Enabled
		m["admin_notification_frequency"] = bpd.AdminNotificationFrequency
		m["method"] = bpd.Method
		m["shields"] = bpd.Shields
	}
	return []interface{}{m}
}

func expandSuspiciousIPThrottling(d *schema.ResourceData) *management.SuspiciousIPThrottling {
	var ipt management.SuspiciousIPThrottling

	List(d, "suspicious_ip_throttling", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		ipt.Enabled = Bool(d, "enabled")

		stateShields := d.Get("shields").([]interface{})
		shields := make([]string, len(stateShields))
		for index, value := range stateShields {
			shields[index] = value.(string)
		}
		ipt.Shields = &shields

		stateAllowList := d.Get("allowlist").([]interface{})
		allowlist := make([]string, len(stateAllowList))
		for index, value := range stateAllowList {
			allowlist[index] = value.(string)
		}
		ipt.AllowList = &allowlist

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

	return &ipt
}

func expandBruteForceProtection(d *schema.ResourceData) *management.BruteForceProtection {
	var bfp management.BruteForceProtection

	List(d, "brute_force_protection", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		bfp.Enabled = Bool(d, "enabled")

		stateShields := d.Get("shields").([]interface{})
		shields := make([]string, len(stateShields))
		for index, value := range stateShields {
			shields[index] = value.(string)
		}
		bfp.Shields = &shields

		stateAllowList := d.Get("allowlist").([]interface{})
		allowlist := make([]string, len(stateAllowList))
		for index, value := range stateAllowList {
			allowlist[index] = value.(string)
		}
		bfp.AllowList = &allowlist

		bfp.Mode = String(d, "mode")
		bfp.MaxAttempts = Int(d, "max_attempts")
	})

	return &bfp
}

func expandBreachedPasswordDetection(d *schema.ResourceData) *management.BreachedPasswordDetection {
	var bpd management.BreachedPasswordDetection

	List(d, "breached_password_detection", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		bpd.Enabled = Bool(d, "enabled")

		stateShields := d.Get("shields").([]interface{})
		shields := make([]string, len(stateShields))
		for index, value := range stateShields {
			shields[index] = value.(string)
		}
		bpd.Shields = &shields

		stateAdminNotificationFrequency := d.Get("admin_notification_frequency").([]interface{})
		adminNotificationFrequency := make([]string, len(stateAdminNotificationFrequency))
		for index, value := range stateAdminNotificationFrequency {
			adminNotificationFrequency[index] = value.(string)
		}
		bpd.AdminNotificationFrequency = &adminNotificationFrequency

		bpd.Method = String(d, "method")
	})

	return &bpd
}
