package branding

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandBranding(config cty.Value) *management.Branding {
	return &management.Branding{
		FaviconURL: value.String(config.GetAttr("favicon_url")),
		LogoURL:    value.String(config.GetAttr("logo_url")),
		Colors:     expandBrandingColors(config.GetAttr("colors")),
		Font:       expandBrandingFont(config.GetAttr("font")),
	}
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
		CaptchaWidgetTheme:      data.Get("colors.0.captcha_widget_theme").(string),
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

func expandPhoneProvider(config cty.Value) *management.BrandingPhoneProvider {
	phoneProvider := &management.BrandingPhoneProvider{
		Name:          value.String(config.GetAttr("name")),
		Disabled:      value.Bool(config.GetAttr("disabled")),
		Configuration: expandPhoneProviderConfiguration(config.GetAttr("configuration")),
		Credentials:   expandPhoneProviderCredentials(config.GetAttr("credentials")),
	}

	return phoneProvider
}

func expandPhoneProviderConfiguration(config cty.Value) *management.BrandingPhoneProviderConfiguration {
	var configuration management.BrandingPhoneProviderConfiguration

	config.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		configuration.DeliveryMethods = value.Strings(config.GetAttr("delivery_methods"))
		configuration.SID = value.String(config.GetAttr("sid"))
		configuration.MSSID = value.String(config.GetAttr("mssid"))
		configuration.DefaultFrom = value.String(config.GetAttr("default_from"))

		return stop
	})

	return &configuration
}

func expandPhoneProviderCredentials(config cty.Value) *management.BrandingPhoneProviderCredential {
	credentials := &management.BrandingPhoneProviderCredential{}

	config.ForEachElement(func(_ cty.Value, credential cty.Value) (stop bool) {
		credentials.AuthToken = value.String(credential.GetAttr("auth_token"))
		return stop
	})

	return credentials
}

func expandPhoneNotificationTemplate(config cty.Value) *management.BrandingPhoneNotificationTemplate {
	template := &management.BrandingPhoneNotificationTemplate{
		Type:     value.String(config.GetAttr("type")),
		Disabled: value.Bool(config.GetAttr("disabled")),
		Content:  expandPhoneNotificationTemplateContent(config.GetAttr("content")),
	}

	return template
}

func expandPhoneNotificationTemplateContent(config cty.Value) *management.BrandingPhoneNotificationTemplateContent {
	var content management.BrandingPhoneNotificationTemplateContent

	config.ForEachElement(func(_ cty.Value, c cty.Value) (stop bool) {
		content.Syntax = value.String(c.GetAttr("syntax"))
		content.From = value.String(c.GetAttr("from"))
		content.Body = expandPhoneNotificationTemplateContentBody(c.GetAttr("body"))
		return stop
	})

	if content == (management.BrandingPhoneNotificationTemplateContent{}) {
		return nil
	}

	return &content
}

func expandPhoneNotificationTemplateContentBody(config cty.Value) *management.BrandingPhoneNotificationTemplateContentBody {
	var body management.BrandingPhoneNotificationTemplateContentBody

	config.ForEachElement(func(_ cty.Value, b cty.Value) (stop bool) {
		body.Text = value.String(b.GetAttr("text"))
		body.Voice = value.String(b.GetAttr("voice"))
		return stop
	})

	if body == (management.BrandingPhoneNotificationTemplateContentBody{}) {
		return nil
	}

	return &body
}
