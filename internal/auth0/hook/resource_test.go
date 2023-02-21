package hook_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccHookEmpty = `
resource "auth0_hook" "my_hook" {
  name = "pre-user-reg-hook"
  script = "function (user, context, callback) { callback(null, { user }); }"
  trigger_id = "pre-user-registration"
}
`

const testAccHookCreate = `
resource "auth0_hook" "my_hook" {
  name = "pre-user-reg-hook"
  script = "function (user, context, callback) { callback(null, { user }); }"
  trigger_id = "pre-user-registration"
  enabled = true
  %s
}
`

func TestAccHook(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccHookEmpty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttrSet("auth0_hook.my_hook", "enabled"),
					resource.TestCheckNoResourceAttr("auth0_hook.my_hook", "secrets"),
					resource.TestCheckNoResourceAttr("auth0_hook.my_hook", "dependencies"),
				),
			},
			{
				Config: fmt.Sprintf(testAccHookCreate, ""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "enabled", "true"),
				),
			},
		},
	})
}

const testAccHookSecrets = `
  dependencies = {
    auth0 = "2.30.0"
  }
  secrets = {
    foo = "alpha"
  }
`

const testAccHookSecretsUpdate = `
  dependencies = {
    auth0 = "2.30.0"
  }
  secrets = {
    foo = "gamma"
    bar = "kappa"
  }
`

const testAccHookSecretsUpdateAndRemoval = `
  dependencies = {
    auth0 = "2.30.0"
  }
  secrets = {
    foo = "delta"
  }
`

const testAccHookSecretsEmpty = `
  dependencies = {}
  secrets = {}
`

func TestAccHookSecrets(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccHookCreate, testAccHookSecrets),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "dependencies.auth0", "2.30.0"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "secrets.foo", "alpha"),
					resource.TestCheckNoResourceAttr("auth0_hook.my_hook", "secrets.bar"),
				),
			},
			{
				Config: fmt.Sprintf(testAccHookCreate, testAccHookSecretsUpdate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "dependencies.auth0", "2.30.0"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "secrets.foo", "gamma"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "secrets.bar", "kappa"),
				),
			},
			{
				Config: fmt.Sprintf(testAccHookCreate, testAccHookSecretsUpdateAndRemoval),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "secrets.foo", "delta"),
					resource.TestCheckNoResourceAttr("auth0_hook.my_hook", "secrets.bar"),
				),
			},
			{
				Config: fmt.Sprintf(testAccHookCreate, testAccHookSecretsEmpty),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "name", "pre-user-reg-hook"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "script", "function (user, context, callback) { callback(null, { user }); }"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "trigger_id", "pre-user-registration"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "enabled", "true"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "secrets.%", "0"),
					resource.TestCheckResourceAttr("auth0_hook.my_hook", "dependencies.%", "0"),
				),
			},
		},
	})
}
