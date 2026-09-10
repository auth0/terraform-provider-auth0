package guardian

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

// resourceDataWithRawConfig builds a *schema.ResourceData whose GetRawConfig()
// reflects the given top-level block values and whose HasChange() reports the
// given keys as changed. This mirrors how Terraform populates RawConfig and the
// diff on a real plan/apply; TestResourceDataRaw alone leaves RawConfig null,
// which the expanders read.
func resourceDataWithRawConfig(
	t *testing.T,
	blocks map[string]cty.Value,
	changedKeys ...string,
) *schema.ResourceData {
	t.Helper()

	sm := schema.InternalMap(NewResource().Schema)
	attrTypes := sm.CoreConfigSchema().ImpliedType().AttributeTypes()

	vals := make(map[string]cty.Value, len(attrTypes))
	for name, ty := range attrTypes {
		vals[name] = cty.NullVal(ty)
	}
	for name, val := range blocks {
		vals[name] = val
	}

	diff := &terraform.InstanceDiff{
		RawConfig:  cty.ObjectVal(vals),
		Attributes: map[string]*terraform.ResourceAttrDiff{},
	}
	for _, key := range changedKeys {
		diff.Attributes[key+".#"] = &terraform.ResourceAttrDiff{Old: "0", New: "1"}
	}

	data, err := sm.Data(nil, diff)
	assert.NoError(t, err)

	data.MarkNewResource()

	return data
}

func settingsBlock(displayCheckbox, rememberMe bool, inactivity, overall int) cty.Value {
	return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"display_remember_me_checkbox":   cty.BoolVal(displayCheckbox),
		"remember_me_default_value":      cty.BoolVal(rememberMe),
		"mfa_session_inactivity_timeout": cty.NumberIntVal(int64(inactivity)),
		"mfa_session_overall_timeout":    cty.NumberIntVal(int64(overall)),
	})})
}

func otpSettingsBlock(otpLength, otpExpirationTime int) cty.Value {
	return cty.ListVal([]cty.Value{cty.ObjectVal(map[string]cty.Value{
		"otp_length":          cty.NumberIntVal(int64(otpLength)),
		"otp_expiration_time": cty.NumberIntVal(int64(otpExpirationTime)),
	})})
}

func TestExpandGuardianSettings(t *testing.T) {
	t.Run("it returns nil when the settings block is not changed", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, nil)

		assert.False(t, data.HasChange("settings"))
		assert.Nil(t, expandGuardianSettings(data))
	})

	t.Run("it returns the populated request when the settings block is configured", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"settings": settingsBlock(false, true, 7200, 86400),
		}, "settings")

		request := expandGuardianSettings(data)

		assert.NotNil(t, request)
		assert.False(t, request.DisplayRememberMeCheckbox)
		assert.True(t, request.RememberMeDefaultValue)
		assert.Equal(t, 7200, request.MfaSessionInactivityTimeout)
		assert.Equal(t, 86400, request.MfaSessionOverallTimeout)
	})

	t.Run("it sends the API defaults verbatim when they are configured explicitly", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"settings": settingsBlock(true, false, 604800, 2592000),
		}, "settings")

		request := expandGuardianSettings(data)

		assert.NotNil(t, request)
		assert.True(t, request.DisplayRememberMeCheckbox)
		assert.False(t, request.RememberMeDefaultValue)
		assert.Equal(t, 604800, request.MfaSessionInactivityTimeout)
		assert.Equal(t, 2592000, request.MfaSessionOverallTimeout)
	})
}

func TestExpandPhoneFactorSettings(t *testing.T) {
	t.Run("it returns nil when the phone_settings block is not changed", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, nil)

		assert.Nil(t, expandPhoneFactorSettings(data))
	})

	t.Run("it returns the populated request when phone_settings is configured", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"phone_settings": otpSettingsBlock(8, 600),
		}, "phone_settings")

		request := expandPhoneFactorSettings(data)

		assert.NotNil(t, request)
		assert.Equal(t, 8, request.OtpLength)
		assert.Equal(t, 600, request.OtpExpirationTime)
	})

	t.Run("it ignores a configured email_settings block", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"email_settings": otpSettingsBlock(4, 30),
		}, "email_settings")

		assert.Nil(t, expandPhoneFactorSettings(data))
	})
}

func TestExpandEmailFactorSettings(t *testing.T) {
	t.Run("it returns nil when the email_settings block is not changed", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, nil)

		assert.Nil(t, expandEmailFactorSettings(data))
	})

	t.Run("it returns the populated request when email_settings is configured", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"email_settings": otpSettingsBlock(10, 3600),
		}, "email_settings")

		request := expandEmailFactorSettings(data)

		assert.NotNil(t, request)
		assert.Equal(t, 10, request.OtpLength)
		assert.Equal(t, 3600, request.OtpExpirationTime)
	})

	t.Run("it ignores a configured phone_settings block", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"phone_settings": otpSettingsBlock(6, 300),
		}, "phone_settings")

		assert.Nil(t, expandEmailFactorSettings(data))
	})
}

func TestExpandOTPSettings(t *testing.T) {
	t.Run("it reads whichever block key it is given", func(t *testing.T) {
		data := resourceDataWithRawConfig(t, map[string]cty.Value{
			"phone_settings": otpSettingsBlock(7, 120),
			"email_settings": otpSettingsBlock(9, 240),
		}, "phone_settings", "email_settings")

		phone := expandOTPSettings(data, "phone_settings")
		assert.NotNil(t, phone)
		assert.Equal(t, 7, phone.otpLength)
		assert.Equal(t, 120, phone.otpExpirationTime)

		email := expandOTPSettings(data, "email_settings")
		assert.NotNil(t, email)
		assert.Equal(t, 9, email.otpLength)
		assert.Equal(t, 240, email.otpExpirationTime)
	})
}
