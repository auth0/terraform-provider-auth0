package connection

import (
	"testing"

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
