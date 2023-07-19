package branding

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var errNoCustomDomain = fmt.Errorf(
	"managing the Universal Login body through the 'auth0_branding' resource requires at least one custom domain " +
		"to be configured for the tenant.\n\nUse the 'auth0_custom_domain' resource to set one up",
)

// NewResource will return a new auth0_branding resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createBranding,
		ReadContext:   readBranding,
		UpdateContext: updateBranding,
		DeleteContext: deleteBranding,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "This resource allows you to manage branding within your Auth0 tenant. Auth0 can be customized " +
			"with a look and feel that aligns with your organization's brand requirements and user expectations.",
		Schema: map[string]*schema.Schema{
			"colors": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for colors for branding.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"primary": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Primary button background color in hexadecimal.",
						},
						"page_background": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Background color of login pages in hexadecimal.",
						},
					},
				},
			},
			"favicon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "URL for the favicon.",
			},
			"logo_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "URL of logo for branding.",
			},
			"font": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings to customize the font.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "URL for the custom font.",
						},
					},
				},
			},
			"universal_login": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration settings for Universal Login.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"body": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "The html template for the New Universal Login Experience.",
						},
					},
				},
			},
		},
	}
}

func createBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(id.UniqueId())
	return updateBranding(ctx, d, m)
}

func readBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	branding, err := api.Branding.Read(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	var universalLoginTemplate *management.BrandingUniversalLogin
	if err := checkForCustomDomains(ctx, api); err == nil {
		universalLoginTemplate, err = api.Branding.UniversalLogin(ctx)
		if err != nil && !internalError.IsStatusNotFound(err) {
			return diag.FromErr(err)
		}
	}

	return diag.FromErr(flattenBranding(d, branding, universalLoginTemplate))
}

func updateBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if branding := expandBranding(d.GetRawConfig()); branding.String() != "{}" {
		if err := api.Branding.Update(ctx, branding); err != nil {
			return diag.FromErr(err)
		}
	}

	oldUL, newUL := d.GetChange("universal_login")
	oldUniversalLogin := oldUL.([]interface{})
	newUniversalLogin := newUL.([]interface{})

	// This indicates that a removal of the block happened, and we need to delete the template.
	if len(newUniversalLogin) == 0 && len(oldUniversalLogin) != 0 {
		if err := api.Branding.DeleteUniversalLogin(ctx); err != nil {
			return diag.FromErr(err)
		}

		return readBranding(ctx, d, m)
	}

	if universalLogin := expandBrandingUniversalLogin(d.GetRawConfig()); universalLogin.GetBody() != "" {
		if err := checkForCustomDomains(ctx, api); err != nil {
			return diag.FromErr(err)
		}

		if err := api.Branding.SetUniversalLogin(ctx, universalLogin); err != nil {
			return diag.FromErr(err)
		}
	}

	return readBranding(ctx, d, m)
}

func deleteBranding(ctx context.Context, _ *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := checkForCustomDomains(ctx, api); err != nil {
		if err == errNoCustomDomain {
			return nil
		}

		return diag.FromErr(err)
	}

	if err := api.Branding.DeleteUniversalLogin(ctx); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func checkForCustomDomains(ctx context.Context, api *management.Management) error {
	customDomains, err := api.CustomDomain.List(ctx)
	if err != nil {
		return err
	}

	if len(customDomains) < 1 {
		return errNoCustomDomain
	}

	return nil
}
