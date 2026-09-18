package attackprotection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"

	apierr "github.com/auth0/terraform-provider-auth0/internal/error"
)

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
			wantDetail: "Bot Detection requires an add-on entitlement not present on this tenant, " +
				"so its current configuration could not be read. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "captcha read",
			feature:     "Captcha",
			consequence: "its current configuration could not be read",
			wantDetail: "Captcha requires an add-on entitlement not present on this tenant, " +
				"so its current configuration could not be read. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "bot detection update",
			feature:     "Bot Detection",
			consequence: "the configuration was not applied",
			wantDetail: "Bot Detection requires an add-on entitlement not present on this tenant, " +
				"so the configuration was not applied. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "captcha update",
			feature:     "Captcha",
			consequence: "the configuration was not applied",
			wantDetail: "Captcha requires an add-on entitlement not present on this tenant, " +
				"so the configuration was not applied. " +
				"Contact Auth0 support to enable this feature.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			warning := apierr.EntitlementWarning(testCase.feature, testCase.consequence)

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
	read := apierr.EntitlementWarning("Bot Detection", "its current configuration could not be read")
	update := apierr.EntitlementWarning("Bot Detection", "the configuration was not applied")

	assert.NotEqual(t, read.Detail, update.Detail)
	assert.NotContains(t, read.Detail, "not applied")
	assert.Contains(t, update.Detail, "was not applied")
}

// TestEntitlementWarningDoesNotEchoBackendMessage guards against regressing to
// the API's own message text, which has a known backend copy-paste bug where
// the captcha endpoint's 403 message mentions "bot detection".
func TestEntitlementWarningDoesNotEchoBackendMessage(t *testing.T) {
	for _, consequence := range []string{
		"its current configuration could not be read",
		"the configuration was not applied",
	} {
		warning := apierr.EntitlementWarning("Captcha", consequence)

		assert.NotContains(t, warning.Detail, "bot detection")
		assert.NotContains(t, warning.Detail, "Bot Detection")
		assert.NotContains(t, warning.Detail, "upgrade your subscription")
	}
}
