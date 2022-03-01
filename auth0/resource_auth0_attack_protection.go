package auth0

import (
	"fmt"
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
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
							MinItems:    0,
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
							MinItems:    0,
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
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
							MinItems:    0,
							Description: "Action to take when a brute force protection threshold is violated.",
						},
						"allowlist": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Optional:    true,
							MinItems:    0,
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				MinItems: 1,
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
							MinItems: 0,
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
							MinItems:    0,
							Description: "List of trusted IP addresses that will not have attack protection enforced against them.",
						},
						"pre_login": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							MinItems:    1,
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
							MinItems:    1,
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

func flattenSuspiciousIPThrottling(ipt *management.SuspiciousIPThrottling) []interface{} {
	m := make(map[string]interface{})
	if ipt != nil {
		m["enabled"] = ipt.Enabled
		m["allowlist"] = ipt.AllowList
		m["shields"] = ipt.Shields
		m["pre_login"] = []interface{}{
			map[string]int{
				"max_attempts": ipt.Stage.PreLogin.GetMaxAttempts(),
				"rate":         ipt.Stage.PreLogin.GetRate(),
			},
		}
		m["pre_user_registration"] = []interface{}{
			map[string]int{
				"max_attempts": ipt.Stage.PreUserRegistration.GetMaxAttempts(),
				"rate":         ipt.Stage.PreUserRegistration.GetRate(),
			},
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

func updateAttackProtection(d *schema.ResourceData, m interface{}) error {
	ipt := expandSuspiciousIPThrottling(d)
	api := m.(*management.Management)
	err := api.AttackProtection.UpdateSuspiciousIPThrottling(ipt)
	if err != nil {
		return err
	}

	return readAttackProtection(d, m)
}

func expandSuspiciousIPThrottling(d *schema.ResourceData) *management.SuspiciousIPThrottling {
	ipt := &management.SuspiciousIPThrottling{}

	List(d, "suspicious_ip_throttling", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		var shields []string
		for _, s := range d.Get("shields").([]interface{}) {
			shields = append(shields, fmt.Sprintf("%s", s))
		}

		var allowlist []string
		for _, a := range d.Get("allowlist").([]interface{}) {
			allowlist = append(allowlist, fmt.Sprintf("%s", a))
		}

		ipt = &management.SuspiciousIPThrottling{
			Enabled:   Bool(d, "enabled"),
			Shields:   &shields,
			AllowList: &allowlist,
			Stage: &management.Stage{
				PreUserRegistration: &management.PreUserRegistration{},
				PreLogin:            &management.PreLogin{},
			},
		}

		List(d, "pre_login").Elem(func(d ResourceData) {
			ipt.Stage.PreLogin.MaxAttempts = Int(d, "max_attempts")
			ipt.Stage.PreLogin.Rate = Int(d, "rate")
		})

		List(d, "pre_user_registration").Elem(func(d ResourceData) {
			ipt.Stage.PreUserRegistration.MaxAttempts = Int(d, "max_attempts")
			ipt.Stage.PreUserRegistration.Rate = Int(d, "rate")
		})
	})

	return ipt
}

func createAttackProtection(d *schema.ResourceData, m interface{}) error {
	d.SetId(resource.UniqueId())
	return updateAttackProtection(d, m)
}

func deleteAttackProtection(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}
