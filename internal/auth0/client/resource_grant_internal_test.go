package client

import "testing"

// TestNormalizeSubjectType validates that the subject_type normalization
// used by createClientGrant's adoption logic maps an empty value to the
// schema-documented default ("client") and otherwise returns the value
// unchanged. This guards the fix for the silent-drift bug where two
// auth0_client_grant resources differing only in subject_type would
// collapse onto the same grant id.
func TestNormalizeSubjectType(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty defaults to client", in: "", want: "client"},
		{name: "client stays client", in: "client", want: "client"},
		{name: "user stays user", in: "user", want: "user"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeSubjectType(tc.in)
			if got != tc.want {
				t.Fatalf("normalizeSubjectType(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
