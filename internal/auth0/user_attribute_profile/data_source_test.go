package userattributeprofile_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceUserAttributeProfileByID = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}}"

	user_id {
		oidc_mapping = "sub"
		saml_mapping = ["urn:oid:0.9.2342.19200300.100.1.1"]
		scim_mapping = "userName"
	}

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label           = "Email"
		profile_required = true
		auth0_mapping   = "email"

		oidc_mapping {
			mapping      = "email"
			display_name = "Email Address"
		}

		saml_mapping = ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]
		scim_mapping = "emails[primary eq true].value"
	}
}

data "auth0_user_attribute_profile" "test" {
	user_attribute_profile_id = auth0_user_attribute_profile.test.id
}
`

const testAccDataSourceUserAttributeProfileByName = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}}"

	user_id {
		oidc_mapping = "sub"
		saml_mapping = ["urn:oid:0.9.2342.19200300.100.1.1"]
		scim_mapping = "userName"
	}

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label           = "Email"
		profile_required = true
		auth0_mapping   = "email"

		oidc_mapping {
			mapping      = "email"
			display_name = "Email Address"
		}

		saml_mapping = ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]
		scim_mapping = "emails[primary eq true].value"
	}
}

data "auth0_user_attribute_profile" "test" {
	name = auth0_user_attribute_profile.test.name
}
`

func TestAccDataSourceUserAttributeProfile_ByID(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUserAttributeProfileByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_user_attribute_profile.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "name", t.Name()),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.oidc_mapping", "sub"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.saml_mapping.0", "urn:oid:0.9.2342.19200300.100.1.1"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.scim_mapping", "userName"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.description", "User's email address"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.label", "Email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.profile_required", "true"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.auth0_mapping", "email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.oidc_mapping.0.mapping", "email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.oidc_mapping.0.display_name", "Email Address"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.scim_mapping", "emails[primary eq true].value"),
				),
			},
		},
	})
}

func TestAccDataSourceUserAttributeProfile_ByName(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUserAttributeProfileByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_user_attribute_profile.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "name", t.Name()),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.oidc_mapping", "sub"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.saml_mapping.0", "urn:oid:0.9.2342.19200300.100.1.1"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_id.0.scim_mapping", "userName"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.description", "User's email address"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.label", "Email"),
					resource.TestCheckResourceAttr("data.auth0_user_attribute_profile.test", "user_attributes.0.profile_required", "true"),
				),
			},
		},
	})
}
