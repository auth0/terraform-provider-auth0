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

const testAccActionModuleWithPublish = `
resource "auth0_action_module" "my_module" {
	name    = "Test Module {{.testName}}"
	publish = true
	code    = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}

data "auth0_action_module_versions" "my_module_versions" {
	module_id = auth0_action_module.my_module.id
}
`

func TestAccActionModuleVersionsDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleWithPublish, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "publish", "true"),
					resource.TestCheckResourceAttrSet("auth0_action_module.my_module", "version_id"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_versions.my_module_versions", "module_id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_versions.my_module_versions", "versions.0.id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.0.version_number", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_versions.my_module_versions", "versions.0.code"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_versions.my_module_versions", "versions.0.created_at"),
				),
			},
		},
	})
}

const testAccActionModuleActionsDataSource = `
resource "auth0_action_module" "my_module" {
	name    = "Test Module {{.testName}}"
	publish = true
	code    = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}

data "auth0_action_module_versions" "my_module_versions" {
	module_id = auth0_action_module.my_module.id
}

resource "auth0_action" "my_action" {
	name   = "Test Action {{.testName}}"
	deploy = true
	code   = <<-EOT
		const myModule = require('my-module');
		exports.onExecutePostLogin = async (event, api) => {
			console.log(myModule.greet(event.user.name));
		};
	EOT

	supported_triggers {
		id      = "post-login"
		version = "v3"
	}

	modules {
		module_id         = auth0_action_module.my_module.id
		module_version_id = data.auth0_action_module_versions.my_module_versions.versions.0.id
	}
}

data "auth0_action_module_actions" "my_module_actions" {
	depends_on = [auth0_action.my_action]
	module_id  = auth0_action_module.my_module.id
}
`

func TestAccActionModuleActionsDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleActionsDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_action.my_action", "modules.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_actions.my_module_actions", "module_id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_actions.my_module_actions", "total", "1"),
					resource.TestCheckResourceAttr("data.auth0_action_module_actions.my_module_actions", "actions.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_actions.my_module_actions", "actions.0.action_id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_actions.my_module_actions", "actions.0.action_name", fmt.Sprintf("Test Action %s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_actions.my_module_actions", "actions.0.module_version_id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_actions.my_module_actions", "actions.0.module_version_number", "1"),
				),
			},
		},
	})
}

const testAccActionModuleVersionDataSource = `
resource "auth0_action_module" "my_module" {
	name    = "Test Module {{.testName}}"
	publish = true
	code    = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}

data "auth0_action_module_versions" "my_module_versions" {
	module_id = auth0_action_module.my_module.id
}

data "auth0_action_module_version" "my_module_version" {
	module_id  = auth0_action_module.my_module.id
	version_id = data.auth0_action_module_versions.my_module_versions.versions.0.id
}
`

func TestAccActionModuleVersionDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleVersionDataSource, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_version.my_module_version", "module_id"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_version.my_module_version", "version_id"),
					resource.TestCheckResourceAttr("data.auth0_action_module_version.my_module_version", "version_number", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_version.my_module_version", "code"),
					resource.TestCheckResourceAttrSet("data.auth0_action_module_version.my_module_version", "created_at"),
				),
			},
		},
	})
}

const testAccActionModuleMultipleVersions = `
resource "auth0_action_module" "my_module" {
	name    = "Test Module {{.testName}}"
	publish = true
	code    = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "!";
			}
		};
	EOT
}

data "auth0_action_module_versions" "my_module_versions" {
	module_id = auth0_action_module.my_module.id
}
`

const testAccActionModuleMultipleVersionsUpdate = `
resource "auth0_action_module" "my_module" {
	name    = "Test Module {{.testName}}"
	publish = true
	code    = <<-EOT
		module.exports = {
			greet: function(name) {
				return "Hello, " + name + "! Updated!";
			}
		};
	EOT
}

data "auth0_action_module_versions" "my_module_versions" {
	module_id = auth0_action_module.my_module.id
}
`

func TestAccActionModuleMultipleVersions(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccActionModuleMultipleVersions, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.0.version_number", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccActionModuleMultipleVersionsUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_action_module.my_module", "name", fmt.Sprintf("Test Module %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.0.version_number", "2"),
					resource.TestCheckResourceAttr("data.auth0_action_module_versions.my_module_versions", "versions.1.version_number", "1"),
				),
			},
		},
	})
}
