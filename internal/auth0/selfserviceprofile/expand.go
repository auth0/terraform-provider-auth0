package selfserviceprofile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandSelfServiceProfiles(data *schema.ResourceData) *management.SelfServiceProfile {
	cfg := data.GetRawConfig()

	// Create profile with all fields
	profile := &management.SelfServiceProfile{
		Name:                   value.String(cfg.GetAttr("name")),
		Description:            value.String(cfg.GetAttr("description")),
		AllowedStrategies:      value.Strings(cfg.GetAttr("allowed_strategies")),
		Branding:               expandBranding(cfg.GetAttr("branding")),
		UserAttributeProfileID: value.String(cfg.GetAttr("user_attribute_profile_id")),
		UserAttributes:         expandSelfServiceProfileUserAttributes(cfg.GetAttr("user_attributes")),
	}

	return profile
}

func expandSelfServiceProfileUserAttributes(userAttr cty.Value) []*management.SelfServiceProfileUserAttributes {
	if userAttr.IsNull() {
		return nil
	}

	var attributes []*management.SelfServiceProfileUserAttributes

	userAttr.ForEachElement(func(_ cty.Value, attr cty.Value) (stop bool) {
		attributes = append(attributes, &management.SelfServiceProfileUserAttributes{
			Name:        value.String(attr.GetAttr("name")),
			Description: value.String(attr.GetAttr("description")),
			IsOptional:  value.Bool(attr.GetAttr("is_optional")),
		})
		return false
	})

	return attributes
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
