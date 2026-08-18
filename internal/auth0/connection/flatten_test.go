package connection

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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

func TestFlattenConnectionOptionsOIDCMetadata(t *testing.T) {
	metadata := map[string]interface{}{
		"issuer":   "https://idp.example.com",
		"jwks_uri": "https://idp.example.com/jwks",
	}
	metadataJSON := `{"issuer":"https://idp.example.com","jwks_uri":"https://idp.example.com/jwks"}`

	t.Run("oidc flattens the document", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOIDC(nil, &management.ConnectionOptionsOIDC{
			OIDCMetadata: metadata,
		})

		assert.Nil(t, diags)
		assert.Equal(t, metadataJSON, result.(map[string]interface{})["oidc_metadata"])
	})

	t.Run("okta flattens the document", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOkta(nil, &management.ConnectionOptionsOkta{
			OIDCMetadata: metadata,
		})

		assert.Nil(t, diags)
		assert.Equal(t, metadataJSON, result.(map[string]interface{})["oidc_metadata"])
	})

	// The API omits the field when it was never set, so state must hold "" and not "null".
	t.Run("oidc yields an empty string when the API omits the field", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOIDC(nil, &management.ConnectionOptionsOIDC{})

		assert.Nil(t, diags)
		assert.Equal(t, "", result.(map[string]interface{})["oidc_metadata"])
	})

	t.Run("okta yields an empty string when the API omits the field", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOkta(nil, &management.ConnectionOptionsOkta{})

		assert.Nil(t, diags)
		assert.Equal(t, "", result.(map[string]interface{})["oidc_metadata"])
	})

	// The four booleans the API defaults in are dropped, so state holds only what was asked
	// for. This is what keeps an edit to some other key from also rendering them as removals.
	t.Run("server-defaulted keys are dropped", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOIDC(nil, &management.ConnectionOptionsOIDC{
			OIDCMetadata: enrichedMetadata(),
		})

		assert.Nil(t, diags)
		assert.Equal(t, `{"issuer":"https://idp.example.com"}`, result.(map[string]interface{})["oidc_metadata"])
	})
}

// enrichedMetadata is what the API returns: a one-key document plus the four defaults.
func enrichedMetadata() map[string]interface{} {
	return map[string]interface{}{
		"issuer":                           "https://idp.example.com",
		"claims_parameter_supported":       false,
		"request_parameter_supported":      false,
		"request_uri_parameter_supported":  false,
		"require_request_uri_registration": false,
	}
}

// TestFlattenOIDCMetadataDropsServerDefaults covers why flatten strips the defaults at
// all. DiffSuppressFunc is all-or-nothing per attribute: on a plan where some other key
// genuinely changed it returns false, and Terraform then renders the whole stored document
// against the configuration, showing the four untouched defaults as removals. Keeping them
// out of state is what makes such a plan show only the edited key.
func TestFlattenOIDCMetadataDropsServerDefaults(t *testing.T) {
	t.Run("okta drops them too", func(t *testing.T) {
		result, diags := flattenConnectionOptionsOkta(
			nil,
			&management.ConnectionOptionsOkta{OIDCMetadata: enrichedMetadata()},
		)

		assert.Nil(t, diags)
		assert.Equal(t, `{"issuer":"https://idp.example.com"}`, result.(map[string]interface{})["oidc_metadata"])
	})

	// Drift the practitioner must see: true is a value someone asked for, not a default.
	t.Run("a default the API returned as true is kept", func(t *testing.T) {
		metadata := enrichedMetadata()
		metadata["request_parameter_supported"] = true

		result, diags := flattenConnectionOptionsOIDC(
			nil,
			&management.ConnectionOptionsOIDC{OIDCMetadata: metadata},
		)

		assert.Nil(t, diags)
		assert.Equal(
			t,
			`{"issuer":"https://idp.example.com","request_parameter_supported":true}`,
			result.(map[string]interface{})["oidc_metadata"],
		)
	})

	// Nothing outside the four keys may be dropped, false or not.
	t.Run("unrelated keys are never dropped", func(t *testing.T) {
		metadata := enrichedMetadata()
		metadata["scopes_supported"] = []interface{}{"openid"}
		metadata["backchannel_logout_supported"] = false

		result, diags := flattenConnectionOptionsOIDC(
			nil,
			&management.ConnectionOptionsOIDC{OIDCMetadata: metadata},
		)

		assert.Nil(t, diags)
		assert.Equal(
			t,
			`{"backchannel_logout_supported":false,"issuer":"https://idp.example.com","scopes_supported":["openid"]}`,
			result.(map[string]interface{})["oidc_metadata"],
		)
	})

	// Flatten must not mutate the caller's map, which the resource reads from afterwards.
	t.Run("the API document is not mutated", func(t *testing.T) {
		metadata := enrichedMetadata()

		_, diags := flattenConnectionOptionsOIDC(nil, &management.ConnectionOptionsOIDC{OIDCMetadata: metadata})

		assert.Nil(t, diags)
		assert.Len(t, metadata, len(enrichedMetadata()))
	})
}
