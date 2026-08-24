package connection

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenConnectionOptions(t *testing.T) {
	result, diags := flattenConnectionOptions(nil, nil)

	if diags != nil {
		t.Errorf("Expected nil diagnostics, got %v", diags)
	}
	if result != nil {
		t.Errorf("Expected nil, got %v", result)
	}
}

func TestFlattenConnectionOptionsEmail(t *testing.T) {
	// Invalid Authentication Params.
	invalidAuthParams := "some non-map value"
	_, diags := flattenConnectionOptionsEmail(nil, &management.ConnectionOptionsEmail{
		AuthParams: invalidAuthParams,
	})

	if len(diags) != 1 {
		t.Errorf("Expected one diagnostic warning, got %d", len(diags))
	}

	if diags[0].Severity != diag.Warning {
		t.Errorf("Expected warning severity, got %v", diags[0].Severity)
	}

	if diags[0].Summary != "Unable to cast auth_params to map[string]string" {
		t.Errorf("Expected specific warning summary, got %q", diags[0].Summary)
	}

	// Valid Authentication Params.
	validAuthParams := map[string]interface{}{
		"foo": "bar",
		"bar": "baz",
	}
	_, diags = flattenConnectionOptionsEmail(nil, &management.ConnectionOptionsEmail{
		AuthParams: validAuthParams,
	})

	if len(diags) != 0 {
		t.Errorf("Expected no diagnostic warnings, got %v", diags)
	}
}

func TestFlattenAuthenticationMethodPassword(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := flattenAuthenticationMethodPassword(nil)
		if result != nil {
			t.Errorf("Expected nil for nil input, got %v", result)
		}
	})

	t.Run("fields explicitly set are echoed back unchanged", func(t *testing.T) {
		apiBehavior := "optional"
		signupBehavior := "block"
		enabled := true

		result := flattenAuthenticationMethodPassword(&management.PasswordAuthenticationMethod{
			Enabled:        &enabled,
			APIBehavior:    &apiBehavior,
			SignupBehavior: &signupBehavior,
		})

		m := result.([]map[string]interface{})[0]
		if m["api_behavior"] != "optional" {
			t.Errorf("Expected api_behavior=optional, got %q", m["api_behavior"])
		}
		if m["signup_behavior"] != "block" {
			t.Errorf("Expected signup_behavior=block, got %q", m["signup_behavior"])
		}
		if m["enabled"] != true {
			t.Errorf("Expected enabled=true, got %v", m["enabled"])
		}
	})

	t.Run("api_behavior defaults to required when API omits the field", func(t *testing.T) {
		result := flattenAuthenticationMethodPassword(&management.PasswordAuthenticationMethod{})

		m := result.([]map[string]interface{})[0]
		if m["api_behavior"] != "required" {
			t.Errorf("Expected api_behavior default=required, got %q", m["api_behavior"])
		}
	})

	t.Run("signup_behavior defaults to allow when API omits the field", func(t *testing.T) {
		result := flattenAuthenticationMethodPassword(&management.PasswordAuthenticationMethod{})

		m := result.([]map[string]interface{})[0]
		if m["signup_behavior"] != "allow" {
			t.Errorf("Expected signup_behavior default=allow, got %q", m["signup_behavior"])
		}
	})
}

func TestFlattenConnectionCrossAppAccessResourceApp(t *testing.T) {
	t.Run("returns nil when resource app is nil", func(t *testing.T) {
		assert.Nil(t, flattenConnectionCrossAppAccessResourceApp(nil))
	})

	t.Run("flattens status enabled", func(t *testing.T) {
		result := flattenConnectionCrossAppAccessResourceApp(&management.CrossAppAccessResourceApp{
			Status: auth0.String("enabled"),
		})

		assert.Len(t, result, 1)
		flat, ok := result[0].(map[string]interface{})
		assert.True(t, ok, "expected result[0] to be a map[string]interface{}")
		assert.Equal(t, "enabled", flat["status"])
	})

	t.Run("flattens status disabled", func(t *testing.T) {
		result := flattenConnectionCrossAppAccessResourceApp(&management.CrossAppAccessResourceApp{
			Status: auth0.String("disabled"),
		})

		assert.Len(t, result, 1)
		flat, ok := result[0].(map[string]interface{})
		assert.True(t, ok, "expected result[0] to be a map[string]interface{}")
		assert.Equal(t, "disabled", flat["status"])
	})
}

// flattenedOptions reads back the single options block written to state.
func flattenedOptions(t *testing.T, data *schema.ResourceData) map[string]interface{} {
	t.Helper()

	rawOptions, ok := data.Get("options").([]interface{})
	assert.True(t, ok, "expected options to be a []interface{}")
	assert.Len(t, rawOptions, 1)

	options, ok := rawOptions[0].(map[string]interface{})
	assert.True(t, ok, "expected options[0] to be a map[string]interface{}")

	return options
}

func TestHideConnectionOptionsClientSecret(t *testing.T) {
	t.Run("blanks client_secret but keeps the rest of options", func(t *testing.T) {
		data := NewDataSource().TestResourceData()
		assert.NoError(t, data.Set("hide_client_secret", true))

		diags := flattenConnectionForDataSource(data, &management.Connection{
			Name:     auth0.String("my-google"),
			Strategy: auth0.String("google-oauth2"),
			Options: &management.ConnectionOptionsGoogleOAuth2{
				ClientID:     auth0.String("my-client-id"),
				ClientSecret: auth0.String("my-client-secret"),
			},
		}, nil)
		assert.False(t, diags.HasError())

		options := flattenedOptions(t, data)
		assert.Empty(t, options["client_secret"])
		assert.Equal(t, "my-client-id", options["client_id"])
		assert.Equal(t, "my-google", data.Get("name"))
	})

	t.Run("leaves client_secret untouched when hide_client_secret is unset", func(t *testing.T) {
		data := NewDataSource().TestResourceData()

		diags := flattenConnectionForDataSource(data, &management.Connection{
			Name:     auth0.String("my-google"),
			Strategy: auth0.String("google-oauth2"),
			Options: &management.ConnectionOptionsGoogleOAuth2{
				ClientSecret: auth0.String("my-client-secret"),
			},
		}, nil)
		assert.False(t, diags.HasError())

		assert.Equal(t, "my-client-secret", flattenedOptions(t, data)["client_secret"])
	})

	t.Run("leaves the other option secrets alone, they are out of scope for now", func(t *testing.T) {
		data := NewDataSource().TestResourceData()
		assert.NoError(t, data.Set("hide_client_secret", true))

		diags := flattenConnectionForDataSource(data, &management.Connection{
			Name:     auth0.String("my-sms"),
			Strategy: auth0.String("sms"),
			Options: &management.ConnectionOptionsSMS{
				TwilioToken: auth0.String("my-twilio-token"),
				GatewayAuthentication: &management.ConnectionGatewayAuthentication{
					Secret: auth0.String("my-gateway-secret"),
				},
			},
		}, nil)
		assert.False(t, diags.HasError())

		options := flattenedOptions(t, data)
		assert.Equal(t, "my-twilio-token", options["twilio_token"])

		gateway, ok := options["gateway_authentication"].([]interface{})
		assert.True(t, ok, "expected gateway_authentication to be a []interface{}")
		assert.Len(t, gateway, 1)

		gatewayMap, ok := gateway[0].(map[string]interface{})
		assert.True(t, ok, "expected gateway_authentication[0] to be a map[string]interface{}")
		assert.Equal(t, "my-gateway-secret", gatewayMap["secret"])
	})

	t.Run("is a no-op for a strategy whose options carry no client_secret", func(t *testing.T) {
		data := NewDataSource().TestResourceData()
		assert.NoError(t, data.Set("hide_client_secret", true))

		diags := flattenConnectionForDataSource(data, &management.Connection{
			Name:     auth0.String("my-sms"),
			Strategy: auth0.String("sms"),
			Options: &management.ConnectionOptionsSMS{
				TwilioSID: auth0.String("my-twilio-sid"),
			},
		}, nil)
		assert.False(t, diags.HasError())

		assert.Equal(t, "my-twilio-sid", flattenedOptions(t, data)["twilio_sid"])
	})

	t.Run("tolerates a connection with no options at all", func(t *testing.T) {
		data := NewDataSource().TestResourceData()
		assert.NoError(t, data.Set("hide_client_secret", true))

		diags := flattenConnectionForDataSource(data, &management.Connection{
			Name:     auth0.String("my-connection"),
			Strategy: auth0.String("auth0"),
		}, nil)

		assert.False(t, diags.HasError())
	})

	// The resource needs the secret in state to manage it.
	t.Run("leaves the resource path untouched", func(t *testing.T) {
		data := NewResource().TestResourceData()
		assert.Nil(t, data.Get("hide_client_secret"))

		diags := flattenConnection(data, &management.Connection{
			Name:     auth0.String("my-github"),
			Strategy: auth0.String("github"),
			Options: &management.ConnectionOptionsGitHub{
				ClientID:     auth0.String("my-client-id"),
				ClientSecret: auth0.String("my-client-secret"),
			},
		})
		assert.False(t, diags.HasError())

		assert.Equal(t, "my-client-secret", flattenedOptions(t, data)["client_secret"])
	})
}
