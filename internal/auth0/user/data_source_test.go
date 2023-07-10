package user_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceUser = `
resource "auth0_role" "owner" {
	name        = "Test Owner {{.testName}}"
	description = "Owner {{.testName}}"
}

resource "auth0_role" "admin" {
	depends_on = [ auth0_role.owner ]

	name        = "Test Admin {{.testName}}"
	description = "Administrator {{.testName}}"
}

resource "auth0_user" "user" {
	depends_on = [ auth0_role.admin ]

	user_id         = "{{.testName}}"
	connection_name = "Username-Password-Authentication"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
	roles           = [ auth0_role.owner.id, auth0_role.admin.id ]
	user_metadata   = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
	app_metadata    = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
}

data "auth0_user" "test" {
	depends_on = [ auth0_user.user ]

	user_id = auth0_user.user.id
}
`

func TestAccDataSourceUser(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUser, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.test", "email", fmt.Sprintf("%s@acceptance.test.com", testName)),
					resource.TestCheckResourceAttr("data.auth0_user.test", "user_id", fmt.Sprintf("auth0|%s", testName)),
					resource.TestCheckResourceAttr("data.auth0_user.test", "username", testName),
					resource.TestCheckResourceAttr("data.auth0_user.test", "email", fmt.Sprintf("%s@acceptance.test.com", testName)),
					resource.TestCheckResourceAttr("data.auth0_user.test", "name", "Firstname Lastname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "given_name", "Firstname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "family_name", "Lastname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "nickname", testName),
					resource.TestCheckResourceAttr("data.auth0_user.test", "picture", "https://www.example.com/picture.jpg"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "roles.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "permissions.#", "0"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "user_metadata", `{"baz":"qux","foo":"bar"}`),
					resource.TestCheckResourceAttr("data.auth0_user.test", "app_metadata", `{"baz":"qux","foo":"bar"}`),
				),
			},
		},
	})
}

const testAccDataSourceUserDoesNotExist = `
data "auth0_user" "test" {
	user_id = "auth0|this-user-id-does-not-exist"
}
`

func TestAccDataSourceUserDoesNotExist(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUserDoesNotExist, testName),
				ExpectError: regexp.MustCompile(
					`data source with that identifier not found \((404\))`,
				),
			},
		},
	})
}
