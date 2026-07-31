package attackprotection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
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
			consequence: entitlementReadConsequence,
			wantDetail: "Bot Detection requires an add-on entitlement not present on this tenant, " +
				"so its current configuration could not be read. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "captcha read",
			feature:     "Captcha",
			consequence: entitlementReadConsequence,
			wantDetail: "Captcha requires an add-on entitlement not present on this tenant, " +
				"so its current configuration could not be read. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "bot detection update",
			feature:     "Bot Detection",
			consequence: entitlementUpdateConsequence,
			wantDetail: "Bot Detection requires an add-on entitlement not present on this tenant, " +
				"so the configuration was not applied. " +
				"Contact Auth0 support to enable this feature.",
		},
		{
			name:        "captcha update",
			feature:     "Captcha",
			consequence: entitlementUpdateConsequence,
			wantDetail: "Captcha requires an add-on entitlement not present on this tenant, " +
				"so the configuration was not applied. " +
				"Contact Auth0 support to enable this feature.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			warning := entitlementWarning(testCase.feature, testCase.consequence)

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
	read := entitlementWarning("Bot Detection", entitlementReadConsequence)
	update := entitlementWarning("Bot Detection", entitlementUpdateConsequence)

	assert.NotEqual(t, read.Detail, update.Detail)
	assert.NotContains(t, read.Detail, "not applied")
	assert.Contains(t, update.Detail, "was not applied")
}

// TestEntitlementWarningDoesNotEchoBackendMessage guards against regressing to
// the API's own message text, which has a known backend copy-paste bug where
// the captcha endpoint's 403 message mentions "bot detection".
func TestEntitlementWarningDoesNotEchoBackendMessage(t *testing.T) {
	for _, consequence := range []string{entitlementReadConsequence, entitlementUpdateConsequence} {
		warning := entitlementWarning("Captcha", consequence)

		assert.NotContains(t, warning.Detail, "bot detection")
		assert.NotContains(t, warning.Detail, "Bot Detection")
		assert.NotContains(t, warning.Detail, "upgrade your subscription")
	}
}
