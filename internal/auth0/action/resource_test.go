package action_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccActionConfigCreateWithOnlyRequiredFields = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	code = "exports.onExecutePostLogin = async (event, api) => {};"

	supported_triggers {
		id = "post-login"
		version = "v3"
	}
}
`

const testAccActionConfigUpdateAllFields = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	code = "exports.onContinuePostLogin = async (event, api) => {};"
	runtime = "node18"
	deploy = true

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	dependencies {
		name    = "auth0"
		version = "2.41.0"
	}

	secrets {
		name = "foo"
		value = "111111"
	}
}
`

const testAccActionConfigUpdateAgain = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	deploy = true
	code = <<-EOT
		exports.onContinuePostLogin = async (event, api) => {
			console.log(event)
		};"
	EOT

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	secrets {
		name = "foo"
		value = "123456"
	}

	secrets {
		name = "bar"
		value = "654321"
	}

	dependencies {
		name    = "auth0"
		version = "2.42.0"
	}

	dependencies {
		name    = "moment"
		version = "2.29.4"
	}
}
`

const testAccActionConfigResetToRequiredFields = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	code = <<-EOT
		exports.onContinuePostLogin = async (event, api) => {
			console.log(event)
		};"
	EOT

	supported_triggers {
		id = "post-login"
		version = "v3"
	}
}
`

func TestAccAction(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionConfigCreateWithOnlyRequiredFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onExecutePostLogin = async (event, api) => {};"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.#", "0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "runtime", "node18"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "deploy", "false"),
					resource.TestCheckNoResourceAttr("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.version", "v3"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionConfigUpdateAllFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onContinuePostLogin = async (event, api) => {};"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "runtime", "node18"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "deploy", "true"),
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.version", "v3"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.0.name", "auth0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.0.version", "2.41.0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.0.name", "foo"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.0.value", "111111"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionConfigUpdateAgain, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onContinuePostLogin = async (event, api) => {\n\tconsole.log(event)\n};\"\n"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "runtime", "node18"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "deploy", "true"),
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.version", "v3"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.#", "2"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.0.name", "auth0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.0.version", "2.42.0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.1.name", "moment"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.1.version", "2.29.4"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "2"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.0.name", "foo"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.0.value", "123456"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.1.name", "bar"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.1.value", "654321"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionConfigResetToRequiredFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onContinuePostLogin = async (event, api) => {\n\tconsole.log(event)\n};\"\n"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "runtime", "node18"),
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.#", "1"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.id", "post-login"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "supported_triggers.0.version", "v3"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "dependencies.#", "0"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "0"),
				),
			},
		},
	})
}

// This config makes use of a crypto dependency definition that causes the
// action build to fail.  This is because the crypto package has been
// deprecated https://www.npmjs.com/package/crypto.
//
// If this is ever fixed in the API, another means of failing the build will
// need to be used here.
const testAccActionConfigCreateWithFailedBuild = `
resource auth0_action my_action {
	name = "Test Action {{.testName}}"
	runtime = "node18"
	deploy = true
	code = <<-EOT
		exports.onContinuePostLogin = async (event, api) => {
			console.log(event)
		};"
	EOT

	supported_triggers {
		id = "post-login"
		version = "v3"
	}

	dependencies {
		name    = "crypto"
		version = "17.7.1"
	}
}
`

func TestAccAction_FailedBuild(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionConfigCreateWithFailedBuild, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
				),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf(
						`action "Test Action %s" failed to build, check the Auth0 UI for errors`,
						t.Name(),
					),
				),
			},
		},
	})
}
