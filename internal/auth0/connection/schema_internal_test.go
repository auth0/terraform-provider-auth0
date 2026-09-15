package connection

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/stretchr/testify/assert"
)

// discoveryDoc is the minimum document the API accepts on an oidc or okta connection.
const discoveryDoc = `{
	"issuer": "https://idp.example.com",
	"jwks_uri": "https://idp.example.com/jwks",
	"authorization_endpoint": "https://idp.example.com/authorize",
	"token_endpoint": "https://idp.example.com/token",
	"response_types_supported": ["code"],
	"subject_types_supported": ["public"],
	"id_token_signing_alg_values_supported": ["RS256"]
}`

// enrichedDiscoveryDoc is what the API stores back: discoveryDoc plus the defaults.
const enrichedDiscoveryDoc = `{
	"issuer": "https://idp.example.com",
	"jwks_uri": "https://idp.example.com/jwks",
	"authorization_endpoint": "https://idp.example.com/authorize",
	"token_endpoint": "https://idp.example.com/token",
	"response_types_supported": ["code"],
	"subject_types_supported": ["public"],
	"id_token_signing_alg_values_supported": ["RS256"],
	"claims_parameter_supported": false,
	"request_parameter_supported": false,
	"request_uri_parameter_supported": false,
	"require_request_uri_registration": false
}`

func TestSuppressOIDCMetadataDiff(t *testing.T) {
	testCases := []struct {
		name     string
		state    string
		config   string
		suppress bool
	}{
		{
			// The case this function exists for.
			name:     "state carries the four server defaults the config omits",
			state:    enrichedDiscoveryDoc,
			config:   discoveryDoc,
			suppress: true,
		},
		{
			name:     "only some of the server defaults were added",
			state:    `{"issuer": "https://idp.example.com", "claims_parameter_supported": false}`,
			config:   `{"issuer": "https://idp.example.com"}`,
			suppress: true,
		},
		{
			name:     "identical documents",
			state:    discoveryDoc,
			config:   discoveryDoc,
			suppress: true,
		},
		{
			name:     "same document reformatted and reordered",
			state:    `{"jwks_uri":"https://idp.example.com/jwks","issuer":"https://idp.example.com"}`,
			config:   "{\n  \"issuer\": \"https://idp.example.com\",\n  \"jwks_uri\": \"https://idp.example.com/jwks\"\n}",
			suppress: true,
		},
		{
			// The API replaces the document wholesale, so a dropped key is a real change.
			name:     "config removes a key",
			state:    `{"issuer": "https://idp.example.com", "end_session_endpoint": "https://idp.example.com/logout"}`,
			config:   `{"issuer": "https://idp.example.com"}`,
			suppress: false,
		},
		{
			// The four keys are settable, so asking for true against false is real drift.
			name:     "config sets a server default to true but state holds false",
			state:    `{"issuer": "https://idp.example.com", "claims_parameter_supported": false}`,
			config:   `{"issuer": "https://idp.example.com", "claims_parameter_supported": true}`,
			suppress: false,
		},
		{
			name:     "state holds a server-default key as true and the config omits it",
			state:    `{"issuer": "https://idp.example.com", "claims_parameter_supported": true}`,
			config:   `{"issuer": "https://idp.example.com"}`,
			suppress: false,
		},
		{
			name:     "a value differs",
			state:    `{"issuer": "https://idp.example.com"}`,
			config:   `{"issuer": "https://other.example.com"}`,
			suppress: false,
		},
		{
			name:     "config adds a key",
			state:    `{"issuer": "https://idp.example.com"}`,
			config:   `{"issuer": "https://idp.example.com", "end_session_endpoint": "https://idp.example.com/logout"}`,
			suppress: false,
		},
		{
			name:     "state adds an unrelated false key",
			state:    `{"issuer": "https://idp.example.com", "some_other_flag": false}`,
			config:   `{"issuer": "https://idp.example.com"}`,
			suppress: false,
		},
		{
			name:     "first write, state empty",
			state:    "",
			config:   discoveryDoc,
			suppress: false,
		},
		{
			name:     "config cleared",
			state:    discoveryDoc,
			config:   "",
			suppress: false,
		},
		{
			name:     "malformed json is never suppressed",
			state:    `{"issuer": "https://idp.example.com"`,
			config:   `{"issuer": "https://idp.example.com"}`,
			suppress: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			assert.Equal(
				t,
				testCase.suppress,
				suppressOIDCMetadataDiff("options.0.oidc_metadata", testCase.state, testCase.config, nil),
			)
		})
	}
}

// TestSuppressOIDCMetadataDiffParityWithJSONDiff guards the blast radius: the attribute
// is shared by every strategy, so away from the one intended departure this has to keep
// behaving like the structure.SuppressJsonDiff it replaced.
func TestSuppressOIDCMetadataDiffParityWithJSONDiff(t *testing.T) {
	documents := []string{
		"",
		discoveryDoc,
		enrichedDiscoveryDoc,
		`{}`,
		`{"issuer": "https://idp.example.com"}`,
		`{"issuer": "https://other.example.com"}`,
		`{"issuer": "https://idp.example.com", "claims_parameter_supported": true}`,
		`{"issuer": "https://idp.example.com", "end_session_endpoint": "https://idp.example.com/logout"}`,
		`{"issuer": "https://idp.example.com"`,
	}

	compared := 0

	for _, state := range documents {
		for _, config := range documents {
			if carriesFalseServerDefault(t, state) || carriesFalseServerDefault(t, config) {
				continue
			}

			compared++

			assert.Equal(
				t,
				structure.SuppressJsonDiff("", state, config, nil),
				suppressOIDCMetadataDiff("", state, config, nil),
				"state %q vs config %q should behave like structure.SuppressJsonDiff", state, config,
			)
		}
	}

	// Guards against the skip predicate widening until the test compares nothing.
	assert.Greater(t, compared, len(documents), "parity test skipped too many pairs to be meaningful")
}

// carriesFalseServerDefault reports the intended departure: a document holding one of the
// server-defaulted booleans as false, which suppressOIDCMetadataDiff ignores on either
// side. Checked per document rather than per pair, since the tolerance is symmetric.
func carriesFalseServerDefault(t *testing.T, document string) bool {
	t.Helper()

	doc, err := structure.ExpandJsonFromString(document)
	if err != nil {
		return false
	}

	for _, key := range oidcMetadataServerDefaults {
		if defaulted, isBool := doc[key].(bool); isBool && !defaulted {
			return true
		}
	}

	return false
}
