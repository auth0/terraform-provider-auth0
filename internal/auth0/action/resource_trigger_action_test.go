package action_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateTriggerAction = `
	resource auth0_action my_action {
		name = "Test Action {{.testName}}"
		code = "exports.onExecutePostLogin = async (event, api) => {};"

		supported_triggers {
			id = "post-login"
			version = "v3"
		}

		deploy = true
	}

	resource auth0_trigger_action my_action_post_login {
		action_id = auth0_action.my_action.id
		trigger = tolist(auth0_action.my_action.supported_triggers)[0].id
	}
`

const testAccCreateAnotherTriggerAction = `
	resource auth0_action another_action {
		name = "Test Action {{.testName}} (another)"
		code = "exports.onExecutePostLogin = async (event, api) => {};"

		supported_triggers {
			id = "post-login"
			version = "v3"
		}

		deploy = true
	}

	resource auth0_trigger_action another_action_post_login {
		action_id = auth0_action.another_action.id
		trigger = tolist(auth0_action.another_action.supported_triggers)[0].id
	}
`

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
				Config: acctest.ParseTestName(testAccCreateAnotherTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.another_action_post_login", "trigger", "post-login"),
					resource.TestCheckResourceAttrSet("auth0_trigger_action.another_action_post_login", "action_id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckResourceAttrSet("auth0_trigger_action.my_action_post_login", "action_id"),
				),
			},
		},
	})
}
