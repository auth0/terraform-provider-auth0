package networkacl

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
)

// fakeAPIError implements management.Error so the 403 enrichment path (Block 6 /
// D2) can be exercised without an unentitled tenant.
type fakeAPIError struct {
	status int
	msg    string
}

func (e *fakeAPIError) Error() string     { return e.msg }
func (e *fakeAPIError) Status() int       { return e.status }
func (e *fakeAPIError) String() string    { return e.msg }
func (e *fakeAPIError) ErrorCode() string { return "insufficient_scope" }

func aclWithAuth0Managed(match, notMatch []string) *management.NetworkACL {
	acl := &management.NetworkACL{Rule: &management.NetworkACLRule{}}
	if match != nil {
		acl.Rule.Match = &management.NetworkACLRuleMatch{Auth0Managed: &match}
	}
	if notMatch != nil {
		acl.Rule.NotMatch = &management.NetworkACLRuleMatch{Auth0Managed: &notMatch}
	}
	return acl
}

const (
	entitlementHint = "advanced-breached-password-detection"
	flagHint        = "tenant_acl_curated_blocklists"
)

func TestEnrichAuth0ManagedError(t *testing.T) {
	var forbidden management.Error = &fakeAPIError{status: http.StatusForbidden, msg: "403 Forbidden"}
	var notFound management.Error = &fakeAPIError{status: http.StatusNotFound, msg: "404 Not Found"}

	tests := []struct {
		name       string
		acl        *management.NetworkACL
		err        error
		wantHint   bool
		wantNilErr bool
	}{
		{
			name:     "403 with auth0_managed on match is enriched",
			acl:      aclWithAuth0Managed([]string{"auth0.icloud_relay_proxy"}, nil),
			err:      forbidden,
			wantHint: true,
		},
		{
			name:     "403 with auth0_managed on not_match is enriched",
			acl:      aclWithAuth0Managed(nil, []string{"auth0.low_reputation"}),
			err:      forbidden,
			wantHint: true,
		},
		{
			name:     "403 WITHOUT auth0_managed is left untouched",
			acl:      &management.NetworkACL{Rule: &management.NetworkACLRule{Match: &management.NetworkACLRuleMatch{}}},
			err:      forbidden,
			wantHint: false,
		},
		{
			name:     "non-403 with auth0_managed is left untouched",
			acl:      aclWithAuth0Managed([]string{"auth0.icloud_relay_proxy"}, nil),
			err:      notFound,
			wantHint: false,
		},
		{
			name:     "empty auth0_managed slice does not trigger enrichment",
			acl:      aclWithAuth0Managed([]string{}, nil),
			err:      forbidden,
			wantHint: false,
		},
		{
			name:     "non-API error with auth0_managed is left untouched",
			acl:      aclWithAuth0Managed([]string{"auth0.icloud_relay_proxy"}, nil),
			err:      errors.New("dial tcp: connection refused"),
			wantHint: false,
		},
		{
			name:       "nil error stays nil",
			acl:        aclWithAuth0Managed([]string{"auth0.icloud_relay_proxy"}, nil),
			err:        nil,
			wantNilErr: true,
		},
		{
			name:     "nil ACL does not panic",
			acl:      nil,
			err:      forbidden,
			wantHint: false,
		},
		{
			name:     "ACL with nil Rule does not panic",
			acl:      &management.NetworkACL{},
			err:      forbidden,
			wantHint: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := enrichAuth0ManagedError(test.acl, test.err)

			if test.wantNilErr {
				if got != nil {
					t.Fatalf("expected nil error, got %v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected an error, got nil")
			}

			hasEntitlement := strings.Contains(got.Error(), entitlementHint)
			hasFlag := strings.Contains(got.Error(), flagHint)

			if test.wantHint {
				if !hasEntitlement || !hasFlag {
					t.Errorf("expected hint naming both gates, got: %v", got)
				}
				// The original error must remain unwrappable.
				if !errors.Is(got, test.err) {
					t.Errorf("enriched error no longer wraps the original: %v", got)
				}
			} else {
				if hasEntitlement || hasFlag {
					t.Errorf("did not expect enrichment, got: %v", got)
				}
				if got.Error() != test.err.Error() {
					t.Errorf("error was modified: want %q, got %q", test.err, got)
				}
			}
		})
	}
}
