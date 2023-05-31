package action_test

import (
	"os"
	"strings"
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

const testAccImportTriggerAction = `
	resource auth0_action my_action {
		name = "Test Action testaccimporttriggeraction"
		code = "exports.onExecutePostLogin = async (event, api) => {};"

		supported_triggers {
			id = "post-login"
			version = "v3"
		}
	}

	resource auth0_trigger_action my_action_post_login {
		action_id = auth0_action.my_action.id
		trigger = tolist(auth0_action.my_action.supported_triggers)[0].id
	}
`

func TestAccTriggerActionImport(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// The test runs only with recordings as it requires an initial setup.
		t.Skip()
	}

	testName := strings.ToLower(t.Name())

	importedActionID := "a6bc1597-1401-4635-b1df-de27e500ff3c"

	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:             acctest.ParseTestName(testAccImportTriggerAction, testName),
				ResourceName:       "auth0_action.my_action",
				ImportState:        true,
				ImportStateId:      importedActionID,
				ImportStatePersist: true,
			},
			{
				Config:             acctest.ParseTestName(testAccImportTriggerAction, testName),
				ResourceName:       "auth0_trigger_action.my_action_post_login",
				ImportState:        true,
				ImportStateId:      "post-login::" + importedActionID,
				ImportStatePersist: true,
			},
			{
				Config: acctest.ParseTestName(testAccImportTriggerAction, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_trigger_action.my_action_post_login", "trigger", "post-login"),
					resource.TestCheckResourceAttrSet("auth0_trigger_action.my_action_post_login", "action_id"),
				),
			},
		},
	})
}
