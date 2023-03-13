package branding

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/value"
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
	api := meta.(*management.Management)

	if existingBrandingTheme, err := api.BrandingTheme.Default(); err == nil {
		data.SetId(existingBrandingTheme.GetID())
		return updateBrandingTheme(ctx, data, meta)
	}

	brandingTheme := expandBrandingTheme(data)
	if err := api.BrandingTheme.Create(&brandingTheme); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(brandingTheme.GetID())

	return readBrandingTheme(ctx, data, meta)
}

func readBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	brandingTheme, err := api.BrandingTheme.Default()
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(brandingTheme.GetID())

	result := multierror.Append(
		data.Set("display_name", brandingTheme.GetDisplayName()),
		data.Set("borders", flattenBrandingThemeBorders(brandingTheme.Borders)),
		data.Set("colors", flattenBrandingThemeColors(brandingTheme.Colors)),
		data.Set("fonts", flattenBrandingThemeFonts(brandingTheme.Fonts)),
		data.Set("page_background", flattenBrandingThemePageBackground(brandingTheme.PageBackground)),
		data.Set("widget", flattenBrandingThemeWidget(brandingTheme.Widget)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	brandingTheme := expandBrandingTheme(data)
	if err := api.BrandingTheme.Update(data.Id(), &brandingTheme); err != nil {
		return diag.FromErr(err)
	}

	return readBrandingTheme(ctx, data, meta)
}

func deleteBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	if err := api.BrandingTheme.Delete(data.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}
		}

		return diag.FromErr(err)
	}

	return nil
}

func expandBrandingTheme(data *schema.ResourceData) management.BrandingTheme {
	config := data.GetRawConfig()

	brandingTheme := management.BrandingTheme{
		DisplayName: value.String(config.GetAttr("display_name")),
	}

	brandingTheme.Borders = management.BrandingThemeBorders{
		ButtonBorderRadius: data.Get("borders.0.button_border_radius").(float64),
		ButtonBorderWeight: data.Get("borders.0.button_border_weight").(float64),
		ButtonsStyle:       data.Get("borders.0.buttons_style").(string),
		InputBorderRadius:  data.Get("borders.0.input_border_radius").(float64),
		InputBorderWeight:  data.Get("borders.0.input_border_weight").(float64),
		InputsStyle:        data.Get("borders.0.inputs_style").(string),
		ShowWidgetShadow:   data.Get("borders.0.show_widget_shadow").(bool),
		WidgetBorderWeight: data.Get("borders.0.widget_border_weight").(float64),
		WidgetCornerRadius: data.Get("borders.0.widget_corner_radius").(float64),
	}

	baseFocusColor := data.Get("colors.0.base_focus_color").(string)
	baseHoverColor := data.Get("colors.0.base_hover_color").(string)

	brandingTheme.Colors = management.BrandingThemeColors{
		BaseFocusColor:          &baseFocusColor,
		BaseHoverColor:          &baseHoverColor,
		BodyText:                data.Get("colors.0.body_text").(string),
		Error:                   data.Get("colors.0.error").(string),
		Header:                  data.Get("colors.0.header").(string),
		Icons:                   data.Get("colors.0.icons").(string),
		InputBackground:         data.Get("colors.0.input_background").(string),
		InputBorder:             data.Get("colors.0.input_border").(string),
		InputFilledText:         data.Get("colors.0.input_filled_text").(string),
		InputLabelsPlaceholders: data.Get("colors.0.input_labels_placeholders").(string),
		LinksFocusedComponents:  data.Get("colors.0.links_focused_components").(string),
		PrimaryButton:           data.Get("colors.0.primary_button").(string),
		PrimaryButtonLabel:      data.Get("colors.0.primary_button_label").(string),
		SecondaryButtonBorder:   data.Get("colors.0.secondary_button_border").(string),
		SecondaryButtonLabel:    data.Get("colors.0.secondary_button_label").(string),
		Success:                 data.Get("colors.0.success").(string),
		WidgetBackground:        data.Get("colors.0.widget_background").(string),
		WidgetBorder:            data.Get("colors.0.widget_border").(string),
	}

	brandingTheme.Fonts = management.BrandingThemeFonts{
		FontURL:           data.Get("fonts.0.font_url").(string),
		LinksStyle:        data.Get("fonts.0.links_style").(string),
		ReferenceTextSize: data.Get("fonts.0.reference_text_size").(float64),
	}

	brandingTheme.Fonts.BodyText = management.BrandingThemeText{
		Bold: data.Get("fonts.0.body_text.0.bold").(bool),
		Size: data.Get("fonts.0.body_text.0.size").(float64),
	}

	brandingTheme.Fonts.ButtonsText = management.BrandingThemeText{
		Bold: data.Get("fonts.0.buttons_text.0.bold").(bool),
		Size: data.Get("fonts.0.buttons_text.0.size").(float64),
	}

	brandingTheme.Fonts.InputLabels = management.BrandingThemeText{
		Bold: data.Get("fonts.0.input_labels.0.bold").(bool),
		Size: data.Get("fonts.0.input_labels.0.size").(float64),
	}

	brandingTheme.Fonts.Links = management.BrandingThemeText{
		Bold: data.Get("fonts.0.links.0.bold").(bool),
		Size: data.Get("fonts.0.links.0.size").(float64),
	}

	brandingTheme.Fonts.Subtitle = management.BrandingThemeText{
		Bold: data.Get("fonts.0.subtitle.0.bold").(bool),
		Size: data.Get("fonts.0.subtitle.0.size").(float64),
	}

	brandingTheme.Fonts.Title = management.BrandingThemeText{
		Bold: data.Get("fonts.0.title.0.bold").(bool),
		Size: data.Get("fonts.0.title.0.size").(float64),
	}

	brandingTheme.PageBackground = management.BrandingThemePageBackground{
		BackgroundColor:    data.Get("page_background.0.background_color").(string),
		BackgroundImageURL: data.Get("page_background.0.background_image_url").(string),
		PageLayout:         data.Get("page_background.0.page_layout").(string),
	}

	brandingTheme.Widget = management.BrandingThemeWidget{
		HeaderTextAlignment: data.Get("widget.0.header_text_alignment").(string),
		LogoHeight:          data.Get("widget.0.logo_height").(float64),
		LogoPosition:        data.Get("widget.0.logo_position").(string),
		LogoURL:             data.Get("widget.0.logo_url").(string),
		SocialButtonsLayout: data.Get("widget.0.social_buttons_layout").(string),
	}

	return brandingTheme
}

func flattenBrandingThemeBorders(borders management.BrandingThemeBorders) []interface{} {
	m := map[string]interface{}{
		"buttons_style":        borders.ButtonsStyle,
		"button_border_radius": borders.ButtonBorderRadius,
		"button_border_weight": borders.ButtonBorderWeight,
		"inputs_style":         borders.InputsStyle,
		"input_border_radius":  borders.InputBorderRadius,
		"input_border_weight":  borders.InputBorderWeight,
		"show_widget_shadow":   borders.ShowWidgetShadow,
		"widget_corner_radius": borders.WidgetCornerRadius,
		"widget_border_weight": borders.WidgetBorderWeight,
	}

	return []interface{}{m}
}

func flattenBrandingThemeColors(colors management.BrandingThemeColors) []interface{} {
	m := map[string]interface{}{
		"base_focus_color":          colors.GetBaseFocusColor(),
		"base_hover_color":          colors.GetBaseHoverColor(),
		"body_text":                 colors.BodyText,
		"error":                     colors.Error,
		"header":                    colors.Header,
		"icons":                     colors.Icons,
		"input_background":          colors.InputBackground,
		"input_border":              colors.InputBorder,
		"input_filled_text":         colors.InputFilledText,
		"input_labels_placeholders": colors.InputLabelsPlaceholders,
		"links_focused_components":  colors.LinksFocusedComponents,
		"primary_button":            colors.PrimaryButton,
		"primary_button_label":      colors.PrimaryButtonLabel,
		"secondary_button_border":   colors.SecondaryButtonBorder,
		"secondary_button_label":    colors.SecondaryButtonLabel,
		"success":                   colors.Success,
		"widget_background":         colors.WidgetBackground,
		"widget_border":             colors.WidgetBorder,
	}

	return []interface{}{m}
}

func flattenBrandingThemeFonts(fonts management.BrandingThemeFonts) []interface{} {
	m := map[string]interface{}{
		"body_text": []interface{}{
			map[string]interface{}{
				"bold": fonts.BodyText.Bold,
				"size": fonts.BodyText.Size,
			},
		},
		"buttons_text": []interface{}{
			map[string]interface{}{
				"bold": fonts.ButtonsText.Bold,
				"size": fonts.ButtonsText.Size,
			},
		},
		"font_url": fonts.FontURL,
		"input_labels": []interface{}{
			map[string]interface{}{
				"bold": fonts.InputLabels.Bold,
				"size": fonts.InputLabels.Size,
			},
		},
		"links": []interface{}{
			map[string]interface{}{
				"bold": fonts.Links.Bold,
				"size": fonts.Links.Size,
			},
		},
		"links_style":         fonts.LinksStyle,
		"reference_text_size": fonts.ReferenceTextSize,
		"subtitle": []interface{}{
			map[string]interface{}{
				"bold": fonts.Subtitle.Bold,
				"size": fonts.Subtitle.Size,
			},
		},
		"title": []interface{}{
			map[string]interface{}{
				"bold": fonts.Title.Bold,
				"size": fonts.Title.Size,
			},
		},
	}

	return []interface{}{m}
}

func flattenBrandingThemePageBackground(pageBackground management.BrandingThemePageBackground) []interface{} {
	m := map[string]interface{}{
		"background_color":     pageBackground.BackgroundColor,
		"background_image_url": pageBackground.BackgroundImageURL,
		"page_layout":          pageBackground.PageLayout,
	}

	return []interface{}{m}
}

func flattenBrandingThemeWidget(widget management.BrandingThemeWidget) []interface{} {
	m := map[string]interface{}{
		"header_text_alignment": widget.HeaderTextAlignment,
		"logo_height":           widget.LogoHeight,
		"logo_position":         widget.LogoPosition,
		"logo_url":              widget.LogoURL,
		"social_buttons_layout": widget.SocialButtonsLayout,
	}

	return []interface{}{m}
}
