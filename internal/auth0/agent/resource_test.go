package agent_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const agentIDPattern = `^agt_[1-9A-HJ-NP-Za-km-z]{22}$`

const testAccAgentCreate = `
resource "auth0_agent" "my_agent" {
	name              = "Agent {{.testName}}"
	external_agent_id = "ext-{{.testName}}"
	metadata          = jsonencode({ team = "platform" })
}
`

const testAccAgentUpdate = `
resource "auth0_agent" "my_agent" {
	name              = "Agent updated {{.testName}}"
	external_agent_id = "ext-{{.testName}}"
	metadata          = jsonencode({ team = "security", tier = "gold" })
}
`

const testAccAgentMetadataRemoved = `
resource "auth0_agent" "my_agent" {
	name              = "Agent updated {{.testName}}"
	external_agent_id = "ext-{{.testName}}"
}
`

const testAccAgentDataSource = testAccAgentMetadataRemoved + `
data "auth0_agent" "my_agent" {
	agent_id = auth0_agent.my_agent.id
}
`

func TestAccAgent(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccAgentCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_agent.my_agent", "id", regexp.MustCompile(agentIDPattern)),
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "name", "Agent TestAccAgent"),
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "external_agent_id", "ext-TestAccAgent"),
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "metadata", `{"team":"platform"}`),
					resource.TestCheckResourceAttrSet("auth0_agent.my_agent", "created_at"),
					resource.TestCheckResourceAttrSet("auth0_agent.my_agent", "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccAgentUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "name", "Agent updated TestAccAgent"),
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "metadata", `{"team":"security","tier":"gold"}`),
				),
			},
			{
				Config: acctest.ParseTestName(testAccAgentMetadataRemoved, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_agent.my_agent", "metadata", ""),
				),
			},
			{
				ResourceName:      "auth0_agent.my_agent",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testAccAgentDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.auth0_agent.my_agent", "id", "auth0_agent.my_agent", "id"),
					resource.TestCheckResourceAttrPair("data.auth0_agent.my_agent", "name", "auth0_agent.my_agent", "name"),
					resource.TestCheckResourceAttr("data.auth0_agent.my_agent", "external_agent_id", "ext-TestAccAgent"),
				),
			},
		},
	})
}

func TestAccAgentImportInvalidID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccAgentCreate, t.Name()),
			},
			{
				ResourceName:  "auth0_agent.my_agent",
				ImportState:   true,
				ImportStateId: "not-an-agent-id",
				ExpectError:   regexp.MustCompile(`agent ID`),
			},
		},
	})
}
