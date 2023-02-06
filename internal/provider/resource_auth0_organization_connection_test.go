package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccOrganizationConnectionGivenAnOrgAndAConnection = `
resource auth0_connection my_connection {
	name = "Acceptance-Test-Connection-First-{{.testName}}"
	strategy = "auth0"
}

resource auth0_organization my_organization {
	depends_on = [auth0_connection.my_connection]
	name = "test-{{.testName}}"
	display_name = "Acme Inc. {{.testName}}"
}
`

const TestAccOrganizationConnectionCreate = testAccOrganizationConnectionGivenAnOrgAndAConnection + `
resource auth0_organization_connection my_org_conn {
	organization_id = auth0_organization.my_organization.id
	connection_id = auth0_connection.my_connection.id
}
`

const TestAccOrganizationConnectionUpdate = testAccOrganizationConnectionGivenAnOrgAndAConnection + `
resource auth0_organization_connection my_org_conn {
	organization_id = auth0_organization.my_organization.id
	connection_id = auth0_connection.my_connection.id
	assign_membership_on_login = true
}
`

func TestAccOrganizationConnection(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: ProviderTestFactories(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(TestAccOrganizationConnectionCreate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"name",
						"Acceptance-Test-Connection-First-"+strings.ToLower(t.Name()),
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"strategy",
						"auth0",
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"assign_membership_on_login",
						"false",
					),
				),
			},
			{
				Config: template.ParseTestName(TestAccOrganizationConnectionUpdate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "organization_id"),
					resource.TestCheckResourceAttrSet("auth0_organization_connection.my_org_conn", "connection_id"),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"name",
						"Acceptance-Test-Connection-First-"+strings.ToLower(t.Name()),
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"strategy",
						"auth0",
					),
					resource.TestCheckResourceAttr(
						"auth0_organization_connection.my_org_conn",
						"assign_membership_on_login",
						"true",
					),
				),
			},
		},
	})
}

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
			data := schema.TestResourceDataRaw(t, newOrganizationConnection().Schema, nil)
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
