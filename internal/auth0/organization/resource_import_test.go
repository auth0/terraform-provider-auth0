package organization

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestImportOrganizationConnection(t *testing.T) {
	var testCases = []struct {
		testName               string
		givenID                string
		expectedOrganizationID string
		expectedConnectionID   string
		expectedError          error
	}{
		{
			testName:               "it correctly parses the resource ID",
			givenID:                "org_1234:conn_5678",
			expectedOrganizationID: "org_1234",
			expectedConnectionID:   "conn_5678",
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:      "it fails when the given ID does not have \":\" as a separator",
			givenID:       "org_1234conn_5678",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<connectionID>"),
		},
		{
			testName:      "it fails when the given ID has too many separators",
			givenID:       "org_1234:conn_5678:",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<connectionID>"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, NewConnectionResource().Schema, nil)
			data.SetId(testCase.givenID)

			actualData, err := importOrganizationConnection(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Get("organization_id").(string), testCase.expectedOrganizationID)
			assert.Equal(t, actualData[0].Get("connection_id").(string), testCase.expectedConnectionID)
			assert.NotEqual(t, actualData[0].Id(), testCase.givenID)
		})
	}
}

func TestImportOrganizationMember(t *testing.T) {
	var testCases = []struct {
		testName               string
		givenID                string
		expectedOrganizationID string
		expectedUserID         string
		expectedError          error
	}{
		{
			testName:               "it correctly parses the resource ID",
			givenID:                "org_1234:auth0|62d82",
			expectedOrganizationID: "org_1234",
			expectedUserID:         "auth0|62d82",
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:      "it fails when the given ID does not have \":\" as a separator",
			givenID:       "org_1234auth0|62d82",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<userID>"),
		},
		{
			testName:      "it fails when the given ID has too many separators",
			givenID:       "org_1234:auth0|62d82:",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<userID>"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, NewMemberResource().Schema, nil)
			data.SetId(testCase.givenID)

			actualData, err := importOrganizationMember(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Get("organization_id").(string), testCase.expectedOrganizationID)
			assert.Equal(t, actualData[0].Get("user_id").(string), testCase.expectedUserID)
			assert.NotEqual(t, actualData[0].Id(), testCase.givenID)
		})
	}
}
