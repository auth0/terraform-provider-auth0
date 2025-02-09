package user_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccUserMissingRequiredParams(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      "resource auth0_user user {}",
				ExpectError: regexp.MustCompile(`The argument "connection_name" is required`),
			},
		},
	})
}

const testAccUserEmpty = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	user_id         = "{{.testName}}"
	username        = "{{.testName}}"
	password        = "passpass$12$12"
	email           = "{{.testName}}@acceptance.test.com"
}
`

const testAccUserUpdate = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	user_id         = "{{.testName}}"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
}
`

const testAccUserUpdateWithMetadataWithTwoElements = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
	user_metadata   = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
	app_metadata    = jsonencode({
		"foo": "bar",
		"baz": "qux"
	})
}
`

const testAccUserUpdateMetadataByRemovingOneElement = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
	user_metadata   = jsonencode({
		"foo": "bars",
	})
	app_metadata    = jsonencode({
		"foo": "bars",
	})
}
`

const testAccUserUpdatingMetadataBySettingToEmpty = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
	user_metadata   =  ""
	app_metadata    =  ""
}
`

const testAccUserUpdateRemovingMetadata = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	username        = "{{.testName}}"
	email           = "{{.testName}}@acceptance.test.com"
	password        = "passpass$12$12"
	name            = "Firstname Lastname"
	given_name      = "Firstname"
	family_name     = "Lastname"
	nickname        = "{{.testName}}"
	picture         = "https://www.example.com/picture.jpg"
}
`

func TestAccUser(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserEmpty, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "connection_name", "Username-Password-Authentication"),
					resource.TestCheckResourceAttr("auth0_user.user", "email", fmt.Sprintf("%s@acceptance.test.com", testName)),
					resource.TestCheckResourceAttr("auth0_user.user", "user_id", fmt.Sprintf("auth0|%s", testName)),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserUpdate, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "connection_name", "Username-Password-Authentication"),
					resource.TestCheckResourceAttr("auth0_user.user", "username", testName),
					resource.TestCheckResourceAttr("auth0_user.user", "user_id", fmt.Sprintf("auth0|%s", testName)),
					resource.TestCheckResourceAttr("auth0_user.user", "email", fmt.Sprintf("%s@acceptance.test.com", testName)),
					resource.TestCheckResourceAttr("auth0_user.user", "name", "Firstname Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "given_name", "Firstname"),
					resource.TestCheckResourceAttr("auth0_user.user", "family_name", "Lastname"),
					resource.TestCheckResourceAttr("auth0_user.user", "nickname", testName),
					resource.TestCheckResourceAttr("auth0_user.user", "picture", "https://www.example.com/picture.jpg"),
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserUpdateWithMetadataWithTwoElements, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", `{"baz":"qux","foo":"bar"}`),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", `{"baz":"qux","foo":"bar"}`),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserUpdateMetadataByRemovingOneElement, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", `{"foo":"bars"}`),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", `{"foo":"bars"}`),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserUpdatingMetadataBySettingToEmpty, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserUpdateRemovingMetadata, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.user", "user_metadata", ""),
					resource.TestCheckResourceAttr("auth0_user.user", "app_metadata", ""),
				),
			},
		},
	})
}

const testAccUserChangeUsernameCreate = `
resource "auth0_user" "auth0_user_change_username" {
	connection_name = "Username-Password-Authentication"
	username        = "user_{{.testName}}"
	email           = "change.username.{{.testName}}@acceptance.test.com"
	email_verified  = true
	password        = "MyPass123$"
}
`

const testAccUserChangeUsernameUpdate = `
resource "auth0_user" "auth0_user_change_username" {
	connection_name = "Username-Password-Authentication"
	username        = "user_x_{{.testName}}"
	email           = "change.username.{{.testName}}@acceptance.test.com"
	email_verified  = true
	password        = "MyPass123$"
}
`

const testAccUserChangeUsernameAndPassword = `
resource "auth0_user" "auth0_user_change_username" {
	connection_name = "Username-Password-Authentication"
	username        = "user_{{.testName}}"
	email           = "change.username.{{.testName}}@acceptance.test.com"
	email_verified  = true
	password        = "MyPass123456$"
}
`

func TestAccUserChangeUsername(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserChangeUsernameCreate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserChangeUsernameUpdate, "terra"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "username", "user_x_terra"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "email", "change.username.terra@acceptance.test.com"),
					resource.TestCheckResourceAttr("auth0_user.auth0_user_change_username", "password", "MyPass123$"),
				),
			},
			{
				Config:      acctest.ParseTestName(testAccUserChangeUsernameAndPassword, "terra"),
				ExpectError: regexp.MustCompile("cannot update username and password simultaneously"),
			},
		},
	})
}
