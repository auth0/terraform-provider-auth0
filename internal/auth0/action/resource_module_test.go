package action_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccActionModuleCreate = `
resource "auth0_action_module" "my_module" {
	name = "Test Module {{.testName}}"
	code = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}
`

const testAccActionModuleUpdateWithDependencies = `
resource "auth0_action_module" "my_module" {
	name = "Test Module {{.testName}}"
	code = <<-EOT
		const _ = require('lodash');
		module.exports = {
			greet: function(name) {
				return "Hello, " + _.capitalize(name) + "!";
			}
		};
	EOT

	dependencies {
		name    = "lodash"
		version = "4.17.21"
	}
}
`

const testAccActionModuleUpdateWithSecrets = `
resource "auth0_action_module" "my_module" {
	name = "Test Module {{.testName}}"
	code = <<-EOT
		const _ = require('lodash');
		module.exports = {
			greet: function(name) {
				return "Hello, " + _.capitalize(name) + "!";
			},
			getApiKey: function() {
				return process.env.API_KEY;
			}
		};
	EOT

	dependencies {
		name    = "lodash"
		version = "4.17.21"
	}

	secrets {
		name  = "API_KEY"
		value = "my-secret-key"
	}
}
`

func TestAccActionModule(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_action_module.my_module", "code"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "dependencies.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionModuleUpdateWithDependencies, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_action_module.my_module", "code"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "dependencies.#", "1"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "dependencies.0.name", "lodash"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "dependencies.0.version", "4.17.21"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionModuleUpdateWithSecrets, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttrSet("auth0_action_module.my_module", "code"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "dependencies.#", "1"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "secrets.#", "1"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "secrets.0.name", "API_KEY"),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "secrets.0.value", "my-secret-key"),
				),
			},
		},
	})
}

const testAccActionModuleDataSource = `
resource "auth0_action_module" "my_module" {
	name = "Test Module {{.testName}}"
	code = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}

data "auth0_action_module" "my_module" {
	id = auth0_action_module.my_module.id
}
`

func TestAccActionModuleDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_action_module.my_module", "code"),
				),
			},
		},
	})
}
