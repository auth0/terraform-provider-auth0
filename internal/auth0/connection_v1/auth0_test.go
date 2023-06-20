package connection_v1_test //nolint:all temporarily until v0 connection resource removed

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccConnection(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccConnectionConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "name", fmt.Sprintf("Acceptance-Test-Connection-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "is_domain_connection", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "strategy", "auth0"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "metadata.key1", "foo"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "metadata.key2", "bar"),
					resource.TestCheckNoResourceAttr("auth0_connection_auth0.my_connection", "show_as_button"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "password_policy", "fair"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "password_no_personal_info.0.enable", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "password_dictionary.0.enable", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "password_complexity_options.0.min_length", "6"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "enabled_database_customization", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "brute_force_protection", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "import_mode", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "disable_signup", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "disable_self_service_change_password", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "requires_username", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "validation.0.username.0.min", "10"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "validation.0.username.0.max", "40"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "custom_scripts.get_user", "myFunction"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "mfa.0.active", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "mfa.0.return_enroll_settings", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "upstream_params", "{\"screen_name\":{\"alias\":\"login_hint\"}}"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "non_persistent_attrs.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_auth0.my_connection", "non_persistent_attrs.*", "hair_color"),
					resource.TestCheckTypeSetElemAttr("auth0_connection_auth0.my_connection", "non_persistent_attrs.*", "gender"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccConnectionConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "brute_force_protection", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "mfa.0.return_enroll_settings", "false"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "upstream_params", ""),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "enable_script_context", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "enabled_database_customization", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "disable_self_service_change_password", "true"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "set_user_root_attributes", "on_first_login"),
					resource.TestCheckResourceAttr("auth0_connection_auth0.my_connection", "non_persistent_attrs.#", "0"),
				),
			},
		},
	})
}

const testAccConnectionConfig = `
resource "auth0_connection_auth0" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true

	metadata = {
		key1 = "foo"
		key2 = "bar"
	}

	password_policy = "fair"
	password_history {
		enable = true
		size = 5
	}
	password_no_personal_info {
		enable = true
	}
	password_dictionary {
		enable = true
		dictionary = [ "password", "admin", "1234" ]
	}
	password_complexity_options {
		min_length = 6
	}
	validation {
		username {
			min = 10
			max = 40
		}
	}
	enabled_database_customization = false
	brute_force_protection = true
	import_mode = false
	requires_username = true
	disable_signup = false
	disable_self_service_change_password = false
	custom_scripts = {
		get_user = "myFunction"
	}
	configuration = {
		foo = "bar"
	}
	mfa {
		active                 = true
		return_enroll_settings = true
	}
	upstream_params = jsonencode({
		"screen_name": {
			"alias": "login_hint"
		}
	})
	non_persistent_attrs = ["gender","hair_color"]
}
`

const testAccConnectionConfigUpdate = `
resource "auth0_connection_auth0" "my_connection" {
	name = "Acceptance-Test-Connection-{{.testName}}"
	is_domain_connection = true

	metadata = {
		key1 = "foo"
		key2 = "bar"
	}

	password_policy = "fair"
	password_history {
		enable = true
		size = 5
	}
	password_no_personal_info {
		enable = true
	}
	enable_script_context = true
	enabled_database_customization = true
	set_user_root_attributes = "on_first_login"
	brute_force_protection = false
	import_mode = false
	disable_signup = false
	disable_self_service_change_password = true
	requires_username = true
	custom_scripts = {
		get_user = "myFunction"
	}
	configuration = {
		foo = "bar"
	}
	mfa {
		active                 = true
		return_enroll_settings = false
	}
	non_persistent_attrs = []
}
`
