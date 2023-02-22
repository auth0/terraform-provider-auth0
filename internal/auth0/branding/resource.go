package branding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
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
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The body of login pages.",
						},
					},
				},
			},
		},
	}
}

func createBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updateBranding(ctx, d, m)
}

func readBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	branding, err := api.Branding.Read()
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("favicon_url", branding.GetFaviconURL()),
		d.Set("logo_url", branding.GetLogoURL()),
	)
	if _, ok := d.GetOk("colors"); ok {
		result = multierror.Append(result, d.Set("colors", flattenBrandingColors(branding.GetColors())))
	}
	if _, ok := d.GetOk("font"); ok {
		result = multierror.Append(result, d.Set("font", flattenBrandingFont(branding.GetFont())))
	}
	if _, ok := d.GetOk("universal_login"); ok {
		if err := checkForCustomDomains(api); err != nil {
			return diag.FromErr(err)
		}

		brandingUniversalLogin, err := flattenBrandingUniversalLogin(api)
		if err != nil {
			return diag.FromErr(err)
		}

		result = multierror.Append(result, d.Set("universal_login", brandingUniversalLogin))
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if branding := expandBranding(d.GetRawConfig()); branding.String() != "{}" {
		if err := api.Branding.Update(branding); err != nil {
			return diag.FromErr(err)
		}
	}

	if universalLogin := expandBrandingUniversalLogin(d.GetRawConfig()); universalLogin.GetBody() != "" {
		if err := checkForCustomDomains(api); err != nil {
			return diag.FromErr(err)
		}

		if err := api.Branding.SetUniversalLogin(universalLogin); err != nil {
			return diag.FromErr(err)
		}
	}

	return readBranding(ctx, d, m)
}

func deleteBranding(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if _, ok := d.GetOk("universal_login"); !ok {
		d.SetId("")
		return nil
	}

	if err := checkForCustomDomains(api); err != nil {
		d.SetId("")
		return diag.Diagnostics{
			{
				Severity: diag.Warning,
				Summary:  "No custom domains configured",
				Detail: "Failed to properly destroy the 'auth0_branding' resource " +
					"because no custom domains are available on the tenant.",
				AttributePath: cty.Path{cty.GetAttrStep{Name: "universal_login"}},
			},
		}
	}

	if err := api.Branding.DeleteUniversalLogin(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandBranding(config cty.Value) *management.Branding {
	branding := &management.Branding{
		FaviconURL: value.String(config.GetAttr("favicon_url")),
		LogoURL:    value.String(config.GetAttr("logo_url")),
		Colors:     expandBrandingColors(config.GetAttr("colors")),
		Font:       expandBrandingFont(config.GetAttr("font")),
	}

	return branding
}

func expandBrandingColors(config cty.Value) *management.BrandingColors {
	var brandingColors management.BrandingColors

	config.ForEachElement(func(_ cty.Value, colors cty.Value) (stop bool) {
		brandingColors.PageBackground = value.String(colors.GetAttr("page_background"))
		brandingColors.Primary = value.String(colors.GetAttr("primary"))
		return stop
	})

	if brandingColors == (management.BrandingColors{}) {
		return nil
	}

	return &brandingColors
}

func expandBrandingFont(config cty.Value) *management.BrandingFont {
	var brandingFont management.BrandingFont

	config.ForEachElement(func(_ cty.Value, font cty.Value) (stop bool) {
		brandingFont.URL = value.String(font.GetAttr("url"))
		return stop
	})

	if brandingFont == (management.BrandingFont{}) {
		return nil
	}

	return &brandingFont
}

func expandBrandingUniversalLogin(config cty.Value) *management.BrandingUniversalLogin {
	var universalLogin management.BrandingUniversalLogin

	config.GetAttr("universal_login").ForEachElement(func(_ cty.Value, ul cty.Value) (stop bool) {
		universalLogin.Body = value.String(ul.GetAttr("body"))
		return stop
	})

	if universalLogin == (management.BrandingUniversalLogin{}) {
		return nil
	}

	return &universalLogin
}

func flattenBrandingColors(brandingColors *management.BrandingColors) []interface{} {
	if brandingColors == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"page_background": brandingColors.GetPageBackground(),
			"primary":         brandingColors.GetPrimary(),
		},
	}
}

func flattenBrandingUniversalLogin(api *management.Management) ([]interface{}, error) {
	universalLogin, err := api.Branding.UniversalLogin()
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			return nil, nil
		}
		return nil, err
	}

	if universalLogin == nil {
		return nil, nil
	}

	flattenedUniversalLogin := []interface{}{
		map[string]interface{}{
			"body": universalLogin.GetBody(),
		},
	}

	return flattenedUniversalLogin, nil
}

func flattenBrandingFont(brandingFont *management.BrandingFont) []interface{} {
	if brandingFont == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"url": brandingFont.GetURL(),
		},
	}
}

func checkForCustomDomains(api *management.Management) error {
	customDomains, err := api.CustomDomain.List()
	if err != nil {
		return err
	}

	if len(customDomains) < 1 {
		return fmt.Errorf(
			"managing the universal login body through the 'auth0_branding' resource requires at least " +
				"one custom domain to be configured for the tenant.\n\n" +
				"Use the 'auth0_custom_domain' resource to set one up.",
		)
	}

	return nil
}
