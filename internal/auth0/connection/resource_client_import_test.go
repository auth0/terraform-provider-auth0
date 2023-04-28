package connection

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestImportConnectionClient(t *testing.T) {
	var testCases = []struct {
		testName             string
		givenID              string
		expectedConnectionID string
		expectedClientID     string
		expectedError        error
	}{
		{
			testName:             "it correctly parses the resource ID",
			givenID:              "conn_5678:client_1234",
			expectedConnectionID: "conn_5678",
			expectedClientID:     "client_1234",
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:      "it fails when the given ID does not have \":\" as a separator",
			givenID:       "client_1234conn_5678",
			expectedError: fmt.Errorf("ID must be formated as <connectionID>:<clientID>"),
		},
		{
			testName:      "it fails when the given ID has too many separators",
			givenID:       "client_1234:conn_5678:",
			expectedError: fmt.Errorf("ID must be formated as <connectionID>:<clientID>"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, NewClientResource().Schema, nil)
			data.SetId(testCase.givenID)

			actualData, err := importConnectionClient(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Get("connection_id").(string), testCase.expectedConnectionID)
			assert.Equal(t, actualData[0].Get("client_id").(string), testCase.expectedClientID)
			assert.NotEqual(t, actualData[0].Id(), testCase.givenID)
		})
	}
}
