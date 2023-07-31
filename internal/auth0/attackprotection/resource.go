package attackprotection

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewResource will return a new auth0_attack_protection resource.
func NewResource() *schema.Resource {
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
							Required:    true,
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
							Description: "Action to take when a breached password is detected. " +
								"Options include: `block` (block compromised user accounts), " +
								"`user_notification` (send an email to user when we detect that they are using compromised credentials) " +
								"and `admin_notification` (send an email with a summary of the number of accounts logging in with compromised credentials).",
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
							RequiredWith: []string{"breached_password_detection.0.shields"},
							Description: "When `admin_notification` is enabled within the `shields` property, " +
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
						"pre_user_registration": {
							Type:     schema.TypeList,
							Optional: true,
							Computed: true,
							MaxItems: 1,
							Description: "Configuration options that apply before every user registration attempt. " +
								"Only available on public tenants.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
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
										Description: "Action to take when a breached password is detected during " +
											"a signup. Possible values: `block` (block compromised credentials for new accounts), " +
											"`admin_notification` (send an email notification with a summary of compromised credentials in new accounts).",
									},
								},
							},
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
							Required:    true,
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
								"Possible values: `block` (block login attempts for a flagged user account), " +
								"`user_notification` (send an email to user when their account has been blocked).",
						},
						"allowlist": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of trusted IP addresses that will not have attack protection enforced " +
								"against them. This field allows you to specify multiple IP addresses, or ranges. " +
								"You can use IPv4 or IPv6 addresses and CIDR notation.",
						},
						"mode": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							ValidateFunc: validation.StringInSlice([]string{
								"count_per_identifier_and_ip", "count_per_identifier",
							}, false),
							Description: "Determines whether the IP address is used when counting failed attempts. " +
								"Possible values: `count_per_identifier_and_ip` (lockout an account from a given IP Address) " +
								"or `count_per_identifier` (lockout an account regardless of IP Address).",
						},
						"max_attempts": {
							Type:         schema.TypeInt,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.IntAtLeast(0),
							Description: "Maximum number of consecutive failed login attempts from a single user " +
								"before blocking is triggered. Only available on public tenants.",
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
							Required:    true,
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
								"Possible values: `block` (throttle traffic from an IP address when there is a high number of login attempts targeting too many different accounts), " +
								"`admin_notification` (send an email notification when traffic is throttled on one or more IP addresses due to high-velocity traffic).",
						},
						"allowlist": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "List of trusted IP addresses that will not have attack protection enforced " +
								"against them. This field allows you to specify multiple IP addresses, or ranges. " +
								"You can use IPv4 or IPv6 addresses and CIDR notation.",
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
										Description:  "The maximum number of failed login attempts allowed from a single IP address.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description: "Interval of time, given in milliseconds at which new login tokens " +
											"will become available after they have been used by an IP address. " +
											"Each login attempt will be added on the defined throttling rate.",
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
										Description:  "The maximum number of sign up attempts allowed from a single IP address.",
									},
									"rate": {
										Type:         schema.TypeInt,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IntAtLeast(0),
										Description: "Interval of time, given in milliseconds " +
											"at which new sign up tokens will become available after they have been used " +
											"by an IP address. Each sign up attempt will be added on the defined throttling rate.",
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

func createAttackProtection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateAttackProtection(ctx, data, meta)
}

func readAttackProtection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	breachedPasswords, err := api.AttackProtection.GetBreachedPasswordDetection(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	bruteForce, err := api.AttackProtection.GetBruteForceProtection(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	ipThrottling, err := api.AttackProtection.GetSuspiciousIPThrottling(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("breached_password_detection", flattenBreachedPasswordProtection(breachedPasswords)),
		data.Set("brute_force_protection", flattenBruteForceProtection(bruteForce)),
		data.Set("suspicious_ip_throttling", flattenSuspiciousIPThrottling(ipThrottling)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateAttackProtection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var result *multierror.Error
	if ipt := expandSuspiciousIPThrottling(data); ipt != nil {
		result = multierror.Append(result, api.AttackProtection.UpdateSuspiciousIPThrottling(ctx, ipt))
	}

	if bfp := expandBruteForceProtection(data); bfp != nil {
		result = multierror.Append(result, api.AttackProtection.UpdateBruteForceProtection(ctx, bfp))
	}

	if bpd := expandBreachedPasswordDetection(data); bpd != nil {
		result = multierror.Append(result, api.AttackProtection.UpdateBreachedPasswordDetection(ctx, bpd))
	}

	if result.ErrorOrNil() != nil {
		return diag.FromErr(result.ErrorOrNil())
	}

	return readAttackProtection(ctx, data, meta)
}

func deleteAttackProtection(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	enabled := false

	result := multierror.Append(
		api.AttackProtection.UpdateBreachedPasswordDetection(
			ctx,
			&management.BreachedPasswordDetection{
				Enabled: &enabled,
			},
		),
		api.AttackProtection.UpdateBruteForceProtection(
			ctx,
			&management.BruteForceProtection{
				Enabled: &enabled,
			},
		),
		api.AttackProtection.UpdateSuspiciousIPThrottling(
			ctx,
			&management.SuspiciousIPThrottling{
				Enabled: &enabled,
			},
		),
	)

	return diag.FromErr(result.ErrorOrNil())
}
