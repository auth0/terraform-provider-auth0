package branding

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// newFreeTenantWithCustomDomain serves the Management API of a tenant on a plan without
// Universal Login templates that has a verified custom domain: branding reads and writes
// succeed, and every request to the template answers with templateStatus.
func newFreeTenantWithCustomDomain(t *testing.T, templateStatus int) *config.Config {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/api/v2/branding/templates/universal-login":
			w.WriteHeader(templateStatus)
			_, _ = fmt.Fprintf(w, `{"statusCode":%d,"error":%q,"message":"A paid subscription is required for this feature."}`,
				templateStatus, http.StatusText(templateStatus))
		case "/api/v2/branding":
			_, _ = w.Write([]byte(`{"favicon_url":"https://example.com/favicon.png","logo_url":"https://example.com/logo.png"}`))
		case "/api/v2/custom-domains":
			_, _ = w.Write([]byte(`[{"custom_domain_id":"cd_1","domain":"login.example.com","status":"ready"}]`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)

	api, err := management.New(strings.TrimPrefix(server.URL, "http://"),
		management.WithStaticToken("test-token"), management.WithInsecure())
	require.NoError(t, err)

	return config.New(api)
}

func TestReadBrandingOnATenantWithoutTemplates(t *testing.T) {
	t.Run("it reads the branding when the template answers 402", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewResource().Schema, nil)
		data.SetId("branding")

		diags := readBranding(context.Background(), data, newFreeTenantWithCustomDomain(t, http.StatusPaymentRequired))

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
		assert.Equal(t, "https://example.com/favicon.png", data.Get("favicon_url"))
		assert.Equal(t, "https://example.com/logo.png", data.Get("logo_url"))
		assert.Empty(t, data.Get("universal_login"))
	})

	t.Run("it still fails when the template answers another error", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewResource().Schema, nil)
		data.SetId("branding")

		diags := readBranding(context.Background(), data, newFreeTenantWithCustomDomain(t, http.StatusForbidden))

		assert.True(t, diags.HasError())
	})
}

func TestDeleteBrandingOnATenantWithoutTemplates(t *testing.T) {
	t.Run("it deletes cleanly when the template answers 402", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewResource().Schema, nil)
		data.SetId("branding")

		diags := deleteBranding(context.Background(), data, newFreeTenantWithCustomDomain(t, http.StatusPaymentRequired))

		assert.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	})

	t.Run("it still fails when the template answers another error", func(t *testing.T) {
		data := schema.TestResourceDataRaw(t, NewResource().Schema, nil)
		data.SetId("branding")

		diags := deleteBranding(context.Background(), data, newFreeTenantWithCustomDomain(t, http.StatusForbidden))

		assert.True(t, diags.HasError())
	})
}
