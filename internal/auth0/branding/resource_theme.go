package branding

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewThemeResource will return a new auth0_branding_theme resource.
func NewThemeResource() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to manage branding themes for your Universal Login page " +
			"within your Auth0 tenant.",
		CreateContext: createBrandingTheme,
		ReadContext:   readBrandingTheme,
		UpdateContext: updateBrandingTheme,
		DeleteContext: deleteBrandingTheme,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"display_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The display name for the branding theme.",
			},
			"borders": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"buttons_style": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"pill", "rounded", "sharp"}, false),
							Optional:     true,
							Default:      "rounded",
							Description:  "Buttons style. Available options: `pill`, `rounded`, `sharp`. Defaults to `rounded`.",
						},
						"button_border_radius": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(1, 10),
							Optional:     true,
							Default:      3.0,
							Description:  "Button border radius. Value needs to be between `1` and `10`. Defaults to `3.0`.",
						},
						"button_border_weight": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(0, 10),
							Optional:     true,
							Default:      1.0,
							Description:  "Button border weight. Value needs to be between `0` and `10`. Defaults to `1.0`.",
						},
						"inputs_style": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"pill", "rounded", "sharp"}, false),
							Optional:     true,
							Default:      "rounded",
							Description:  "Inputs style. Available options: `pill`, `rounded`, `sharp`. Defaults to `rounded`.",
						},
						"input_border_radius": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(0, 10),
							Optional:     true,
							Default:      3.0,
							Description:  "Input border radius. Value needs to be between `0` and `10`. Defaults to `3.0`.",
						},
						"input_border_weight": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(0, 3),
							Optional:     true,
							Default:      1.0,
							Description:  "Input border weight. Value needs to be between `0` and `3`. Defaults to `1.0`.",
						},
						"show_widget_shadow": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Show widget shadow. Defaults to `true`.",
						},
						"widget_corner_radius": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(0, 50),
							Optional:     true,
							Default:      5.0,
							Description:  "Widget corner radius. Value needs to be between `0` and `50`. Defaults to `5.0`.",
						},
						"widget_border_weight": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(0, 10),
							Optional:     true,
							Default:      0.0,
							Description:  "Widget border weight. Value needs to be between `0` and `10`. Defaults to `0.0`.",
						},
					},
				},
			},
			"colors": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"base_focus_color": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#635dff",
							Description: "Base focus color. Defaults to `#635dff`.",
						},
						"base_hover_color": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#000000",
							Description: "Base hover color. Defaults to `#000000`.",
						},
						"body_text": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#1e212a",
							Description: "Body text. Defaults to `#1e212a`.",
						},
						"error": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#d03c38",
							Description: "Error. Defaults to `#d03c38`.",
						},
						"header": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#1e212a",
							Description: "Header. Defaults to `#1e212a`.",
						},
						"icons": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#65676e",
							Description: "Icons. Defaults to `#65676e`.",
						},
						"input_background": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#ffffff",
							Description: "Input background. Defaults to `#ffffff`.",
						},
						"input_border": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#c9cace",
							Description: "Input border. Defaults to `#c9cace`.",
						},
						"input_filled_text": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#000000",
							Description: "Input filled text. Defaults to `#000000`.",
						},
						"input_labels_placeholders": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#65676e",
							Description: "Input labels & placeholders. Defaults to `#65676e`.",
						},
						"links_focused_components": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#635dff",
							Description: "Links & focused components. Defaults to `#635dff`.",
						},
						"primary_button": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#635dff",
							Description: "Primary button. Defaults to `#635dff`.",
						},
						"primary_button_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#ffffff",
							Description: "Primary button label. Defaults to `#ffffff`.",
						},
						"secondary_button_border": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#c9cace",
							Description: "Secondary button border. Defaults to `#c9cace`.",
						},
						"secondary_button_label": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#1e212a",
							Description: "Secondary button label. Defaults to `#1e212a`.",
						},
						"success": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#13a688",
							Description: "Success. Defaults to `#13a688`.",
						},
						"widget_background": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#ffffff",
							Description: "Widget background. Defaults to `#ffffff`.",
						},
						"widget_border": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#c9cace",
							Description: "Widget border. Defaults to `#c9cace`.",
						},
					},
				},
			},
			"fonts": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"body_text": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Body text.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Body text bold. Defaults to `false`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 150),
										Optional:     true,
										Default:      87.5,
										Description:  "Body text size. Value needs to be between `0` and `150`. Defaults to `87.5`.",
									},
								},
							},
						},
						"buttons_text": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Buttons text.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Buttons text bold. Defaults to `false`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 150),
										Optional:     true,
										Default:      100.0,
										Description:  "Buttons text size. Value needs to be between `0` and `150`. Defaults to `100.0`.",
									},
								},
							},
						},
						"font_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Font URL. Defaults to an empty string.",
						},
						"input_labels": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Input labels.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Input labels bold. Defaults to `false`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 150),
										Optional:     true,
										Default:      100.0,
										Description:  "Input labels size. Value needs to be between `0` and `150`. Defaults to `100.0`.",
									},
								},
							},
						},
						"links": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Links.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     true,
										Description: "Links bold. Defaults to `true`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 150),
										Optional:     true,
										Default:      87.5,
										Description:  "Links size. Value needs to be between `0` and `150`. Defaults to `87.5`.",
									},
								},
							},
						},
						"links_style": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"normal", "underlined"}, false),
							Default:      "normal",
							Description:  "Links style. Defaults to `normal`.",
						},
						"reference_text_size": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(12, 24),
							Optional:     true,
							Default:      16.0,
							Description:  "Reference text size. Value needs to be between `12` and `24`. Defaults to `16.0`.",
						},
						"subtitle": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Subtitle.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Subtitle bold. Defaults to `false`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(0, 150),
										Optional:     true,
										Default:      87.5,
										Description:  "Subtitle size. Value needs to be between `0` and `150`. Defaults to `87.5`.",
									},
								},
							},
						},
						"title": {
							Type:        schema.TypeList,
							Required:    true,
							MaxItems:    1,
							Description: "Title.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"bold": {
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										Description: "Title bold. Defaults to `false`.",
									},
									"size": {
										Type:         schema.TypeFloat,
										ValidateFunc: validation.FloatBetween(75, 150),
										Optional:     true,
										Default:      150.0,
										Description:  "Title size. Value needs to be between `75` and `150`. Defaults to `150.0`.",
									},
								},
							},
						},
					},
				},
			},
			"page_background": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"background_color": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "#000000",
							Description: "Background color. Defaults to `#000000`.",
						},
						"background_image_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Background image url. Defaults to an empty string.",
						},
						"page_layout": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"center", "left", "right"}, false),
							Optional:     true,
							Default:      "center",
							Description:  "Page layout. Available options: `center`, `left`, `right`. Defaults to `center`.",
						},
					},
				},
			},
			"widget": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"header_text_alignment": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"center", "left", "right"}, false),
							Optional:     true,
							Default:      "center",
							Description:  "Header text alignment. Available options: `center`, `left`, `right`. Defaults to `center`.",
						},
						"logo_height": {
							Type:         schema.TypeFloat,
							ValidateFunc: validation.FloatBetween(1, 100),
							Optional:     true,
							Default:      52.0,
							Description:  "Logo height. Value needs to be between `1` and `100`. Defaults to `52.0`.",
						},
						"logo_position": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"center", "left", "right", "none"}, false),
							Optional:     true,
							Default:      "center",
							Description:  "Logo position. Available options: `center`, `left`, `right`, `none`. Defaults to `center`.",
						},
						"logo_url": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Logo url. Defaults to an empty string.",
						},
						"social_buttons_layout": {
							Type:         schema.TypeString,
							ValidateFunc: validation.StringInSlice([]string{"bottom", "top"}, false),
							Optional:     true,
							Default:      "bottom",
							Description:  "Social buttons layout. Available options: `bottom`, `top`. Defaults to `bottom`.",
						},
					},
				},
			},
		},
	}
}

func createBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if existingBrandingTheme, err := api.BrandingTheme.Default(ctx); err == nil {
		data.SetId(existingBrandingTheme.GetID())
		return updateBrandingTheme(ctx, data, meta)
	}

	brandingTheme := expandBrandingTheme(data)
	if err := api.BrandingTheme.Create(ctx, &brandingTheme); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(brandingTheme.GetID())

	return readBrandingTheme(ctx, data, meta)
}

func readBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	brandingTheme, err := api.BrandingTheme.Default(ctx)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(brandingTheme.GetID())

	return diag.FromErr(flattenBrandingTheme(data, brandingTheme))
}

func updateBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	brandingTheme := expandBrandingTheme(data)

	if err := api.BrandingTheme.Update(ctx, data.Id(), &brandingTheme); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readBrandingTheme(ctx, data, meta)
}

func deleteBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.BrandingTheme.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
