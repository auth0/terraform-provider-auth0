package action_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const givenAnAction = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	code = "exports.onExecutePostLogin = async (event, api) => {};"

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	deploy = true
}
`

const testAccCreateTriggerAction = givenAnAction + `
	resource auth0_trigger_action my_action_post_login {
		action_id = auth0_action.my_action.id
		trigger = "post-login"
	}
`

const testAccRemoveTriggerAction = givenAnAction

func TestAccTriggerAction(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckResourceAttrSet("auth0_trigger_action.my_action_post_login", "action_id"),
				),
			},
			{
				// This step necessary to orchestrate teardown of resources; cannot delete action without unbinding it from trigger first.
				Config: acctest.ParseTestName(testAccRemoveTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "name"),
				),
			},
		},
	})
}
