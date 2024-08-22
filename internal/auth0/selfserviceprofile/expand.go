package selfserviceprofile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandSelfServiceProfiles(data *schema.ResourceData) *management.SelfServiceProfile {
	cfg := data.GetRawConfig()

	return &management.SelfServiceProfile{
		UserAttributes: expandSelfServiceProfileUserAttributes(cfg.GetAttr("user_attributes")),
		Branding:       expandBranding(cfg.GetAttr("branding")),
	}
}

func expandSelfServiceProfileUserAttributes(userAttr cty.Value) []*management.SelfServiceProfileUserAttributes {
	if userAttr.IsNull() {
		return nil
	}

	SelfServiceProfileUserAttributes := make([]*management.SelfServiceProfileUserAttributes, 0)

	userAttr.ForEachElement(func(_ cty.Value, attr cty.Value) (stop bool) {
		SelfServiceProfileUserAttributes = append(SelfServiceProfileUserAttributes, &management.SelfServiceProfileUserAttributes{
			Name:        value.String(attr.GetAttr("name")),
			Description: value.String(attr.GetAttr("description")),
			IsOptional:  value.Bool(attr.GetAttr("is_optional")),
		})
		return stop
	})

	return SelfServiceProfileUserAttributes
}

func expandBranding(config cty.Value) *management.Branding {
	var branding management.Branding

	config.ForEachElement(func(_ cty.Value, b cty.Value) (stop bool) {
		branding.LogoURL = value.String(b.GetAttr("logo_url"))
		branding.Colors = expandBrandingColors(b.GetAttr("colors"))
		return stop
	})

	if branding == (management.Branding{}) {
		return nil
	}

	return &branding
}

func expandBrandingColors(config cty.Value) *management.BrandingColors {
	var brandingColors management.BrandingColors

	config.ForEachElement(func(_ cty.Value, colors cty.Value) (stop bool) {
		brandingColors.Primary = value.String(colors.GetAttr("primary"))
		return stop
	})

	if brandingColors == (management.BrandingColors{}) {
		return nil
	}

	return &brandingColors
}
