package branding

import (
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenBrandingTheme(t *testing.T) {
	mockResourceData := schema.TestResourceDataRaw(t, NewThemeResource().Schema, map[string]interface{}{})

	t.Run("it sets identifiers to nil if remote theme does not have them set", func(t *testing.T) {
		brandingTheme := management.BrandingTheme{
			Identifiers: nil,
		}

		err := flattenBrandingTheme(mockResourceData, &brandingTheme)

		assert.NoError(t, err)
		assert.Equal(t, mockResourceData.Get("identifiers"), []interface{}{})
	})

	t.Run("it sets identifiers if remote theme has them set", func(t *testing.T) {
		brandingTheme := management.BrandingTheme{
			Identifiers: &management.BrandingThemeIdentifiers{
				LoginDisplay:    "unified",
				OTPAutocomplete: true,
				PhoneDisplay: management.BrandingThemePhoneDisplay{
					Formatting: "international",
					Masking:    "mask_digits",
				},
			},
		}

		err := flattenBrandingTheme(mockResourceData, &brandingTheme)

		assert.NoError(t, err)
		identifiers := mockResourceData.Get("identifiers").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "unified", identifiers["login_display"])
		assert.Equal(t, true, identifiers["otp_autocomplete"])
		phoneDisplay := identifiers["phone_display"].([]interface{})[0].(map[string]interface{})
		assert.Equal(t, "international", phoneDisplay["formatting"])
		assert.Equal(t, "mask_digits", phoneDisplay["masking"])
	})
}
