package connection_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccConnectionProfiles(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccConnectionProfileConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_connection_profile.my_profile", "id"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "name", fmt.Sprintf("Test-Profile-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "organization.0.show_as_button", "required"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "organization.0.assign_membership_on_login", "optional"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "connection_name_prefix_template", "template1"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "enabled_features.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccConnectionProfileConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "name", fmt.Sprintf("Test-Profile-Updated-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "organization.0.show_as_button", "optional"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "organization.0.assign_membership_on_login", "required"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "connection_name_prefix_template", "template2"),
					resource.TestCheckResourceAttr("auth0_connection_profile.my_profile", "enabled_features.#", "1"),
				),
			},
		},
	})
}

func TestAccConnectionProfileDataSource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccConnectionProfileDataSourceConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_connection_profile.my_profile_ds", "id"),
					resource.TestCheckResourceAttr("data.auth0_connection_profile.my_profile_ds", "name", fmt.Sprintf("Test-Profile-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_connection_profile.my_profile_ds", "organization.0.show_as_button", "optional"),
					resource.TestCheckResourceAttr("data.auth0_connection_profile.my_profile_ds", "organization.0.assign_membership_on_login", "required"),
					resource.TestCheckResourceAttr("data.auth0_connection_profile.my_profile_ds", "connection_name_prefix_template", "template1"),
					resource.TestCheckResourceAttr("data.auth0_connection_profile.my_profile_ds", "enabled_features.#", "2"),
				),
			},
		},
	})
}

const testAccConnectionProfileConfig = `
resource "auth0_connection_profile" "my_profile" {
	name = "Test-Profile-{{.testName}}"

	organization {
		show_as_button            = "required"
		assign_membership_on_login = "optional"
	}

	connection_name_prefix_template = "template1"

	enabled_features = [
		"scim",
		"universal_logout"
	]

	connection_config {
	}
}
`

const testAccConnectionProfileConfigUpdate = `
resource "auth0_connection_profile" "my_profile" {
	name = "Test-Profile-Updated-{{.testName}}"

	organization {
		show_as_button            = "optional"
		assign_membership_on_login = "required"
	}

	connection_name_prefix_template = "template2"

	enabled_features = [
		"scim"
	]

	connection_config {
	}
}
`

const testAccConnectionProfileDataSourceConfig = `
resource "auth0_connection_profile" "my_profile" {
	name = "Test-Profile-{{.testName}}"

	organization {
		show_as_button            = "optional"
		assign_membership_on_login = "required"
	}

	connection_name_prefix_template = "template1"

	enabled_features = [
		"scim",
		"universal_logout"
	]

	connection_config {
	}
}

data "auth0_connection_profile" "my_profile_ds" {
	id = auth0_connection_profile.my_profile.id
}
`
