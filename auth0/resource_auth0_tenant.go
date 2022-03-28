package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	internalValidation "github.com/auth0/terraform-provider-auth0/auth0/internal/validation"
)

func newTenant() *schema.Resource {
	return &schema.Resource{
		Create: createTenant,
		Read:   readTenant,
		Update: updateTenant,
		Delete: deleteTenant,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"change_password": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"html": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"guardian_mfa_page": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"html": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"default_audience": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"default_directory": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"error_page": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"html": {
							Type:     schema.TypeString,
							Required: true,
						},
						"show_log_link": {
							Type:     schema.TypeBool,
							Required: true,
						},
						"url": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"friendly_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"picture_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_email": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"support_url": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"allowed_logout_urls": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"session_lifetime": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0.01),
			},
			"sandbox_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"idle_session_lifetime": {
				Type:         schema.TypeFloat,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.FloatAtLeast(0.01),
			},
			"enabled_locales": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
			},
			"flags": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enable_client_connections": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_apis_section": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_pipeline2": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_dynamic_client_registration": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_custom_domain_in_emails": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"universal_login": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_legacy_logs_search_v2": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"disable_clickjack_protection_headers": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"enable_public_signup_user_exists_error": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
						"use_scope_descriptions_for_consent": {
							Type:     schema.TypeBool,
							Optional: true,
							Computed: true,
						},
					},
				},
			},
			"universal_login": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"colors": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
									"page_background": {
										Type:     schema.TypeString,
										Optional: true,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"default_redirection_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.All(
					internalValidation.IsURLWithNoFragment,
					validation.IsURLWithScheme([]string{"https"}),
				),
			},
		},
	}
}

func createTenant(d *schema.ResourceData, m interface{}) error {
	d.SetId(resource.UniqueId())
	return updateTenant(d, m)
}

func readTenant(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	tenant, err := api.Tenant.Read()
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("change_password", flattenTenantChangePassword(tenant.ChangePassword))
	d.Set("guardian_mfa_page", flattenTenantGuardianMFAPage(tenant.GuardianMFAPage))
	d.Set("default_audience", tenant.DefaultAudience)
	d.Set("default_directory", tenant.DefaultDirectory)
	d.Set("default_redirection_uri", tenant.DefaultRedirectionURI)
	d.Set("friendly_name", tenant.FriendlyName)
	d.Set("picture_url", tenant.PictureURL)
	d.Set("support_email", tenant.SupportEmail)
	d.Set("support_url", tenant.SupportURL)
	d.Set("allowed_logout_urls", tenant.AllowedLogoutURLs)
	d.Set("session_lifetime", tenant.SessionLifetime)
	d.Set("idle_session_lifetime", tenant.IdleSessionLifetime)
	d.Set("sandbox_version", tenant.SandboxVersion)
	d.Set("enabled_locales", tenant.EnabledLocales)
	d.Set("error_page", flattenTenantErrorPage(tenant.ErrorPage))
	d.Set("flags", flattenTenantFlags(tenant.Flags))
	d.Set("universal_login", flattenTenantUniversalLogin(tenant.UniversalLogin))

	return nil
}

func updateTenant(d *schema.ResourceData, m interface{}) error {
	tenant := buildTenant(d)
	api := m.(*management.Management)
	if err := api.Tenant.Update(tenant); err != nil {
		return err
	}

	return readTenant(d, m)
}

func deleteTenant(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

func buildTenant(d *schema.ResourceData) *management.Tenant {
	tenant := &management.Tenant{
		DefaultAudience:       String(d, "default_audience"),
		DefaultDirectory:      String(d, "default_directory"),
		DefaultRedirectionURI: String(d, "default_redirection_uri"),
		FriendlyName:          String(d, "friendly_name"),
		PictureURL:            String(d, "picture_url"),
		SupportEmail:          String(d, "support_email"),
		SupportURL:            String(d, "support_url"),
		AllowedLogoutURLs:     Slice(d, "allowed_logout_urls"),
		SessionLifetime:       Float64(d, "session_lifetime"),
		SandboxVersion:        String(d, "sandbox_version"),
		IdleSessionLifetime:   Float64(d, "idle_session_lifetime", IsNewResource(), HasChange()),
		EnabledLocales:        List(d, "enabled_locales").List(),
		ChangePassword:        expandTenantChangePassword(d),
		GuardianMFAPage:       expandTenantGuardianMFAPage(d),
		ErrorPage:             expandTenantErrorPage(d),
		Flags:                 expandTenantFlags(d),
		UniversalLogin:        expandTenantUniversalLogin(d),
	}

	return tenant
}
