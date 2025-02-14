package branding

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenBranding(
	data *schema.ResourceData,
	branding *management.Branding,
	template *management.BrandingUniversalLogin,
) error {
	result := multierror.Append(
		data.Set("favicon_url", branding.GetFaviconURL()),
		data.Set("logo_url", branding.GetLogoURL()),
		data.Set("colors", flattenBrandingColors(branding.GetColors())),
		data.Set("font", flattenBrandingFont(branding.GetFont())),
		data.Set("universal_login", flattenUniversalLoginTemplate(template)),
	)

	return result.ErrorOrNil()
}

func flattenBrandingTheme(data *schema.ResourceData, brandingTheme *management.BrandingTheme) error {
	result := multierror.Append(
		data.Set("display_name", brandingTheme.GetDisplayName()),
		data.Set("borders", flattenBrandingThemeBorders(brandingTheme.Borders)),
		data.Set("colors", flattenBrandingThemeColors(brandingTheme.Colors)),
		data.Set("fonts", flattenBrandingThemeFonts(brandingTheme.Fonts)),
		data.Set("page_background", flattenBrandingThemePageBackground(brandingTheme.PageBackground)),
		data.Set("widget", flattenBrandingThemeWidget(brandingTheme.Widget)),
	)

	return result.ErrorOrNil()
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

func flattenUniversalLoginTemplate(template *management.BrandingUniversalLogin) []interface{} {
	if template == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"body": template.GetBody(),
		},
	}
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

func flattenPhoneProvider(data *schema.ResourceData, phoneProvider *management.BrandingPhoneProvider) error {
	result := multierror.Append(
		data.Set("name", phoneProvider.GetName()),
		data.Set("disabled", phoneProvider.GetDisabled()),
		data.Set("credentials", flattenPhoneProviderCredentials(data)),
		data.Set("configuration", flattenPhoneProviderConfiguration(phoneProvider.GetConfiguration())),
		data.Set("tenant", phoneProvider.GetTenant()),
		data.Set("channel", phoneProvider.GetChannel()),
	)

	return result.ErrorOrNil()
}

func flattenPhoneProviderConfiguration(configuration *management.BrandingPhoneProviderConfiguration) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"delivery_methods": configuration.GetDeliveryMethods(),
			"sid":              configuration.GetSID(),
			"mssid":            configuration.GetMSSID(),
			"default_from":     configuration.GetDefaultFrom(),
		},
	}
}

func flattenPhoneProviderCredentials(data *schema.ResourceData) []interface{} {
	credentials := []interface{}{
		map[string]interface{}{
			"auth_token": data.Get("credentials.0.auth_token").(string),
		},
	}

	return credentials
}
