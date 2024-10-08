package flow_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testFlowCreateInvalidAction = `
resource "auth0_flow" "my_flow" {
    name = "updated-test-flow-{{.testName}}"
	actions = "invalid-json"
}
`

const testFlowCreate = `
resource "auth0_flow" "my_flow" {
    name = "test-flow-{{.testName}}"

}
`

const testFlowUpdate = `
resource "auth0_flow" "my_flow" {
    name = "updated-test-flow-{{.testName}}"
}
`

const testFlowDelete = `
resource "auth0_flow" "my_flow" {
    name = "updated-test-flow-{{.testName}}"
}
`

func TestAccFlow(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testFlowCreateInvalidAction, t.Name()),
				ExpectError: regexp.MustCompile(`contains an invalid JSON`),
			},
			{
				Config: acctest.ParseTestName(testFlowCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_flow.my_flow", "name", fmt.Sprintf("test-flow-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_flow.my_flow", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testFlowUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_flow.my_flow", "name", fmt.Sprintf("updated-test-flow-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_flow.my_flow", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testFlowDelete, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_flow.my_flow", "name", fmt.Sprintf("updated-test-flow-%s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_flow.my_flow", "id"),
				),
				Destroy: true,
			},
		},
	})
}
