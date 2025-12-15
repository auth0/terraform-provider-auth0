package eventstream

import (
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Tests for getTokenWO() are not included here because it relies on GetRawConfig()
// which is not populated by TestResourceDataRaw in unit tests. This function is tested
// through integration tests where the full Terraform lifecycle is available.

// Note: hasTokenWOVersionChanged() is tested through integration tests where HasChange()
// works properly with the full Terraform lifecycle. Unit testing this function is limited
// because TestResourceDataRaw doesn't properly simulate HasChange() behavior.

func TestExpandEventStreamDestination_Webhook(t *testing.T) {
	testCases := []struct {
		name                    string
		configData              map[string]interface{}
		isNewResource           bool
		expectedDestinationType string
		expectedEndpoint        string
		expectedAuthMethod      string
		shouldHaveToken         bool
		shouldHaveBasicAuth     bool
	}{
		{
			name: "webhook with bearer token",
			configData: map[string]interface{}{
				"name":             "test-stream",
				"destination_type": "webhook",
				"subscriptions":    []interface{}{"user.created"},
				"webhook_configuration": []interface{}{
					map[string]interface{}{
						"webhook_endpoint": "https://example.com/webhook",
						"webhook_authorization": []interface{}{
							map[string]interface{}{
								"method": "bearer",
								"token":  "test-token-123",
							},
						},
					},
				},
			},
			isNewResource:           true,
			expectedDestinationType: "webhook",
			expectedEndpoint:        "https://example.com/webhook",
			expectedAuthMethod:      "bearer",
			shouldHaveToken:         true,
			shouldHaveBasicAuth:     false,
		},
		{
			name: "webhook with basic authentication",
			configData: map[string]interface{}{
				"name":             "test-stream",
				"destination_type": "webhook",
				"subscriptions":    []interface{}{"user.created"},
				"webhook_configuration": []interface{}{
					map[string]interface{}{
						"webhook_endpoint": "https://example.com/webhook",
						"webhook_authorization": []interface{}{
							map[string]interface{}{
								"method":   "basic",
								"username": "testuser",
								"password": "testpass",
							},
						},
					},
				},
			},
			isNewResource:           true,
			expectedDestinationType: "webhook",
			expectedEndpoint:        "https://example.com/webhook",
			expectedAuthMethod:      "basic",
			shouldHaveToken:         false,
			shouldHaveBasicAuth:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resourceSchema := NewResource().Schema
			data := schema.TestResourceDataRaw(t, resourceSchema, tc.configData)

			if !tc.isNewResource {
				data.SetId("existing-id")
			}

			destination := expandEventStreamDestination(data)

			require.NotNil(t, destination)
			assert.Equal(t, tc.expectedDestinationType, *destination.EventStreamDestinationType)

			config := destination.EventStreamDestinationConfiguration
			require.NotNil(t, config)
			assert.Equal(t, tc.expectedEndpoint, config["webhook_endpoint"])

			authConfig, ok := config["webhook_authorization"].(map[string]interface{})
			require.True(t, ok, "webhook_authorization should be present in config")
			assert.Equal(t, tc.expectedAuthMethod, authConfig["method"])

			if tc.shouldHaveToken {
				_, hasToken := authConfig["token"]
				assert.True(t, hasToken, "should have token for bearer auth")
			}

			if tc.shouldHaveBasicAuth {
				_, hasUsername := authConfig["username"]
				_, hasPassword := authConfig["password"]
				assert.True(t, hasUsername, "should have username for basic auth")
				assert.True(t, hasPassword, "should have password for basic auth")
			}
		})
	}
}

// Note: Tests for expandEventStreamDestination with EventBridge are not included here because
// the function's behavior depends on IsNewResource() which doesn't work properly with
// TestResourceDataRaw in unit tests. This function is tested through integration tests where
// the full Terraform lifecycle is available.

func TestExpandEventStreamSubscriptions(t *testing.T) {
	testCases := []struct {
		name             string
		inputValue       cty.Value
		expectedSubCount int
		expectedSubTypes []string
	}{
		{
			name: "expands multiple subscriptions",
			inputValue: cty.SetVal([]cty.Value{
				cty.StringVal("user.created"),
				cty.StringVal("user.updated"),
				cty.StringVal("user.deleted"),
			}),
			expectedSubCount: 3,
			expectedSubTypes: []string{"user.created", "user.updated", "user.deleted"},
		},
		{
			name: "expands single subscription",
			inputValue: cty.SetVal([]cty.Value{
				cty.StringVal("user.created"),
			}),
			expectedSubCount: 1,
			expectedSubTypes: []string{"user.created"},
		},
		{
			name:             "expands empty subscriptions",
			inputValue:       cty.SetValEmpty(cty.String),
			expectedSubCount: 0,
			expectedSubTypes: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := expandEventStreamSubscriptions(tc.inputValue)

			require.NotNil(t, result)
			assert.Len(t, *result, tc.expectedSubCount)

			// Collect actual subscription types.
			actualTypes := make([]string, 0, len(*result))
			for _, sub := range *result {
				actualTypes = append(actualTypes, *sub.EventStreamSubscriptionType)
			}

			// Check that all expected types are present (order doesn't matter for sets).
			for _, expectedType := range tc.expectedSubTypes {
				assert.Contains(t, actualTypes, expectedType)
			}
		})
	}
}

// Note: Tests for expandEventStream() are not included here because it relies on GetRawConfig()
// which is not populated by TestResourceDataRaw in unit tests. This function is tested
// through integration tests where the full Terraform lifecycle is available.
