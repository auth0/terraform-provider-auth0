package connection

import (
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/connection"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewDatabaseResource will return a new auth0_connection_database resource.
func NewDatabaseResource() *schema.Resource {
	baseResource := connection.NewBaseConnectionResource(
		map[string]*schema.Schema{
			"validation": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Validation of the minimum and maximum values allowed for a user to have as username.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"username": {
							Optional:    true,
							Type:        schema.TypeList,
							MaxItems:    1,
							Description: "Specifies the `min` and `max` values of username length.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"min": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"max": {
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
								},
							},
						},
					},
				},
			},
			"password_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none", "low", "fair", "good", "excellent",
				}, false),
				Description: "Indicates level of password strength to enforce during authentication. " +
					"A strong password policy will make it difficult, if not improbable, for someone " +
					"to guess a password through either manual or automated means. " +
					"Options include `none`, `low`, `fair`, `good`, `excellent`.",
			},
			"password_history": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Description: "Configuration settings for the password history " +
					"that is maintained for each user to prevent the reuse of passwords.",
			},
			"password_dictionary": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Optional: true,
							Description: "Indicates whether the password dictionary check " +
								"is enabled for this connection.",
						},
						"dictionary": {
							Type:     schema.TypeSet,
							Elem:     &schema.Schema{Type: schema.TypeString},
							Optional: true,
							Description: "Customized contents of the password dictionary. By default, " +
								"the password dictionary contains a list of the " +
								"[10,000 most common passwords](https://github.com/danielmiessler/SecLists/blob/master/Passwords/Common-Credentials/10k-most-common.txt); " +
								"your customized content is used in addition to the default password dictionary. " +
								"Matching is not case-sensitive.",
						},
					},
				},
				Description: "Configuration settings for the password dictionary check, " +
					"which does not allow passwords that are part of the password dictionary.",
			},
			"password_no_personal_info": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable": {
							Type:     schema.TypeBool,
							Optional: true,
						},
					},
				},
				Description: "Configuration settings for the password personal info check, " +
					"which does not allow passwords that contain any part " +
					"of the user's personal data, including user's `name`, `username`, `nickname`, " +
					"`user_metadata.name`, `user_metadata.first`, `user_metadata.last`, user's `email`, " +
					"or first part of the user's `email`.",
			},

			"password_complexity_options": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							ValidateFunc: validation.IntAtLeast(1),
							Description:  "Minimum number of characters allowed in passwords.",
						},
					},
				},
				Description: "Configuration settings for password complexity.",
			},
			"enable_script_context": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Set to `true` to inject context into custom DB scripts " +
					"(warning: cannot be disabled once enabled).",
			},
			"enabled_database_customization": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Set to `true` to use a legacy user store.",
			},
			"import_mode": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Indicates whether you have a legacy user store and want to gradually migrate " +
					"those users to the Auth0 user store.",
			},
			"disable_self_service_change_password": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether to remove the forgot password link within the New Universal Login.",
			},
			"requires_username": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Indicates whether the user is required to provide a username " +
					"in addition to an email address.",
			},
			"custom_scripts": {
				Type:        schema.TypeMap,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "A map of scripts used to integrate with a custom database.",
			},
			"configuration": {
				Type:      schema.TypeMap,
				Elem:      schema.TypeString,
				Sensitive: true,
				Optional:  true,
				Description: "A case-sensitive map of key value pairs used as configuration variables " +
					"for the `custom_script`.",
			},
			"brute_force_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Indicates whether to enable brute force protection, which will limit " +
					"the number of signups and failed logins from a suspicious IP address.",
			},
			"disable_signup": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether to allow user sign-ups to your application.",
			},
			"mfa": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Computed:    true,
				Optional:    true,
				Description: "Configuration options for multifactor authentication.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"active": {
							Type:     schema.TypeBool,
							Optional: true,
							Description: "Indicates whether multifactor authentication " +
								"is enabled for this connection.",
						},
						"return_enroll_settings": {
							Type:     schema.TypeBool,
							Optional: true,
							Description: "Indicates whether multifactor authentication " +
								"enrollment settings will be returned.",
						},
					},
				},
			},
			"set_user_root_attributes": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"on_each_login", "on_first_login"}, false),
				Description: "Determines whether to sync user profile attributes (`name`, `given_name`, " +
					"`family_name`, `nickname`, `picture`) at each login or only on the first login. Options " +
					"include: `on_each_login`, `on_first_login`. Default value: `on_each_login`.",
			},
		},
		expandConnectionAuth0,
		flattenConnectionAuth0,
	)

	return baseResource
}

func flattenConnectionAuth0(
	d *schema.ResourceData,
	options *management.ConnectionOptions,
) (map[string]interface{}, diag.Diagnostics) {
	dbSecretConfig, ok := d.GetOk("options.0.configuration")
	if !ok {
		dbSecretConfig = make(map[string]interface{})
	}

	m := map[string]interface{}{
		"strategy":                             "auth0",
		"password_policy":                      options.GetPasswordPolicy(),
		"enable_script_context":                options.GetEnableScriptContext(),
		"enabled_database_customization":       options.GetEnabledDatabaseCustomization(),
		"brute_force_protection":               options.GetBruteForceProtection(),
		"import_mode":                          options.GetImportMode(),
		"disable_signup":                       options.GetDisableSignup(),
		"disable_self_service_change_password": options.GetDisableSelfServiceChangePassword(),
		"requires_username":                    options.GetRequiresUsername(),
		"custom_scripts":                       options.GetCustomScripts(),
		"configuration":                        dbSecretConfig, // Values do not get read back.
		"non_persistent_attrs":                 options.GetNonPersistentAttrs(),
		"set_user_root_attributes":             options.GetSetUserAttributes(),
	}

	if options.PasswordComplexityOptions != nil {
		m["password_complexity_options"] = []interface{}{options.PasswordComplexityOptions}
	}
	if options.PasswordDictionary != nil {
		m["password_dictionary"] = []interface{}{options.PasswordDictionary}
	}
	if options.PasswordNoPersonalInfo != nil {
		m["password_no_personal_info"] = []interface{}{options.PasswordNoPersonalInfo}
	}
	if options.PasswordHistory != nil {
		m["password_history"] = []interface{}{options.PasswordHistory}
	}
	if options.MFA != nil {
		m["mfa"] = []interface{}{options.MFA}
	}
	if options.Validation != nil {
		m["validation"] = []interface{}{
			map[string]interface{}{
				"username": []interface{}{
					options.Validation["username"],
				},
			},
		}
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func expandConnectionAuth0(
	conn *management.Connection,
	d *schema.ResourceData,
	api *management.Management,
) (*management.Connection, diag.Diagnostics) {
	config := d.GetRawConfig()

	options := &management.ConnectionOptions{
		PasswordPolicy:                   value.String(config.GetAttr("password_policy")),
		NonPersistentAttrs:               value.Strings(config.GetAttr("non_persistent_attrs")),
		SetUserAttributes:                value.String(config.GetAttr("set_user_root_attributes")),
		EnableScriptContext:              value.Bool(config.GetAttr("enable_script_context")),
		EnabledDatabaseCustomization:     value.Bool(config.GetAttr("enabled_database_customization")),
		BruteForceProtection:             value.Bool(config.GetAttr("brute_force_protection")),
		ImportMode:                       value.Bool(config.GetAttr("import_mode")),
		DisableSignup:                    value.Bool(config.GetAttr("disable_signup")),
		DisableSelfServiceChangePassword: value.Bool(config.GetAttr("disable_self_service_change_password")),
		RequiresUsername:                 value.Bool(config.GetAttr("requires_username")),
		CustomScripts:                    value.MapOfStrings(config.GetAttr("custom_scripts")),
		Configuration:                    value.MapOfStrings(config.GetAttr("configuration")),
	}

	config.GetAttr("validation").ForEachElement(
		func(_ cty.Value, validation cty.Value) (stop bool) {
			validationOption := make(map[string]interface{})

			validation.GetAttr("username").ForEachElement(
				func(_ cty.Value, username cty.Value) (stop bool) {
					usernameValidation := make(map[string]*int)

					if min := value.Int(username.GetAttr("min")); min != nil {
						usernameValidation["min"] = min
					}
					if max := value.Int(username.GetAttr("max")); max != nil {
						usernameValidation["max"] = max
					}

					if len(usernameValidation) > 0 {
						validationOption["username"] = usernameValidation
					}

					return stop
				},
			)

			if len(validationOption) > 0 {
				options.Validation = validationOption
			}

			return stop
		},
	)

	config.GetAttr("password_history").ForEachElement(
		func(_ cty.Value, passwordHistory cty.Value) (stop bool) {
			passwordHistoryOption := make(map[string]interface{})

			if enable := value.Bool(passwordHistory.GetAttr("enable")); enable != nil {
				passwordHistoryOption["enable"] = enable
			}

			if size := value.Int(passwordHistory.GetAttr("size")); size != nil && *size != 0 {
				passwordHistoryOption["size"] = size
			}

			if len(passwordHistoryOption) > 0 {
				options.PasswordHistory = passwordHistoryOption
			}

			return stop
		},
	)

	config.GetAttr("password_no_personal_info").ForEachElement(
		func(_ cty.Value, passwordNoPersonalInfo cty.Value) (stop bool) {
			if enable := value.Bool(passwordNoPersonalInfo.GetAttr("enable")); enable != nil {
				options.PasswordNoPersonalInfo = map[string]interface{}{
					"enable": enable,
				}
			}

			return stop
		},
	)

	config.GetAttr("password_dictionary").ForEachElement(
		func(_ cty.Value, passwordDictionary cty.Value) (stop bool) {
			passwordDictionaryOption := make(map[string]interface{})

			if enable := value.Bool(passwordDictionary.GetAttr("enable")); enable != nil {
				passwordDictionaryOption["enable"] = enable
			}
			if dictionary := value.Strings(passwordDictionary.GetAttr("dictionary")); dictionary != nil {
				passwordDictionaryOption["dictionary"] = dictionary
			}

			if len(passwordDictionaryOption) > 0 {
				options.PasswordDictionary = passwordDictionaryOption
			}

			return stop
		},
	)

	config.GetAttr("password_complexity_options").ForEachElement(
		func(_ cty.Value, passwordComplexity cty.Value) (stop bool) {
			if minLength := value.Int(passwordComplexity.GetAttr("min_length")); minLength != nil {
				options.PasswordComplexityOptions = map[string]interface{}{
					"min_length": minLength,
				}
			}

			return stop
		},
	)

	config.GetAttr("mfa").ForEachElement(
		func(_ cty.Value, mfa cty.Value) (stop bool) {
			mfaOption := make(map[string]interface{})

			if active := value.Bool(mfa.GetAttr("active")); active != nil {
				mfaOption["active"] = active
			}
			if returnEnrollSettings := value.Bool(mfa.GetAttr("return_enroll_settings")); returnEnrollSettings != nil {
				mfaOption["return_enroll_settings"] = returnEnrollSettings
			}

			if len(mfaOption) > 0 {
				options.MFA = mfaOption
			}

			return stop
		},
	)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if !d.IsNewResource() {
		apiConn, err := api.Connection.Read(d.Id())
		if err != nil {
			return nil, diag.FromErr(err)
		}

		diags := checkForUnmanagedConfigurationSecrets(
			options.GetConfiguration(),
			apiConn.Options.(*management.ConnectionOptions).GetConfiguration(),
		)

		if diags.HasError() {
			return nil, diags
		}
	} else {
		conn.Strategy = auth0.String("auth0")
	}

	conn.Options = options

	return conn, nil
}

// checkForUnmanagedConfigurationSecrets is used to assess keys diff because values are sent back encrypted.
func checkForUnmanagedConfigurationSecrets(configFromTF, configFromAPI map[string]string) diag.Diagnostics {
	var warnings diag.Diagnostics

	for key := range configFromAPI {
		if _, ok := configFromTF[key]; !ok {
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unmanaged Configuration Secret",
				Detail: fmt.Sprintf("Detected a configuration secret not managed through terraform: %q. "+
					"If you proceed, this configuration secret will get deleted. It is required to "+
					"add this configuration secret to your custom database settings to "+
					"prevent unintentionally destructive results.",
					key,
				),
				AttributePath: cty.Path{cty.GetAttrStep{Name: "options.configuration"}},
			})
		}
	}

	return warnings
}
