package selfserviceprofile

import (
	"bytes"
	"encoding/json"
	"fmt"

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

func flattenSSOCustomText(data *schema.ResourceData, customText map[string]interface{}) error {
	body, err := marshalCustomTextBody(customText)
	if err != nil {
		return err
	}

	return data.Set("body", body)
}

func marshalCustomTextBody(b map[string]interface{}) (string, error) {
	if b == nil {
		return "{}", nil
	}

	bodyBytes, err := json.Marshal(b)
	if err != nil {
		return "", fmt.Errorf("failed to serialize the custom texts to JSON: %w", err)
	}

	var buffer bytes.Buffer
	const jsonIndentation = "    "
	if err := json.Indent(&buffer, bodyBytes, "", jsonIndentation); err != nil {
		return "", fmt.Errorf("failed to format the custom texts JSON: %w", err)
	}

	return buffer.String(), nil
}
