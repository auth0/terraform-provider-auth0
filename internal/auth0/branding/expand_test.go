package branding

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestExpandBrandingThemeIdentifiers(t *testing.T) {
	t.Run("it returns nil when identifiers block is not configured", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewThemeResource().Schema, map[string]interface{}{})

		identifiers := expandBrandingThemeIdentifiers(data)

		assert.Nil(t, identifiers)
	})

	t.Run("it returns the populated struct when identifiers block is configured", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewThemeResource().Schema, map[string]interface{}{
			"identifiers": []interface{}{
				map[string]interface{}{
					"login_display":    "separate",
					"otp_autocomplete": false,
					"phone_display": []interface{}{
						map[string]interface{}{
							"formatting": "regional",
							"masking":    "show_all",
						},
					},
				},
			},
		})

		identifiers := expandBrandingThemeIdentifiers(data)

		assert.NotNil(t, identifiers)
		assert.Equal(t, "separate", identifiers.LoginDisplay)
		assert.Equal(t, false, identifiers.OTPAutocomplete)
		assert.Equal(t, "regional", identifiers.PhoneDisplay.Formatting)
		assert.Equal(t, "show_all", identifiers.PhoneDisplay.Masking)
	})
}
