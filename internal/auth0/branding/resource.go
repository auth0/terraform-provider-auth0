package branding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
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
				Default:     nil,
				MaxItems:    1,
				Description: "Configuration settings to customize the font.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Type:        schema.TypeString,
							Required:    true,
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
							ValidateFunc: internalValidation.UniversalLoginTemplateContainsCorrectTags,
							Description:  "The html template for the New Universal Login Experience.",
						},
					},
				},
			},
		},
	}
}

func createBranding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateBranding(ctx, data, meta)
}

func readBranding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

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

	return diag.FromErr(flattenBranding(data, branding, universalLoginTemplate))
}

func updateBranding(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if branding := expandBranding(data.GetRawConfig()); branding.String() != "{}" {
		if err := api.Branding.Update(ctx, branding); err != nil {
			return diag.FromErr(err)
		}
	}

	if isFontConfigurationNull(data) {
		if err := api.Request(ctx, http.MethodPatch, api.URI("branding"), map[string]interface{}{
			"font": nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if isColorsConfigurationNull(data) && !data.IsNewResource() {
		if err := api.Request(ctx, http.MethodPatch, api.URI("branding"), map[string]interface{}{
			"colors": nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	oldUL, newUL := data.GetChange("universal_login")
	oldUniversalLogin := oldUL.([]interface{})
	newUniversalLogin := newUL.([]interface{})

	// This indicates that a removal of the block happened, and we need to delete the template.
	if len(newUniversalLogin) == 0 && len(oldUniversalLogin) != 0 {
		if err := api.Branding.DeleteUniversalLogin(ctx); err != nil {
			return diag.FromErr(err)
		}

		return readBranding(ctx, data, meta)
	}

	if universalLogin := expandBrandingUniversalLogin(data.GetRawConfig()); universalLogin.GetBody() != "" {
		if err := checkForCustomDomains(ctx, api); err != nil {
			return diag.FromErr(err)
		}

		if err := api.Branding.SetUniversalLogin(ctx, universalLogin); err != nil {
			return diag.FromErr(err)
		}
	}

	return readBranding(ctx, data, meta)
}

func deleteBranding(ctx context.Context, _ *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := checkForCustomDomains(ctx, api); err == nil {
		if err := api.Branding.DeleteUniversalLogin(ctx); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func isFontConfigurationNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("font") {
		return false
	}

	empty := true
	rawConfig := data.GetRawConfig().GetAttr("font")

	if rawConfig.IsNull() || rawConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		url := cfg.GetAttr("url")

		if !url.IsNull() && url.AsString() != "" {
			empty = false
		}
		return stop
	}) {
		// font is explicitly null
		return true
	}

	return empty
}

func isColorsConfigurationNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("colors") {
		return false
	}

	empty := true
	rawConfig := data.GetRawConfig().GetAttr("colors")

	if rawConfig.IsNull() || rawConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		pageBackground := cfg.GetAttr("page_background")
		primary := cfg.GetAttr("primary")

		if (!pageBackground.IsNull() && pageBackground.AsString() != "") ||
			(!primary.IsNull() && primary.AsString() != "") {
			empty = false
		}
		return stop
	}) {
		// colors is explicitly null
		return true
	}

	return empty
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
