package connection

import (
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
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
