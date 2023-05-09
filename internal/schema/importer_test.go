package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestImportResourcePairID(t *testing.T) {
	var testCases = []struct {
		testName         string
		givenID          string
		expectedFirstID  string
		expectedSecondID string
		expectedError    error
	}{
		{
			testName:         "it correctly parses the resource ID (org:conn)",
			givenID:          "org_1234:conn_5678",
			expectedFirstID:  "org_1234",
			expectedSecondID: "conn_5678",
		},
		{
			testName:         "it correctly parses the resource ID (org:user)",
			givenID:          "org_1234:auth0|62d82",
			expectedFirstID:  "org_1234",
			expectedSecondID: "auth0|62d82",
		},
		{
			testName:         "it correctly parses the resource ID (conn:client)",
			givenID:          "conn_5678:client_1234",
			expectedFirstID:  "conn_5678",
			expectedSecondID: "client_1234",
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:      "it fails when the given ID does not have \":\" as a separator",
			givenID:       "org_1234conn_5678",
			expectedError: fmt.Errorf("ID must be formatted as <resource_a_id>:<resource_b_id>"),
		},
		{
			testName:      "it fails when the given ID has too many separators",
			givenID:       "org_1234:conn_5678:",
			expectedError: fmt.Errorf("ID must be formatted as <resource_a_id>:<resource_b_id>"),
		},
	}

	testSchema := map[string]*schema.Schema{
		"resource_a_id": {
			Type:     schema.TypeString,
			Required: true,
		},
		"resource_b_id": {
			Type:     schema.TypeString,
			Required: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, testSchema, nil)
			data.SetId(testCase.givenID)

			importFunc := ImportResourcePairID("resource_a_id", "resource_b_id")
			actualData, err := importFunc(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Get("resource_a_id").(string), testCase.expectedFirstID)
			assert.Equal(t, actualData[0].Get("resource_b_id").(string), testCase.expectedSecondID)
			assert.Equal(t, actualData[0].Id(), testCase.givenID)
		})
	}
}
