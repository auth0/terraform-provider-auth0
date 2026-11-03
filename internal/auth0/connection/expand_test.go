package connection

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("keeps a configured false when the email attribute is being added", func(t *testing.T) {
		options := optionsWithEmailUnique(auth0.Bool(false))

		echoEmailAttributeUnique(options, &management.ConnectionOptions{})

		assert.Equal(t, auth0.Bool(false), options.GetAttributes().GetEmail().Unique)
	})

	t.Run("overwrites a configured value that disagrees with the stored one", func(t *testing.T) {
		options := optionsWithEmailUnique(auth0.Bool(false))

		echoEmailAttributeUnique(options, optionsWithEmailUnique(auth0.Bool(true)))

		assert.Equal(t, auth0.Bool(true), options.GetAttributes().GetEmail().Unique)
	})
}

// TestExpandConnectionOptionsEmailAttributeUnique asserts that unique is expanded
// from the configuration on every request, not only on create. The API rejects a
// PATCH that omits it when the stored value is false, and an email attribute added
// on update has no stored value for echoEmailAttributeUnique to supply.
func TestExpandConnectionOptionsEmailAttributeUnique(t *testing.T) {
	attributesConfig := func(unique cty.Value) cty.Value {
		return cty.ObjectVal(map[string]cty.Value{
			"email": cty.ListVal([]cty.Value{
				cty.ObjectVal(map[string]cty.Value{
					"identifier": cty.ListValEmpty(cty.Object(map[string]cty.Type{
						"active":         cty.Bool,
						"default_method": cty.String,
					})),
					"profile_required":    cty.True,
					"verification_method": cty.StringVal("link"),
					"signup": cty.ListValEmpty(cty.Object(map[string]cty.Type{
						"status":       cty.String,
						"verification": cty.List(cty.Object(map[string]cty.Type{"active": cty.Bool})),
					})),
					"unique": unique,
				}),
			}),
		})
	}

	t.Run("expands a configured false", func(t *testing.T) {
		emailAttribute := expandConnectionOptionsEmailAttribute(attributesConfig(cty.False))

		assert.Equal(t, auth0.Bool(false), emailAttribute.Unique)
	})

	t.Run("expands a configured true", func(t *testing.T) {
		emailAttribute := expandConnectionOptionsEmailAttribute(attributesConfig(cty.True))

		assert.Equal(t, auth0.Bool(true), emailAttribute.Unique)
	})

	t.Run("omits unique when the configuration leaves it unset", func(t *testing.T) {
		emailAttribute := expandConnectionOptionsEmailAttribute(attributesConfig(cty.NullVal(cty.Bool)))

		assert.Nil(t, emailAttribute.Unique)
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

// The connection PATCH replaces the whole options object, and the API does not
// materialize otp_settings defaults on read. So the request must omit otp_settings
// (and any unset channel or sub-field) entirely rather than send zero values.
func TestConnectionOptionsOTPSettingsOmitsWhenNil(t *testing.T) {
	t.Run("omits otp_settings when nil", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptions{})
		assert.NoError(t, err)
		assert.NotContains(t, string(payload), "otp_settings")
	})

	t.Run("omits the phone channel when only email is set", func(t *testing.T) {
		options := &management.ConnectionOptions{
			OTPSettings: &management.ConnectionOptionsOTPSettings{
				Email: &management.ConnectionOptionsOTPChannelSettings{
					OTPLength: auth0.Int(8),
					OTPExpiry: auth0.Int(600),
				},
			},
		}
		payload, err := json.Marshal(options)
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"otp_settings":{"email":{"otp_length":8,"otp_expiry":600}}`)
		assert.NotContains(t, string(payload), "phone")
	})

	t.Run("omits an unset sub-field within a channel", func(t *testing.T) {
		options := &management.ConnectionOptions{
			OTPSettings: &management.ConnectionOptionsOTPSettings{
				Email: &management.ConnectionOptionsOTPChannelSettings{
					OTPLength: auth0.Int(7),
				},
			},
		}
		payload, err := json.Marshal(options)
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"otp_settings":{"email":{"otp_length":7}}`)
		assert.NotContains(t, string(payload), "otp_expiry")
	})

	t.Run("includes both channels when set", func(t *testing.T) {
		options := &management.ConnectionOptions{
			OTPSettings: &management.ConnectionOptionsOTPSettings{
				Email: &management.ConnectionOptionsOTPChannelSettings{OTPLength: auth0.Int(8), OTPExpiry: auth0.Int(600)},
				Phone: &management.ConnectionOptionsOTPChannelSettings{OTPLength: auth0.Int(6), OTPExpiry: auth0.Int(900)},
			},
		}
		payload, err := json.Marshal(options)
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"email":{"otp_length":8,"otp_expiry":600}`)
		assert.Contains(t, string(payload), `"phone":{"otp_length":6,"otp_expiry":900}`)
	})
}

// The API rejects `oidc_metadata: null` with `"metadata_response" must be of type object`,
// so an unset attribute has to be omitted from the request rather than nulled.
func TestConnectionOptionsOIDCMetadataOmitsWhenNil(t *testing.T) {
	metadata := map[string]interface{}{"issuer": "https://idp.example.com"}

	t.Run("oidc omits oidc_metadata when nil", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOIDC{})
		assert.NoError(t, err)
		assert.NotContains(t, string(payload), "oidc_metadata")
	})

	t.Run("oidc includes oidc_metadata when set", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOIDC{OIDCMetadata: metadata})
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"oidc_metadata":{"issuer":"https://idp.example.com"}`)
	})

	t.Run("okta omits oidc_metadata when nil", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOkta{})
		assert.NoError(t, err)
		assert.NotContains(t, string(payload), "oidc_metadata")
	})

	t.Run("okta includes oidc_metadata when set", func(t *testing.T) {
		payload, err := json.Marshal(&management.ConnectionOptionsOkta{OIDCMetadata: metadata})
		assert.NoError(t, err)
		assert.Contains(t, string(payload), `"oidc_metadata":{"issuer":"https://idp.example.com"}`)
	})
}

func TestExpandConnectionOptionsSAMLAssertionDecryptionSettings(t *testing.T) {
	settingsType := cty.Object(map[string]cty.Type{
		"algorithm_profile":    cty.String,
		"algorithm_exceptions": cty.List(cty.String),
	})
	givenConfig := func(settings ...cty.Value) cty.Value {
		list := cty.ListValEmpty(settingsType)
		if len(settings) > 0 {
			list = cty.ListVal(settings)
		}

		return cty.ObjectVal(map[string]cty.Value{"assertion_decryption_settings": list})
	}

	var testCases = []struct {
		name     string
		config   cty.Value
		expected *management.ConnectionOptionsSAMLAssertionDecryptionSettings
	}{
		{
			name:     "block not in config gives nil",
			config:   givenConfig(),
			expected: nil,
		},
		{
			name: "algorithm_exceptions not in config is omitted",
			config: givenConfig(cty.ObjectVal(map[string]cty.Value{
				"algorithm_profile":    cty.StringVal("v2026-1"),
				"algorithm_exceptions": cty.NullVal(cty.List(cty.String)),
			})),
			expected: &management.ConnectionOptionsSAMLAssertionDecryptionSettings{
				AlgorithmProfile: auth0.String("v2026-1"),
			},
		},
		{
			name: "empty algorithm_exceptions is sent as empty",
			config: givenConfig(cty.ObjectVal(map[string]cty.Value{
				"algorithm_profile":    cty.StringVal("v2026-1"),
				"algorithm_exceptions": cty.ListValEmpty(cty.String),
			})),
			expected: &management.ConnectionOptionsSAMLAssertionDecryptionSettings{
				AlgorithmProfile:    auth0.String("v2026-1"),
				AlgorithmExceptions: &[]string{},
			},
		},
		{
			name: "algorithm_exceptions keep their order",
			config: givenConfig(cty.ObjectVal(map[string]cty.Value{
				"algorithm_profile": cty.StringVal("v2026-1"),
				"algorithm_exceptions": cty.ListVal([]cty.Value{
					cty.StringVal("http://www.w3.org/2001/04/xmlenc#tripledes-cbc"),
					cty.StringVal("http://www.w3.org/2001/04/xmlenc#aes128-cbc"),
				}),
			})),
			expected: &management.ConnectionOptionsSAMLAssertionDecryptionSettings{
				AlgorithmProfile: auth0.String("v2026-1"),
				AlgorithmExceptions: &[]string{
					"http://www.w3.org/2001/04/xmlenc#tripledes-cbc",
					"http://www.w3.org/2001/04/xmlenc#aes128-cbc",
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			actual := expandConnectionOptionsSAMLAssertionDecryptionSettings(testCase.config)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

// The API removes assertion_decryption_settings when a PATCH sends options without it,
// so the pass-through must keep the stored value when the configuration does not set it,
// and must not add the field to a connection that never had it.
func TestPassThroughConnectionOptionsAssertionDecryptionSettings(t *testing.T) {
	stored := &management.ConnectionOptionsSAMLAssertionDecryptionSettings{
		AlgorithmProfile:    auth0.String("v2026-1"),
		AlgorithmExceptions: &[]string{"http://www.w3.org/2001/04/xmlenc#tripledes-cbc"},
	}
	configured := &management.ConnectionOptionsSAMLAssertionDecryptionSettings{
		AlgorithmProfile:    auth0.String("v2026-1"),
		AlgorithmExceptions: &[]string{},
	}
	storedJSON := `"assertion_decryption_settings":{"algorithm_profile":"v2026-1",` +
		`"algorithm_exceptions":["http://www.w3.org/2001/04/xmlenc#tripledes-cbc"]}`

	givenAPI := func(t *testing.T, strategy string, withSettings bool) *management.Management {
		t.Helper()

		options := `{}`
		if withSettings {
			options = `{` + storedJSON + `}`
		}

		server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != "/api/v2/connections/con_123" {
				http.NotFound(writer, request)
				return
			}

			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"id":"con_123","strategy":"` + strategy + `","options":` + options + `}`))
		}))
		t.Cleanup(server.Close)

		api, err := management.New(server.URL, management.WithInsecure())
		require.NoError(t, err)

		return api
	}

	passThroughFuncs := map[string]func(*management.Connection) *management.ConnectionOptionsSAMLAssertionDecryptionSettings{
		management.ConnectionStrategySAML: func(connection *management.Connection) *management.ConnectionOptionsSAMLAssertionDecryptionSettings {
			return connection.Options.(*management.ConnectionOptionsSAML).AssertionDecryptionSettings
		},
		management.ConnectionStrategyPingFederate: func(connection *management.Connection) *management.ConnectionOptionsSAMLAssertionDecryptionSettings {
			return connection.Options.(*management.ConnectionOptionsPingFederate).AssertionDecryptionSettings
		},
	}
	givenConnection := func(strategy string, settings *management.ConnectionOptionsSAMLAssertionDecryptionSettings) *management.Connection {
		if strategy == management.ConnectionStrategySAML {
			return &management.Connection{Options: &management.ConnectionOptionsSAML{AssertionDecryptionSettings: settings}}
		}

		return &management.Connection{Options: &management.ConnectionOptionsPingFederate{AssertionDecryptionSettings: settings}}
	}

	for strategy, settingsOf := range passThroughFuncs {
		t.Run(strategy+" keeps the stored value when the config does not set it", func(t *testing.T) {
			connection := givenConnection(strategy, nil)

			err := passThroughUnconfigurableConnectionOptions(context.Background(), givenAPI(t, strategy, true), "con_123", strategy, connection)
			require.NoError(t, err)

			assert.Equal(t, stored, settingsOf(connection))

			payload, err := json.Marshal(connection.Options)
			require.NoError(t, err)
			assert.Contains(t, string(payload), storedJSON)
		})

		t.Run(strategy+" does not add the field when the connection does not have it", func(t *testing.T) {
			connection := givenConnection(strategy, nil)

			err := passThroughUnconfigurableConnectionOptions(context.Background(), givenAPI(t, strategy, false), "con_123", strategy, connection)
			require.NoError(t, err)

			assert.Nil(t, settingsOf(connection))

			payload, err := json.Marshal(connection.Options)
			require.NoError(t, err)
			assert.NotContains(t, string(payload), "assertion_decryption_settings")
		})

		t.Run(strategy+" sends the configured value instead of the stored value", func(t *testing.T) {
			connection := givenConnection(strategy, configured)

			err := passThroughUnconfigurableConnectionOptions(context.Background(), givenAPI(t, strategy, true), "con_123", strategy, connection)
			require.NoError(t, err)

			assert.Equal(t, configured, settingsOf(connection))
		})
	}
}
