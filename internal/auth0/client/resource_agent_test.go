package client_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientWithAgentTwoAgents = `
resource "auth0_agent" "first" {
	name              = "First {{.testName}}"
	external_agent_id = "first-{{.testName}}"
}

resource "auth0_agent" "second" {
	name              = "Second {{.testName}}"
	external_agent_id = "second-{{.testName}}"
}
`

const testAccClientWithAgentUnlinked = testAccClientWithAgentTwoAgents + `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Agent - {{.testName}}"
	app_type = "non_interactive"
}
`

const testAccClientWithAgentLinkedToFirst = testAccClientWithAgentTwoAgents + `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Agent - {{.testName}}"
	app_type = "non_interactive"
	agent_id = auth0_agent.first.id
}
`

const testAccClientWithAgentLinkedToSecond = testAccClientWithAgentTwoAgents + `
resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Agent - {{.testName}}"
	app_type = "non_interactive"
	agent_id = auth0_agent.second.id
}
`

const testAccClientWithAgentReplaced = `
resource "auth0_agent" "first" {
	name              = "First {{.testName}}"
	external_agent_id = "first-replaced-{{.testName}}"
}

resource "auth0_agent" "second" {
	name              = "Second {{.testName}}"
	external_agent_id = "second-{{.testName}}"
}

resource "auth0_client" "my_client" {
	name     = "Acceptance Test - Agent - {{.testName}}"
	app_type = "non_interactive"
	agent_id = auth0_agent.first.id
}
`

func TestAccClientWithAgent(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccClientWithAgentUnlinked, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "agent_id", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientWithAgentLinkedToFirst, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_client.my_client", "agent_id", "auth0_agent.first", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientWithAgentLinkedToSecond, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_client.my_client", "agent_id", "auth0_agent.second", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientWithAgentUnlinked, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client.my_client", "agent_id", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientWithAgentLinkedToFirst, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_client.my_client", "agent_id", "auth0_agent.first", "id"),
				),
			},
			{
				// Changing external_agent_id forces the agent to be replaced, and the client follows the new ID.
				Config: acctest.ParseTestName(testAccClientWithAgentReplaced, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_client.my_client", "agent_id", "auth0_agent.first", "id"),
				),
			},
		},
	})
}
