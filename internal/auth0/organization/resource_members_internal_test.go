package organization

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// memberDeleteRecorder plays the Management API for DELETE /organizations/{id}/members.
// Like the real API, it rejects any request carrying more than 10 user IDs.
type memberDeleteRecorder struct {
	mu      sync.Mutex
	batches [][]string
}

func (rec *memberDeleteRecorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete || !strings.HasSuffix(r.URL.Path, "/members") {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var body struct {
		Members []string `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rec.mu.Lock()
	rec.batches = append(rec.batches, body.Members)
	rec.mu.Unlock()

	if len(body.Members) > 10 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, `{"statusCode":400,"error":"Bad Request","message":"Payload validation error: 'Array is too long (%d), maximum 10' on property members (List of user IDs to remove from the organization)."}`, len(body.Members))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func newMembersTestConfig(t *testing.T, handler http.Handler) *config.Config {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	api, err := management.New(strings.TrimPrefix(server.URL, "http://"),
		management.WithStaticToken("test-token"), management.WithInsecure(), management.WithNoRetries())
	require.NoError(t, err)

	return config.New(api)
}

func generateMemberIDs(count int) []string {
	ids := make([]string, 0, count)
	for i := 0; i < count; i++ {
		ids = append(ids, fmt.Sprintf("auth0|user%02d", i))
	}

	return ids
}

// TestDeleteOrganizationMembersInBatches destroys an auth0_organization_members
// resource holding 15 members. The API accepts at most 10 IDs per request, so the
// destroy used to fail with "Array is too long (15), maximum 10".
func TestDeleteOrganizationMembersInBatches(t *testing.T) {
	for _, count := range []int{1, 10, 11, 15, 25} {
		t.Run(fmt.Sprintf("%d members", count), func(t *testing.T) {
			recorder := &memberDeleteRecorder{}
			meta := newMembersTestConfig(t, recorder)

			memberIDs := generateMemberIDs(count)
			memberValues := make([]cty.Value, 0, count)
			attributes := map[string]string{
				"id":              "org_test",
				"organization_id": "org_test",
				"members.#":       fmt.Sprint(count),
			}
			for _, id := range memberIDs {
				memberValues = append(memberValues, cty.StringVal(id))
			}

			resource := NewMembersResource()
			data := resource.Data(&terraform.InstanceState{
				ID:         "org_test",
				Attributes: attributes,
				RawState: cty.ObjectVal(map[string]cty.Value{
					"id":              cty.StringVal("org_test"),
					"organization_id": cty.StringVal("org_test"),
					"members":         cty.SetVal(memberValues),
				}),
			})

			diagnostics := deleteOrganizationMembers(context.Background(), data, meta)
			require.False(t, diagnostics.HasError(), "%v", diagnostics)

			var sent []string
			for _, batch := range recorder.batches {
				assert.LessOrEqual(t, len(batch), 10)
				sent = append(sent, batch...)
			}
			assert.ElementsMatch(t, memberIDs, sent, "every member must be removed exactly once")
			assert.Len(t, recorder.batches, (count+9)/10)
		})
	}
}
