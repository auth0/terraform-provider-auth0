package connection

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateConnection covers the plan time guard on
// options.attributes.email.unique. The guard can hard fail a plan, so both the
// changes it must reject and the ones it must let through are asserted here.
func TestValidateConnection(t *testing.T) {
	const connectionID = "con_testConnection123"

	stateWithStoredUnique := func(unique string) map[string]string {
		return map[string]string{
			"id":                                    connectionID,
			"name":                                  "my-connection",
			"strategy":                              "auth0",
			"options.#":                             "1",
			"options.0.attributes.#":                "1",
			"options.0.attributes.0.email.#":        "1",
			"options.0.attributes.0.email.0.unique": unique,
		}
	}

	configWithUnique := func(name string, unique interface{}) map[string]interface{} {
		emailConfig := map[string]interface{}{"profile_required": true}
		if unique != nil {
			emailConfig["unique"] = unique
		}

		return map[string]interface{}{
			"name":     name,
			"strategy": "auth0",
			"options": []interface{}{
				map[string]interface{}{
					"attributes": []interface{}{
						map[string]interface{}{
							"email": []interface{}{emailConfig},
						},
					},
				},
			},
		}
	}

	testCases := []struct {
		name          string
		state         map[string]string
		config        map[string]interface{}
		expectedError string
	}{
		{
			name:   "allows any value on create",
			state:  nil,
			config: configWithUnique("my-connection", true),
		},
		{
			name:   "allows a stored false to stay false",
			state:  stateWithStoredUnique("false"),
			config: configWithUnique("my-connection", false),
		},
		{
			name:   "allows a configuration that omits unique to adopt the stored false",
			state:  stateWithStoredUnique("false"),
			config: configWithUnique("my-connection", nil),
		},
		{
			name:   "allows a stored true to stay true",
			state:  stateWithStoredUnique("true"),
			config: configWithUnique("my-connection", true),
		},
		{
			name: "allows adding an email attribute absent from state",
			state: map[string]string{
				"id":       connectionID,
				"name":     "my-connection",
				"strategy": "auth0",
			},
			config: configWithUnique("my-connection", true),
		},
		{
			name: "allows adding an email attribute to an empty attributes block",
			state: map[string]string{
				"id":                             connectionID,
				"name":                           "my-connection",
				"strategy":                       "auth0",
				"options.#":                      "1",
				"options.0.attributes.#":         "1",
				"options.0.attributes.0.email.#": "0",
			},
			config: configWithUnique("my-connection", true),
		},
		{
			name:          "rejects flipping a stored false to true",
			state:         stateWithStoredUnique("false"),
			config:        configWithUnique("my-connection", true),
			expectedError: "options.attributes.email.unique cannot be changed after the connection is created",
		},
		{
			name:          "rejects flipping a stored true to false",
			state:         stateWithStoredUnique("true"),
			config:        configWithUnique("my-connection", false),
			expectedError: "options.attributes.email.unique cannot be changed after the connection is created",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var instanceState *terraform.InstanceState
			if testCase.state != nil {
				instanceState = &terraform.InstanceState{ID: connectionID, Attributes: testCase.state}
			}

			_, err := NewResource().Diff(
				context.Background(),
				instanceState,
				terraform.NewResourceConfigRaw(testCase.config),
				nil,
			)

			if testCase.expectedError == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), testCase.expectedError)
		})
	}
}
