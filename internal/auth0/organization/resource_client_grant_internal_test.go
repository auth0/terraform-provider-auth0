package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// newOrganizationClientGrantsTestAPI serves totalGrants client grants for
// organization org_1, paginated with the page and per_page query parameters.
func newOrganizationClientGrantsTestAPI(t *testing.T, totalGrants int) *management.Management {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/v2/organizations/org_1/client-grants" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
		if perPage == 0 {
			perPage = 50
		}

		start := page * perPage
		end := min(start+perPage, totalGrants)

		grants := []map[string]string{}
		for i := start; i < end; i++ {
			grants = append(grants, map[string]string{"id": fmt.Sprintf("cgr_%d", i)})
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"client_grants": grants,
			"start":         start,
			"limit":         perPage,
			"total":         totalGrants,
		})
	}))
	t.Cleanup(server.Close)

	api, err := management.New(
		strings.TrimPrefix(server.URL, "http://"),
		management.WithStaticToken("test-token"),
		management.WithInsecure(),
	)
	require.NoError(t, err)

	return api
}

func TestReadOrganizationClientGrantPastFirstPage(t *testing.T) {
	api := newOrganizationClientGrantsTestAPI(t, 120)

	data := NewOrganizationClientGrantResource().TestResourceData()
	data.SetId("org_1::cgr_110")
	require.NoError(t, data.Set("organization_id", "org_1"))
	require.NoError(t, data.Set("grant_id", "cgr_110"))

	diags := readOrganizationClientGrant(context.Background(), data, config.New(api))

	require.False(t, diags.HasError(), "unexpected diagnostics: %v", diags)
	assert.Equal(t, "org_1::cgr_110", data.Id())
}

func TestFetchAllOrganizationClientGrantsReadsEveryPage(t *testing.T) {
	api := newOrganizationClientGrantsTestAPI(t, 120)

	clientGrants, err := fetchAllOrganizationClientGrants(context.Background(), api, "org_1")

	require.NoError(t, err)
	require.Len(t, clientGrants, 120)
	assert.Equal(t, "cgr_119", clientGrants[119].GetID())
}
