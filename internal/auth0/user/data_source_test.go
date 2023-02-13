package user_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccDataSourceUser = `
resource "auth0_role" "owner" {
	name        = "owner"
	description = "Owner"
}

resource "auth0_role" "admin" {
	depends_on = [auth0_role.owner]

	name        = "admin"
	description = "Administrator"
}

resource "auth0_user" "user" {
	user_id = "{{.testName}}"
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

data "auth0_user" "test" {
	user_id = auth0_user.user.id
}
`

func TestAccDataSourceUser(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(httpRecorder),
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccDataSourceUser, strings.ToLower(t.Name())),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user.test", "email", fmt.Sprintf("%s@acceptance.test.com", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("data.auth0_user.test", "user_id", fmt.Sprintf("auth0|%s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("data.auth0_user.test", "username", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("data.auth0_user.test", "email", fmt.Sprintf("%s@acceptance.test.com", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("data.auth0_user.test", "name", "Firstname Lastname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "given_name", "Firstname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "family_name", "Lastname"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "nickname", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr("data.auth0_user.test", "picture", "https://www.example.com/picture.jpg"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "roles.#", "2"),
					resource.TestCheckResourceAttr("data.auth0_user.test", "user_metadata", `{"baz":"qux","foo":"bar"}`),
					resource.TestCheckResourceAttr("data.auth0_user.test", "app_metadata", `{"baz":"qux","foo":"bar"}`),
				),
			},
		},
	})
}
