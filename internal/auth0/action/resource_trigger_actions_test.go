package action_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenTwoActions = `
resource "auth0_action" "action_foo" {
	name = "Test Trigger Binding Foo {{.testName}}"
	deploy = true

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log("foo")
	};"
	EOT
}

resource "auth0_action" "action_bar" {
	depends_on = [ auth0_action.action_foo ]

	name = "Test Trigger Binding Bar {{.testName}}"
	deploy = true

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log("bar")
	};"
	EOT
}
`

const testAccTriggerActionsConfigCreateWithOneAction = testAccGivenTwoActions + `
resource "auth0_trigger_actions" "login_flow" {
	depends_on = [ auth0_action.action_bar ]

	trigger = "post-login"

	actions {
		id           = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}
}
`

const testAccTriggerActionsConfigUpdateWithTwoActions = testAccGivenTwoActions + `
resource "auth0_trigger_actions" "login_flow" {
	depends_on = [ auth0_action.action_bar ]

	trigger = "post-login"

	actions {
		id           = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}

	actions {
		id           = auth0_action.action_bar.id
		display_name = auth0_action.action_bar.name
	}
}
`

const testAccTriggerActionsConfigUpdateAgain = testAccGivenTwoActions + `
resource "auth0_trigger_actions" "login_flow" {
	depends_on = [ auth0_action.action_bar ]

	trigger = "post-login"

	actions {
		id           = auth0_action.action_bar.id # <----- change the order of the actions
		display_name = auth0_action.action_bar.name
	}

	actions {
		id           = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}
}
`

const testAccTriggerActionsConfigRemoveOneAction = testAccGivenTwoActions + `
resource "auth0_trigger_actions" "login_flow" {
	depends_on = [ auth0_action.action_bar ]

	trigger = "post-login"

	actions {
		id           = auth0_action.action_bar.id
		display_name = auth0_action.action_bar.name
	}
}
`

const testAccTriggerActionsImportSetup = testAccGivenTwoActions + `
resource "auth0_trigger_action" "action_1" {
	depends_on = [ auth0_action.action_bar ]

	action_id    = auth0_action.action_foo.id
	trigger      = auth0_action.action_foo.supported_triggers[0].id
	display_name = auth0_action.action_foo.name
}

resource "auth0_trigger_action" "action_2" {
	depends_on = [ auth0_trigger_action.action_1 ]

	action_id    = auth0_action.action_bar.id
	trigger      = auth0_action.action_bar.supported_triggers[0].id
	display_name = auth0_action.action_bar.name
}
`

const testAccTriggerActionsImportCheck = testAccTriggerActionsImportSetup + `
resource "auth0_trigger_actions" "login_flow" {
	depends_on = [ auth0_trigger_action.action_2 ]

	trigger = "post-login"

	actions {
		id           = auth0_action.action_foo.id
		display_name = auth0_action.action_foo.name
	}

	actions {
		id           = auth0_action.action_bar.id
		display_name = auth0_action.action_bar.name
	}
}
`

func TestAccTriggerActions(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigCreateWithOneAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.0.id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigUpdateWithTwoActions, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.0.id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.1.id", "auth0_action.action_bar", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.1.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.0.id", "auth0_action.action_bar", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.1.id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.1.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsConfigRemoveOneAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.0.id", "auth0_action.action_bar", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccGivenTwoActions, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsImportSetup, t.Name()),
			},
			{
				Config:             acctest.ParseTestName(testAccTriggerActionsImportCheck, t.Name()),
				ResourceName:       "auth0_trigger_actions.login_flow",
				ImportState:        true,
				ImportStateId:      "post-login",
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionsImportCheck, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.0.id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.0.display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_actions.login_flow", "actions.1.id", "auth0_action.action_bar", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_actions.login_flow", "actions.1.display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
		},
	})
}
