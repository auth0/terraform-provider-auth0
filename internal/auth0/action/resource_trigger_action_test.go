package action_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccCreateTriggerAction = testAccGivenTwoActions + `
resource "auth0_trigger_action" "my_action_post_login" {
	depends_on = [ auth0_action.action_bar ]

	action_id = auth0_action.action_foo.id
	trigger   = auth0_action.action_foo.supported_triggers[0].id
}
`

const testAccUpdateTriggerAction = testAccGivenTwoActions + `
resource "auth0_trigger_action" "my_action_post_login" {
	depends_on = [ auth0_action.action_bar ]

	action_id    = auth0_action.action_foo.id
	trigger      = auth0_action.action_foo.supported_triggers[0].id
	display_name = format("%s %s", auth0_action.action_foo.name,"(new display name)")
}
`

const testAccCreateAnotherTriggerAction = testAccGivenTwoActions + `
resource "auth0_trigger_action" "my_action_post_login" {
	depends_on = [ auth0_action.action_bar ]

	action_id    = auth0_action.action_foo.id
	trigger      = auth0_action.action_foo.supported_triggers[0].id
	display_name = format("%s %s", auth0_action.action_foo.name,"(new display name)")
}

resource "auth0_trigger_action" "another_action_post_login" {
	depends_on = [ auth0_trigger_action.my_action_post_login ]

	action_id    = auth0_action.action_bar.id
	trigger      = auth0_action.action_bar.supported_triggers[0].id
	display_name = auth0_action.action_bar.name
}
`

const testAccTriggerActionImportSetup = testAccGivenTwoActions + `
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

const testAccTriggerActionImportCheck = testAccTriggerActionImportSetup + `
resource "auth0_trigger_action" "action_1" {
	depends_on = [ auth0_trigger_actions.login_flow ]

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

func TestAccTriggerAction(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccCreateTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_action.my_action_post_login", "action_id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "display_name", fmt.Sprintf("Test Trigger Binding Foo %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUpdateTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_action.my_action_post_login", "action_id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "display_name", fmt.Sprintf("Test Trigger Binding Foo %s (new display name)", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccCreateAnotherTriggerAction, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_action.my_action_post_login", "action_id", "auth0_action.action_foo", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "display_name", fmt.Sprintf("Test Trigger Binding Foo %s (new display name)", t.Name())),
					resource.TestCheckResourceAttr("auth0_trigger_action.another_action_post_login", "trigger", "post-login"),
					resource.TestCheckTypeSetElemAttrPair("auth0_trigger_action.another_action_post_login", "action_id", "auth0_action.action_bar", "id"),
					resource.TestCheckResourceAttr("auth0_trigger_action.another_action_post_login", "display_name", fmt.Sprintf("Test Trigger Binding Bar %s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccGivenTwoActions, t.Name()),
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionImportSetup, t.Name()),
			},
			{
				Config:       acctest.ParseTestName(testAccTriggerActionImportCheck, t.Name()),
				ResourceName: "auth0_trigger_action.action_1",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					actionID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_action.action_foo", "id")
					assert.NoError(t, err)

					return "post-login::" + actionID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config:       acctest.ParseTestName(testAccTriggerActionImportCheck, t.Name()),
				ResourceName: "auth0_trigger_action.action_2",
				ImportState:  true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					actionID, err := acctest.ExtractResourceAttributeFromState(state, "auth0_action.action_bar", "id")
					assert.NoError(t, err)

					return "post-login::" + actionID, nil
				},
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccTriggerActionImportCheck, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}
