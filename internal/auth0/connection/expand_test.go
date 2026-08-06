package connection

import (
	"encoding/json"
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
)

func TestCheckForUnmanagedConfigurationSecrets(t *testing.T) {
	var testCases = []struct {
		name                string
		givenConfigFromTF   map[string]string
		givenConfigFromAPI  map[string]string
		expectedDiagnostics diag.Diagnostics
	}{
		{
			name:                "custom database has no configuration",
			givenConfigFromTF:   map[string]string{},
			givenConfigFromAPI:  map[string]string{},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name: "custom database has no unmanaged configuration",
			givenConfigFromTF: map[string]string{
				"foo": "bar",
			},
			givenConfigFromAPI: map[string]string{
				"foo": "bar",
			},
			expectedDiagnostics: diag.Diagnostics(nil),
		},
		{
			name: "custom database has unmanaged configuration",
			givenConfigFromTF: map[string]string{
				"foo": "bar",
			},
			givenConfigFromAPI: map[string]string{
				"foo":        "bar",
				"anotherFoo": "anotherBar",
			},
			expectedDiagnostics: diag.Diagnostics{
				diag.Diagnostic{
					Severity:      diag.Error,
					Summary:       "Unmanaged Configuration Secret",
					Detail:        "Detected a configuration secret not managed through terraform: \"anotherFoo\". If you proceed, this configuration secret will get deleted. It is required to add this configuration secret to your custom database settings to prevent unintentionally destructive results.",
					AttributePath: cty.Path{cty.GetAttrStep{Name: "options.configuration"}},
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actualDiagnostics := checkForUnmanagedConfigurationSecrets(
				testCase.givenConfigFromTF,
				testCase.givenConfigFromAPI,
			)

			assert.Equal(t, testCase.expectedDiagnostics, actualDiagnostics)
		})
	}
}

// TestEchoEmailAttributeUnique guards the workaround for the Management API
// rejecting a PATCH that omits options.attributes.email.unique when the stored
// value is false.
func TestEchoEmailAttributeUnique(t *testing.T) {
	optionsWithEmailUnique := func(unique *bool) *management.ConnectionOptions {
		return &management.ConnectionOptions{
			Attributes: &management.ConnectionOptionsAttributes{
				Email: &management.ConnectionOptionsEmailAttribute{Unique: unique},
			},
		}
	}

	t.Run("echoes back a stored unique of false so the API accepts the patch", func(t *testing.T) {
		options := optionsWithEmailUnique(nil)

		echoEmailAttributeUnique(options, optionsWithEmailUnique(auth0.Bool(false)))

		assert.Equal(t, auth0.Bool(false), options.GetAttributes().GetEmail().Unique)
	})

	t.Run("treats a unique absent from the API response as true", func(t *testing.T) {
		options := optionsWithEmailUnique(nil)

		echoEmailAttributeUnique(options, optionsWithEmailUnique(nil))

		assert.Equal(t, auth0.Bool(true), options.GetAttributes().GetEmail().Unique)
	})

	t.Run("leaves unique untouched when the email attribute is being added", func(t *testing.T) {
		options := optionsWithEmailUnique(nil)

		echoEmailAttributeUnique(options, &management.ConnectionOptions{})

		assert.Nil(t, options.GetAttributes().GetEmail().Unique)
	})
}

func TestEmailAttributeUnique(t *testing.T) {
	assert.True(t, emailAttributeUnique(nil))
	assert.True(t, emailAttributeUnique(&management.ConnectionOptionsEmailAttribute{}))
	assert.False(t, emailAttributeUnique(&management.ConnectionOptionsEmailAttribute{Unique: auth0.Bool(false)}))
}

// TestConnectionOptionsTypeOmitsWhenNil guards against a regression where an unset
// connection `type` was serialized as `"type":null`, which the Auth0 API rejects
// with `"options.type" must be ...`. The field must be omitted entirely when nil,
// and only sent when explicitly configured. This affects both the Okta and OIDC
// strategies, which share the same `type` option.
func TestConnectionOptionsTypeOmitsWhenNil(t *testing.T) {
	t.Run("okta omits type when nil", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOkta{})
		assert.NoError(t, err)
		assert.NotContains(t, string(payload), "type")
	})

	t.Run("okta includes type when set", func(t *testing.T) {
		options := &management.ConnectionOptionsOkta{Type: auth0.String("back_channel")}
		payload, err := json.Marshal(options)
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"type":"back_channel"`)
	})

	t.Run("oidc omits type when nil", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOIDC{})
		assert.NoError(t, err)
		assert.NotContains(t, string(payload), "type")
	})

	t.Run("oidc includes type when set", func(t *testing.T) {
		options := &management.ConnectionOptionsOIDC{Type: auth0.String("back_channel")}
		payload, err := json.Marshal(options)
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"type":"back_channel"`)
	})
}
