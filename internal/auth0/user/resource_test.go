package user_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func TestAccUserMissingRequiredParams(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      "resource auth0_user user {}",
				ExpectError: regexp.MustCompile(`The argument "connection_name" is required`),
			},
		},
	})
}

const testAccUserEmpty = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	user_id = "{{.testName}}"
	username = "{{.testName}}"
	password = "passpass$12$12"
	email = "{{.testName}}@acceptance.test.com"
}
`

const testAccUserUpdate = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	user_id = "{{.testName}}"
	username = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
}
`

const testAccUserUpdateWithRolesAndMetadata = `
resource auth0_user user {
	depends_on = [auth0_role.owner, auth0_role.admin]
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	roles = [ auth0_role.owner.id, auth0_role.admin.id ]
	user_metadata = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
	app_metadata = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
}

resource auth0_role owner {
	name = "owner"
	description = "Owner"
}

resource auth0_role admin {
	name = "admin"
	description = "Administrator"
	depends_on = [auth0_role.owner]
}
`

const testAccUserUpdateRemovingOneRoleAndUpdatingMetadata = `
resource auth0_user user {
	depends_on = [auth0_role.admin]
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	roles = [ auth0_role.admin.id ]
	user_metadata = jsonencode({
		"foo": "bars",
	})
	app_metadata = jsonencode({
		"foo": "bars",
	})
}

resource auth0_role admin {
	name = "admin"
	description = "Administrator"
}
`

const testAccUserUpdateRemovingAllRolesAndUpdatingMetadata = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
	user_metadata = jsonencode({
		"foo": "barss",
		"foo2": "bar2",
	})
	app_metadata = jsonencode({
		"foo": "barss",
		"foo2": "bar2",
	})
}
`

const testAccUserUpdateRemovingMetadata = `
resource auth0_user user {
	connection_name = "Username-Password-Authentication"
	username = "{{.testName}}"
	email = "{{.testName}}@acceptance.test.com"
	password = "passpass$12$12"
	name = "Firstname Lastname"
	given_name = "Firstname"
	family_name = "Lastname"
	nickname = "{{.testName}}"
	picture = "https://www.example.com/picture.jpg"
}
`

func TestAccUser(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccUserEmpty, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "connection_name", "Username-Password-Authentication"),
					resource.TestCheckResourceAttr("auth0_user.user", "email", fmt.Sprintf("%s@acceptance.test.com", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_user.user", "user_id", fmt.Sprintf("auth0|%s", strings.ToLower(t.Name()))),
				),
			},
			{
				Config: template.ParseTestName(testAccUserUpdate, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "connection_name", "Username-Password-Authentication"),
					resource.TestCheckResourceAttr("auth0_user.user", "username", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("auth0_user.user", "user_id", fmt.Sprintf("auth0|%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_user.user", "email", fmt.Sprintf("%s@acceptance.test.com", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("auth0_user.user", "name", "Firstname Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "given_name", "Firstname"),
					resource.TestCheckResourceAttr("auth0_user.user", "family_name", "Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "nickname", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("auth0_user.user", "picture", "https://www.example.com/picture.jpg"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserUpdateWithRolesAndMetadata, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "2"),
					resource.TestCheckResourceAttr("auth0_role.owner", "name", "owner"),
					resource.TestCheckResourceAttr("auth0_role.admin", "name", "admin"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", `{"baz":"qux","foo":"bar"}`),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", `{"baz":"qux","foo":"bar"}`),
				),
			},
			{
				Config: template.ParseTestName(testAccUserUpdateRemovingOneRoleAndUpdatingMetadata, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "1"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", `{"foo":"bars"}`),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", `{"foo":"bars"}`),
				),
			},
			{
				Config: template.ParseTestName(testAccUserUpdateRemovingAllRolesAndUpdatingMetadata, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", `{"foo":"barss","foo2":"bar2"}`),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", `{"foo":"barss","foo2":"bar2"}`),
				),
			},
			{
				Config: template.ParseTestName(testAccUserUpdateRemovingMetadata, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "roles.#", "0"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", ""),
				),
			},
		},
	})
}

const testAccUserChangeUsernameCreate = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123$"
}
`

const testAccUserChangeUsernameUpdate = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_x_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123$"
}
`

const testAccUserChangeUsernameAndPassword = `
resource auth0_user auth0_user_change_username {
  connection_name = "Username-Password-Authentication"
  username = "user_{{.testName}}"
  email = "change.username.{{.testName}}@acceptance.test.com"
  email_verified = true
  password = "MyPass123456$"
}
`

func TestAccUserChangeUsername(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccUserChangeUsernameCreate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config: template.ParseTestName(testAccUserChangeUsernameUpdate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_x_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config:      template.ParseTestName(testAccUserChangeUsernameAndPassword, "terra"),
				ExpectError: regexp.MustCompile("cannot update username and password simultaneously"),
			},
		},
	})
}
