package selfserviceprofile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenSelfServiceProfile(data *schema.ResourceData, selfServiceProfile *management.SelfServiceProfile) error {
	result := multierror.Append(
		data.Set("user_attributes", flattenUserAttributes(selfServiceProfile.UserAttributes)),
		data.Set("branding", flattenBranding(selfServiceProfile.GetBranding())),
		data.Set("created_at", selfServiceProfile.GetCreatedAt().String()),
		data.Set("updated_at", selfServiceProfile.GetUpdatedAt().String()),
	)

	return result.ErrorOrNil()
}

func flattenUserAttributes(userAttributes []*management.SelfServiceProfileUserAttributes) []interface{} {
	var result []interface{}

	for _, userAttribute := range userAttributes {
		result = append(result, map[string]interface{}{
			"name":        userAttribute.GetName(),
			"description": userAttribute.GetDescription(),
			"is_optional": userAttribute.GetIsOptional(),
		})
	}

	return result
}

func flattenBranding(branding *management.Branding) []interface{} {
	if branding == nil {
		return nil
	}
	return []interface{}{
		map[string]interface{}{
			"logo_url": branding.GetLogoURL(),
			"colors":   flattenBrandingColors(branding.GetColors()),
		},
	}
}

func flattenBrandingColors(brandingColors *management.BrandingColors) []interface{} {
	if brandingColors == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"primary": brandingColors.GetPrimary(),
		},
	}
}
