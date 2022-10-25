package provider

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccCreateConnectionClient = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client" {
	name = "Acceptance-Test-Client-{{.testName}}"
}

resource "auth0_connection_client" "my_conn_client_assoc" {
	connection_id = auth0_connection.my_conn.id
	client_id     = auth0_client.my_client.id
}
`

const testAccDeleteConnectionClient = `
resource "auth0_connection" "my_conn" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	strategy = "auth0"
}

resource "auth0_client" "my_client" {
	name = "Acceptance-Test-Client-{{.testName}}"
}
`

func TestAccConnectionClient(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccCreateConnectionClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance-Test-Client-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "connection_id"),
					resource.TestCheckResourceAttrSet("auth0_connection_client.my_conn_client_assoc", "client_id"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_client.my_conn_client_assoc", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "enabled_clients.#", "1"),
				),
			},
			{
				Config: template.ParseTestName(testAccDeleteConnectionClient, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_client.my_client", "name", fmt.Sprintf("Acceptance-Test-Client-%s", t.Name())),
				),
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection.my_conn", "enabled_clients.#", "0"),
				),
			},
		},
	})
}

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
			data := schema.TestResourceDataRaw(t, newConnectionClient().Schema, nil)
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
