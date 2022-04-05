package auth0

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
)

func TestAccAction(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccActionConfigCreate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_action.my_action", "name", "Test Action {{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onExecutePostLogin = async (event, api) => {};"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "1"),
				),
			},
			{
				Config: random.Template(testAccActionConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_action.my_action", "name", "Test Action {{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_action.my_action", "code", "exports.onContinuePostLogin = async (event, api) => {};"),
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "1"),
				),
			},
			{
				Config: random.Template(testAccActionConfigUpdateAgain, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_action.my_action", "name", "Test Action {{.random}}", rand),
					resource.TestCheckResourceAttrSet("auth0_action.my_action", "version_id"),
					resource.TestCheckResourceAttr("auth0_action.my_action", "secrets.#", "0"),
				),
			},
		},
	})
}

func TestAccAction_FailedBuild(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"auth0": Provider(),
		},
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccActionConfigCreateWithFailedBuild, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_action.my_action", "name", "Test Action {{.random}}", rand),
				),
				ExpectError: regexp.MustCompile(
					fmt.Sprintf(`action "Test Action %s" failed to build, check the Auth0 UI for errors`, rand),
				),
			},
		},
	})
}

const testAccActionConfigCreate = `
resource auth0_action my_action {
	name = "Test Action {{.random}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	secrets {
		name = "foo"
		value = "123"
	}
	code = "exports.onExecutePostLogin = async (event, api) => {};"
}
`

const testAccActionConfigUpdate = `
resource auth0_action my_action {
	name = "Test Action {{.random}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	secrets {
		name = "foo"
		value = "123456"
	}
	code = "exports.onContinuePostLogin = async (event, api) => {};"
	deploy = true
}
`

const testAccActionConfigUpdateAgain = `
resource auth0_action my_action {
	name = "Test Action {{.random}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log(event)
	};"
	EOT
	deploy = true
}
`

// This config makes use of a crypto dependency definition that causes the
// action build to fail.  This is presumably because the crypto package has been
// deprecated: https://www.npmjs.com/package/crypto
//
// If this is ever fixed in the API, another means of failing the build will
// need to be used here.
const testAccActionConfigCreateWithFailedBuild = `
resource auth0_action my_action {
	name = "Test Action {{.random}}"
	supported_triggers {
		id = "post-login"
		version = "v2"
	}
	code = <<-EOT
	exports.onContinuePostLogin = async (event, api) => {
		console.log(event)
	};"
	EOT
	runtime = "node16"
	dependencies {
		name    = "crypto"
		version = "17.7.1"
	}
	deploy = true
}
`
