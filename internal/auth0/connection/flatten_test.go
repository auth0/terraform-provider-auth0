package connection

import (
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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

// newResourceDataWithSecret builds a *schema.ResourceData whose
// options.0.client_secret is pre-populated with priorSecret.
func newResourceDataWithSecret(t *testing.T, priorSecret string) *schema.ResourceData {
	t.Helper()
	d := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{})
	if err := d.Set("options", []interface{}{
		map[string]interface{}{"client_secret": priorSecret},
	}); err != nil {
		t.Fatalf("failed to set options: %v", err)
	}
	return d
}

// TestFlattenConnectionOptionsOAuth2_ClientSecretFallback verifies that the
// API response's redacted/empty client_secret is replaced by the value
// already stored in state, preventing spurious diffs.
func TestFlattenConnectionOptionsOAuth2_ClientSecretFallback(t *testing.T) {
	priorSecret := "stored-oauth2-secret"
	d := newResourceDataWithSecret(t, priorSecret)

	rawResult, diags := flattenConnectionOptionsOAuth2(d, &management.ConnectionOptionsOAuth2{
		// Auth0 API returns an empty string for secrets after creation.
		ClientSecret: nil,
	})
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	m, ok := rawResult.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", rawResult)
	}

	if got := m["client_secret"]; got != priorSecret {
		t.Errorf("client_secret = %q, want %q", got, priorSecret)
	}
}

// TestFlattenConnectionOptionsGitHub_ClientSecretFallback performs the same
// check for the GitHub connection strategy.
func TestFlattenConnectionOptionsGitHub_ClientSecretFallback(t *testing.T) {
	priorSecret := "stored-github-secret"
	d := newResourceDataWithSecret(t, priorSecret)

	rawResult, diags := flattenConnectionOptionsGitHub(d, &management.ConnectionOptionsGitHub{
		ClientSecret: nil,
	})
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %v", diags)
	}

	m, ok := rawResult.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", rawResult)
	}

	if got := m["client_secret"]; got != priorSecret {
		t.Errorf("client_secret = %q, want %q", got, priorSecret)
	}
}

