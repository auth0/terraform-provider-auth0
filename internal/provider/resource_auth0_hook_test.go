package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

func TestAccHook(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
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

func TestAccHookSecrets(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
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

func TestHookNameRegexp(t *testing.T) {
	for givenHookName, expectedError := range map[string]bool{
		"my-hook-1":                 false,
		"hook 2 name with spaces":   false,
		" hook with a space prefix": true,
		"hook with a space suffix ": true,
		" ":                         true,
		"   ":                       true,
	} {
		validationResult := validateHookName()(givenHookName, cty.Path{cty.GetAttrStep{Name: "name"}})
		assert.Equal(t, expectedError, validationResult.HasError())
	}
}
