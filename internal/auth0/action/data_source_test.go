package action_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAnAction = `resource "auth0_action" "my_action" {
	name = "Test Action {{.testName}}"
	code = "exports.onExecutePostLogin = async (event, api) => {};"
	runtime = "node22"
	supported_triggers {
		id      = "post-login"
		version = "v3"
	}
}`

const testAccDataSourceActionWithActionID = givenAnAction + `
data "auth0_action" "my_action" {
    id = auth0_action.my_action.id
}
`

const testAccDataSourceActionWithActionName = givenAnAction + `
data "auth0_action" "my_action" {
	name = "Test Action {{.testName}}"
}
`

const testAccDataSourceActionWithInvalidActionName = givenAnAction + `
data "auth0_action" "my_action" {
	name = "random-name"
}
`

const testAccDataSourceActionWithInvalidActionID = givenAnAction + `
data "auth0_action" "my_action" {
	id = "09ee2b9b-282d-4a65-87d1-f01adf007745"
}
`

const testAccDataSourceActionWithNoParam = givenAnAction + `
data "auth0_action" "my_action" {}
`

func TestAccDataSourceAction(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccDataSourceActionWithInvalidActionName, t.Name()),
				ExpectError: regexp.MustCompile(`Error: No action found with "name" = "random-name"`),
			},
			{
				Config:      acctest.ParseTestName(testAccDataSourceActionWithInvalidActionID, t.Name()),
				ExpectError: regexp.MustCompile(`Error: 404 Not Found: That action does not exist.`),
			},
			{
				Config:      acctest.ParseTestName(testAccDataSourceActionWithNoParam, t.Name()),
				ExpectError: regexp.MustCompile(`Error: Missing required argument`),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceActionWithActionID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_action.my_action", "id"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "supported_triggers.0.version", "v3"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "runtime", "node22"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceActionWithActionName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_action.my_action", "id"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "supported_triggers.0.version", "v3"),
					resource.TestCheckResourceAttr("data.auth0_action.my_action", "runtime", "node22"),
				),
			},
		},
	})
}
