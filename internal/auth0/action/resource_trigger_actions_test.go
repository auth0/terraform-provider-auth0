package action_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccTriggerActionsAction = `
resource auth0_action action_foo {
	name = "Test Trigger Binding Foo {{.testName}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log("foo")
	};"
	EOT
	deploy = true
}

resource auth0_action action_bar {
	depends_on = [auth0_action.action_foo]
	name = "Test Trigger Binding Bar {{.testName}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log("bar")
	};"
	EOT
	deploy = true
}
`

const testAccTriggerActionsConfigCreate = testAccTriggerActionsAction + `
resource auth0_trigger_actions login_flow {
	trigger = "post-login"
	actions {
		id = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}
}
`

const testAccTriggerActionsConfigUpdate = testAccTriggerActionsAction + `
resource auth0_trigger_actions login_flow {
	trigger = "post-login"
	actions {
		id = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}
	actions {
		id = auth0_action.action_bar.id
		display_name = auth0_action.action_bar.name
	}
}
`

const testAccTriggerActionsConfigUpdateAgain = testAccTriggerActionsAction + `
resource auth0_trigger_actions login_flow {
	trigger = "post-login"
	actions {
		id = auth0_action.action_bar.id # <----- change the order of the actions
		display_name = auth0_action.action_bar.name
	}
	actions {
		id = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}
}
`

const testAccTriggerActionsConfigRemoveAction = testAccTriggerActionsAction + `
resource auth0_trigger_actions login_flow {
	trigger = "post-login"
	actions {
		id = auth0_action.action_bar.id
		display_name = auth0_action.action_bar.name
	}
}
`

func TestAccTriggerActions(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.action_foo", "name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.action_bar", "name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "1"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.action_foo", "name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.action_bar", "name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "2"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.1.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.action_foo", "name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.action_bar", "name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "2"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.1.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigRemoveAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.action_foo", "name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.action_bar", "name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "1"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
		},
	})
}

func TestAccTriggerActions_Import(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// Only run with recorded HTTP requests, as it is required to import an already
		// existing trigger binding that is created outside the scope of terraform.
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: `
					resource "auth0_trigger_actions" "test_flow"{
						trigger = "post-user-registration"
						actions {
							id = "c2d5b219-4390-4bea-8a1f-c61672b54db3"
							display_name = "Test"
						}
					}
				`,
				ResourceName:       "auth0_trigger_actions.test_flow",
				ImportState:        true,
				ImportStateId:      "post-user-registration",
				ImportStatePersist: true,
			},
			{
				Config: `
					resource "auth0_trigger_actions" "test_flow"{
						trigger = "post-user-registration"
						actions {
							id = "c2d5b219-4390-4bea-8a1f-c61672b54db3"
							display_name = "Test"
						}
					}
				`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.test_flow", "id", "post-user-registration"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.test_flow", "trigger", "post-user-registration"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.test_flow", "actions.#", "1"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.test_flow", "actions.0.display_name", "Test"),
				),
			},
		},
	})
}
