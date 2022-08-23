package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newBrandingTheme() *schema.Resource {
	return &schema.Resource{
		Description: "This resource allows you to manage branding themes for your universal login page " +
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "Buttons style.",
						},
						"button_border_radius": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Button border radius.",
						},
						"button_border_weight": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Button border weight.",
						},
						"inputs_style": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Inputs style.",
						},
						"input_border_radius": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Input border radius.",
						},
						"input_border_weight": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Input border weight.",
						},
						"show_widget_shadow": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Show widget shadow.",
						},
						"widget_corner_radius": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Widget corner radius.",
						},
						"widget_border_weight": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Widget border weight.",
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
							Description: "Base focus color.",
						},
						"base_hover_color": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Base hover color.",
						},
						"body_text": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Body text.",
						},
						"error": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Error.",
						},
						"header": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Header.",
						},
						"icons": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Icons.",
						},
						"input_background": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Input background.",
						},
						"input_border": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Input border.",
						},
						"input_filled_text": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Input filled text.",
						},
						"input_labels_placeholders": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Input labels & placeholders.",
						},
						"links_focused_components": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Links & focused components.",
						},
						"primary_button": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Primary button",
						},
						"primary_button_label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Primary button label.",
						},
						"secondary_button_border": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Secondary button border.",
						},
						"secondary_button_label": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Secondary button label.",
						},
						"success": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Success.",
						},
						"widget_background": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Widget background.",
						},
						"widget_border": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Widget border.",
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
										Required:    true,
										Description: "Body text bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Body text size.",
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
										Required:    true,
										Description: "Buttons text bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Buttons text size.",
									},
								},
							},
						},
						"font_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Font URL.",
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
										Required:    true,
										Description: "Input labels bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Input labels size.",
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
										Required:    true,
										Description: "Links bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Links size.",
									},
								},
							},
						},
						"links_style": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Links style.",
						},
						"reference_text_size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Reference text size.",
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
										Required:    true,
										Description: "Subtitle bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Subtitle size.",
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
										Required:    true,
										Description: "Title bold.",
									},
									"size": {
										Type:        schema.TypeInt,
										Required:    true,
										Description: "Title size.",
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
							Required:    true,
							Description: "Background color.",
						},
						"background_image_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Background image url.",
						},
						"page_layout": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Page layout.",
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
							Type:        schema.TypeString,
							Required:    true,
							Description: "Header text alignment.",
						},
						"logo_height": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Logo height.",
						},
						"logo_position": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Logo position.",
						},
						"logo_url": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Logo url.",
						},
						"social_buttons_layout": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Social buttons layout.",
						},
					},
				},
			},
		},
	}
}

func createBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	brandingTheme := expandBrandingTheme(data)
	if err := api.BrandingTheme.Create(&brandingTheme); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(brandingTheme.GetID())

	return readBrandingTheme(ctx, data, meta)
}

func readBrandingTheme(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*management.Management)

	brandingTheme, err := api.BrandingTheme.Read(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}
		}

		return diag.FromErr(err)
	}

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
	brandingTheme := management.BrandingTheme{
		DisplayName: String(data, "display_name"),
	}

	List(data, "borders").Elem(func(d ResourceData) {
		brandingTheme.Borders = management.BrandingThemeBorders{
			ButtonBorderRadius: d.Get("button_border_radius").(int),
			ButtonBorderWeight: d.Get("button_border_weight").(int),
			ButtonsStyle:       d.Get("buttons_style").(string),
			InputBorderRadius:  d.Get("input_border_radius").(int),
			InputBorderWeight:  d.Get("input_border_weight").(int),
			InputsStyle:        d.Get("inputs_style").(string),
			ShowWidgetShadow:   d.Get("show_widget_shadow").(bool),
			WidgetBorderWeight: d.Get("widget_border_weight").(int),
			WidgetCornerRadius: d.Get("widget_corner_radius").(int),
		}
	})

	List(data, "colors").Elem(func(d ResourceData) {
		brandingTheme.Colors = management.BrandingThemeColors{
			BaseFocusColor:          String(d, "base_focus_color"),
			BaseHoverColor:          String(d, "base_hover_color"),
			BodyText:                d.Get("body_text").(string),
			Error:                   d.Get("error").(string),
			Header:                  d.Get("header").(string),
			Icons:                   d.Get("icons").(string),
			InputBackground:         d.Get("input_background").(string),
			InputBorder:             d.Get("input_border").(string),
			InputFilledText:         d.Get("input_filled_text").(string),
			InputLabelsPlaceholders: d.Get("input_labels_placeholders").(string),
			LinksFocusedComponents:  d.Get("links_focused_components").(string),
			PrimaryButton:           d.Get("primary_button").(string),
			PrimaryButtonLabel:      d.Get("primary_button_label").(string),
			SecondaryButtonBorder:   d.Get("secondary_button_border").(string),
			SecondaryButtonLabel:    d.Get("secondary_button_label").(string),
			Success:                 d.Get("success").(string),
			WidgetBackground:        d.Get("widget_background").(string),
			WidgetBorder:            d.Get("widget_border").(string),
		}
	})

	List(data, "fonts").Elem(func(d ResourceData) {
		brandingTheme.Fonts = management.BrandingThemeFonts{
			FontURL:           d.Get("font_url").(string),
			LinksStyle:        d.Get("links_style").(string),
			ReferenceTextSize: d.Get("reference_text_size").(int),
		}

		List(d, "body_text").Elem(func(d ResourceData) {
			brandingTheme.Fonts.BodyText = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})

		List(d, "buttons_text").Elem(func(d ResourceData) {
			brandingTheme.Fonts.ButtonsText = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})

		List(d, "input_labels").Elem(func(d ResourceData) {
			brandingTheme.Fonts.InputLabels = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})

		List(d, "links").Elem(func(d ResourceData) {
			brandingTheme.Fonts.Links = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})

		List(d, "subtitle").Elem(func(d ResourceData) {
			brandingTheme.Fonts.Subtitle = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})

		List(d, "title").Elem(func(d ResourceData) {
			brandingTheme.Fonts.Title = management.BrandingThemeText{
				Bold: d.Get("bold").(bool),
				Size: d.Get("size").(int),
			}
		})
	})

	List(data, "page_background").Elem(func(d ResourceData) {
		brandingTheme.PageBackground = management.BrandingThemePageBackground{
			BackgroundColor:    d.Get("background_color").(string),
			BackgroundImageURL: d.Get("background_image_url").(string),
			PageLayout:         d.Get("page_layout").(string),
		}
	})

	List(data, "widget").Elem(func(d ResourceData) {
		brandingTheme.Widget = management.BrandingThemeWidget{
			HeaderTextAlignment: d.Get("header_text_alignment").(string),
			LogoHeight:          d.Get("logo_height").(int),
			LogoPosition:        d.Get("logo_position").(string),
			LogoURL:             d.Get("logo_url").(string),
			SocialButtonsLayout: d.Get("social_buttons_layout").(string),
		}
	})

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
