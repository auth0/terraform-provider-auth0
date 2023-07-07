package page

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/tenant"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

// NewResource will return a new auth0_pages resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		Description: "With this resource you can manage custom HTML for the " +
			"Login, Reset Password, Multi-Factor Authentication and Error pages.",
		CreateContext: createPages,
		ReadContext:   readPages,
		UpdateContext: updatePages,
		DeleteContext: deletePages,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"login": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for customizing the Login page.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether to use the custom Login page HTML (`true`) or the default Auth0 page (`false`).",
						},
						"html": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Customized content for the Login page. " +
								"HTML format with supported [Liquid syntax](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers).",
						},
					},
				},
			},
			"change_password": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for customizing the Password Reset page.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether to use the custom Reset Password HTML (`true`) or the default Auth0 page (`false`).",
						},
						"html": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Customized content for the Reset Password page. " +
								"HTML format with supported [Liquid syntax](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers).",
						},
					},
				},
			},
			"guardian_mfa": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for customizing the Guardian Multi-Factor Authentication page.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"enabled": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether to use the custom Guardian MFA HTML (`true`) or the default Auth0 page (`false`).",
						},
						"html": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Customized content for the Guardian MFA page. " +
								"HTML format with supported [Liquid syntax](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers).",
						},
					},
				},
			},
			"error": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for the Error pages.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"html": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "Customized content for the Error page. " +
								"HTML format with supported [Liquid syntax](https://github.com/Shopify/liquid/wiki/Liquid-for-Designers).",
						},
						"show_log_link": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Indicates whether to show the link to logs as part of the default error page.",
						},
						"url": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
							Description:  "URL to redirect to when an error occurs, instead of showing the default error page.",
						},
					},
				},
			},
		},
	}
}

func createPages(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updatePages(ctx, data, meta)
}

func readPages(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	globalClientID, err := fetchGlobalClientID(api)
	if err != nil {
		return diag.FromErr(err)
	}

	clientWithLoginPage, err := api.Client.Read(globalClientID)
	if err != nil {
		return diag.FromErr(err)
	}

	tenantPages, err := api.Tenant.Read()
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("login", flattenLoginPage(clientWithLoginPage)),
		data.Set("change_password", tenant.FlattenTenantChangePassword(tenantPages.GetChangePassword())),
		data.Set("guardian_mfa", tenant.FlattenTenantGuardianMFAPage(tenantPages.GetGuardianMFAPage())),
		data.Set("error", tenant.FlattenTenantErrorPage(tenantPages.GetErrorPage())),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updatePages(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if clientWithLoginPage := expandLoginPage(data); clientWithLoginPage != nil {
		globalClientID, err := fetchGlobalClientID(api)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := api.Client.Update(globalClientID, clientWithLoginPage); err != nil {
			return diag.FromErr(err)
		}
	}

	if tenantPages := expandTenantPages(data.GetRawConfig()); tenantPages != nil {
		if err := api.Tenant.Update(tenantPages); err != nil {
			return diag.FromErr(err)
		}
	}

	return readPages(ctx, data, meta)
}

func deletePages(_ context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	globalClientID, err := fetchGlobalClientID(api)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Client.Update(globalClientID, &management.Client{
		CustomLoginPageOn: auth0.Bool(false),
	}); err != nil {
		return diag.FromErr(err)
	}

	if err := api.Tenant.Update(&management.Tenant{
		ChangePassword: &management.TenantChangePassword{
			Enabled: auth0.Bool(false),
		},
		ErrorPage: &management.TenantErrorPage{
			ShowLogLink: auth0.Bool(false),
			URL:         auth0.String(""),
			HTML:        auth0.String(""),
		},
		GuardianMFAPage: &management.TenantGuardianMFAPage{
			Enabled: auth0.Bool(false),
		},
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func fetchGlobalClientID(api *management.Management) (string, error) {
	clientList, err := api.Client.List(
		management.Parameter("is_global", "true"),
		management.IncludeFields("client_id"),
	)
	if err != nil {
		return "", err
	}

	if len(clientList.Clients) == 0 {
		return "", fmt.Errorf("no global client found")
	}

	return clientList.Clients[0].GetClientID(), nil
}
