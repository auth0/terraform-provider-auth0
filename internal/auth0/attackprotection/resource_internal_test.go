package attackprotection

import (
	"testing"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"

	apierr "github.com/auth0/terraform-provider-auth0/internal/error"
)

func insufficientEntitlementErr() error {
	return &managementv3.ForbiddenError{
		Body: map[string]interface{}{
			"statusCode": float64(403),
			"error":      "Forbidden",
			"message":    "Please upgrade your subscription to use bot detection",
			"errorCode":  "insufficient_entitlement",
		},
	}
}

func TestEntitlementWarning(t *testing.T) {
	var testCases = []struct {
		name        string
		feature     string
		consequence string
		wantDetail  string
	}{
		{
			name:        "bot detection read",
			feature:     "Bot Detection",
			consequence: "its current configuration could not be read",
			wantDetail: `Bot Detection requires an add-on entitlement not present on this tenant, ` +
				`so its current configuration could not be read (error code: "insufficient_entitlement"). ` +
				`Contact Auth0 support to enable this feature.`,
		},
		{
			name:        "captcha read",
			feature:     "Captcha",
			consequence: "its current configuration could not be read",
			wantDetail: `Captcha requires an add-on entitlement not present on this tenant, ` +
				`so its current configuration could not be read (error code: "insufficient_entitlement"). ` +
				`Contact Auth0 support to enable this feature.`,
		},
		{
			name:        "bot detection update",
			feature:     "Bot Detection",
			consequence: "the configuration was not applied",
			wantDetail: `Bot Detection requires an add-on entitlement not present on this tenant, ` +
				`so the configuration was not applied (error code: "insufficient_entitlement"). ` +
				`Contact Auth0 support to enable this feature.`,
		},
		{
			name:        "captcha update",
			feature:     "Captcha",
			consequence: "the configuration was not applied",
			wantDetail: `Captcha requires an add-on entitlement not present on this tenant, ` +
				`so the configuration was not applied (error code: "insufficient_entitlement"). ` +
				`Contact Auth0 support to enable this feature.`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			warning := apierr.EntitlementWarning(testCase.feature, testCase.consequence, insufficientEntitlementErr())

			assert.Equal(t, diag.Warning, warning.Severity)
			assert.Equal(t, testCase.feature+" entitlement not available", warning.Summary)
			assert.Equal(t, testCase.wantDetail, warning.Detail)
		})
	}
}

// TestEntitlementConsequencesAreDistinct guards the read/update split: a read
// failure never attempted a write, so it must not claim the configuration was
// not applied.
func TestEntitlementConsequencesAreDistinct(t *testing.T) {
	err := insufficientEntitlementErr()
	read := apierr.EntitlementWarning("Bot Detection", "its current configuration could not be read", err)
	update := apierr.EntitlementWarning("Bot Detection", "the configuration was not applied", err)

	assert.NotEqual(t, read.Detail, update.Detail)
	assert.NotContains(t, read.Detail, "not applied")
	assert.Contains(t, update.Detail, "was not applied")
}

// TestEntitlementWarningDoesNotEchoBackendMessage guards against regressing to
// the API's own message text, which has a known backend copy-paste bug where
// the captcha endpoint's 403 message mentions "bot detection".
func TestEntitlementWarningDoesNotEchoBackendMessage(t *testing.T) {
	backendMessage := "Please upgrade your subscription to use bot detection"
	err := &managementv3.ForbiddenError{
		Body: map[string]interface{}{
			"statusCode": float64(403),
			"error":      "Forbidden",
			"message":    backendMessage,
			"errorCode":  "insufficient_entitlement",
		},
	}

	for _, consequence := range []string{
		"its current configuration could not be read",
		"the configuration was not applied",
	} {
		warning := apierr.EntitlementWarning("Captcha", consequence, err)

		assert.NotContains(t, warning.Detail, backendMessage)
		assert.NotContains(t, warning.Detail, "bot detection")
		assert.NotContains(t, warning.Detail, "Bot Detection")
		assert.NotContains(t, warning.Detail, "upgrade your subscription")
		assert.Contains(t, warning.Detail, `(error code: "insufficient_entitlement")`)
	}
}
