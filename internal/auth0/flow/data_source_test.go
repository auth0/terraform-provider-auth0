package flow_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAGivenFlow = `
resource "auth0_flow" "my_flow" {
    name = "test-flow-{{.testName}}"
}
`

const testDataResourceWithoutID = testAGivenFlow + `
data "auth0_flow" "my_flow" {
	depends_on = [resource.auth0_flow.my_flow]
}`

const testDataResourceWithInvalidID = testAGivenFlow + `
data "auth0_flow" "my_flow" {
	depends_on = [resource.auth0_flow.my_flow]
	id = "af_bskks8aGbiq7qS13umnuvX"
}`

const testDataResourceWithValidID = testAGivenFlow + `
data "auth0_flow" "my_flow" {
	depends_on = [resource.auth0_flow.my_flow]
	id = resource.auth0_flow.my_flow.id
}`

func TestAccFlowDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testDataResourceWithoutID, t.Name()),
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: 404 Not Found`,
				),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_flow.my_flow", "name"),
				),
			},
		},
	})
}
