package schema

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestImportResourceGroupID(t *testing.T) {
	var testCases = []struct {
		testName           string
		givenID            string
		givenSeparator     string
		expectedAttributes map[int]map[string]string
		expectedError      error
	}{
		{
			testName:       "it correctly parses the resource ID (org:conn)",
			givenID:        "org_1234:conn_5678",
			givenSeparator: ":",
			expectedAttributes: map[int]map[string]string{
				1: {"organization_id": "org_1234"},
				2: {"connection_id": "conn_5678"},
			},
		},
		{
			testName:       "it correctly parses the resource ID (org:user)",
			givenID:        "org_1234:auth0|62d82",
			givenSeparator: ":",
			expectedAttributes: map[int]map[string]string{
				1: {"organization_id": "org_1234"},
				2: {"user_id": "auth0|62d82"},
			},
		},
		{
			testName:       "it correctly parses the resource ID (conn:client)",
			givenID:        "conn_5678::client_1234",
			givenSeparator: "::",
			expectedAttributes: map[int]map[string]string{
				1: {"connection_id": "conn_5678"},
				2: {"client_id": "client_1234"},
			},
		},
		{
			testName:       "it correctly parses the resource ID (org:member:role)",
			givenID:        "org_5678:user_1234:role_5341",
			givenSeparator: ":",
			expectedAttributes: map[int]map[string]string{
				1: {"organization_id": "org_5678"},
				2: {"user_id": "user_1234"},
				3: {"role_id": "role_5341"},
			},
		},
		{
			testName:       "it correctly parses the resource ID (user:api:permission)",
			givenID:        "user_5678::https://api::read:books",
			givenSeparator: "::",
			expectedAttributes: map[int]map[string]string{
				1: {"user_id": "user_5678"},
				2: {"resource_server_identifier": "https://api"},
				3: {"permission": "read:books"},
			},
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:       "it fails when the given ID does not have \":\" as a separator",
			givenID:        "org_1234conn_5678",
			givenSeparator: ":",
			expectedAttributes: map[int]map[string]string{
				1: {"organization_id": ""},
				2: {"connection_id": ""},
			},
			expectedError: fmt.Errorf("ID must be formatted as <organization_id>:<connection_id>"),
		},
		{
			testName:       "it fails when the given ID has too many separators",
			givenID:        "org_1234:conn_5678:",
			givenSeparator: ":",
			expectedAttributes: map[int]map[string]string{
				1: {"organization_id": ""},
				2: {"connection_id": ""},
			},
			expectedError: fmt.Errorf("ID must be formatted as <organization_id>:<connection_id>"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			testSchema := make(map[string]*schema.Schema, 0)
			resourceKeys := make([]string, 0)

			for index := 1; index <= len(testCase.expectedAttributes); index++ {
				for key := range testCase.expectedAttributes[index] {
					resourceKeys = append(resourceKeys, key)
					testSchema[key] = &schema.Schema{
						Type:     schema.TypeString,
						Required: true,
					}
				}
			}

			data := schema.TestResourceDataRaw(t, testSchema, nil)
			data.SetId(testCase.givenID)

			importFunc := ImportResourceGroupID(testCase.givenSeparator, resourceKeys...)
			actualData, err := importFunc(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Id(), testCase.givenID)

			for index, key := range resourceKeys {
				assert.Equal(t, actualData[0].Get(key).(string), testCase.expectedAttributes[index+1][key])
			}
		})
	}
}
