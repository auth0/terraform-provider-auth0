package user_attribute_profile_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccUserAttributeProfileEmpty = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label           = "Email"
		profile_required = true
		auth0_mapping   = "email"
	}
}
`

const testAccUserAttributeProfileComplete = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label           = "Email"
		profile_required = true
		auth0_mapping   = "email"
		scim_mapping = "emails[primary eq true].value"
		saml_mapping = [
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
		]
	}

	user_attributes {
		name             = "name"
		description      = "User's full name"
		label           = "Name"
		profile_required = false
		auth0_mapping   = "name"
		scim_mapping = "displayName"
	}
}
`

const testAccUserAttributeProfileUpdate = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} Updated"

	user_attributes {
		name             = "email"
		description      = "User's updated email address"
		label           = "Email Address"
		profile_required = false
		auth0_mapping   = "email"
		saml_mapping = [
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier",
		]
	}

	user_attributes {
		name           = "given_name"
		description    = "User's first name"
		label         = "First Name"
		profile_required = true
		auth0_mapping = "given_name"
		scim_mapping  = "name.givenName"
	}
}
`

const testAccUserAttributeProfileUpdateComputed = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} Updated with Computed"

	user_attributes {
		name             = "email"
		description      = "User's email with updated mappings"
		label           = "Updated Email"
		profile_required = true
		auth0_mapping   = "email"
		saml_mapping = [
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
			"http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname",
		]
		scim_mapping = "emails[primary eq true].value"
		oidc_mapping {
			mapping      = "email_verified"
			display_name = "Email Verified Status"
		}
	}

	user_attributes {
		name           = "given_name"
		description    = "Updated first name attribute"
		label         = "Updated First Name"
		profile_required = false
		auth0_mapping = "given_name"
		scim_mapping  = "name.familyName"  # Changed from givenName
		saml_mapping = [
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
		]
	}
}
`

const testAccUserAttributeProfileWithFields = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} With Fields"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
		saml_mapping     = ["http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"]
		scim_mapping     = "emails[primary eq true].value"
	}
}
`

const testAccUserAttributeProfileFieldsRemoved = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} Fields Removed"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
		# saml_mapping and scim_mapping removed - should be cleared from API
	}
}
`

const testAccUserAttributeProfileUserIDWithFields = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} UserID With Fields"

	user_id {
		oidc_mapping = "sub"
		saml_mapping = ["urn:oid:0.9.2342.19200300.100.1.1"]
		scim_mapping = "externalId"
	}

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}
`

const testAccUserAttributeProfileUserIDFieldsRemoved = `
resource "auth0_user_attribute_profile" "test" {
	name = "{{.testName}} UserID Fields Removed"

	user_id {
		oidc_mapping = "sub"
		# saml_mapping and scim_mapping removed - should be cleared from API
	}

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}
`

func TestAccUserAttributeProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s", t.Name())),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s Updated", t.Name())),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.description", "User's updated email address"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.label", "Email Address"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.profile_required", "false"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.name", "given_name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.description", "User's first name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.label", "First Name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.profile_required", "true"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.auth0_mapping", "given_name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.scim_mapping", "name.givenName"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileUpdateComputed, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s Updated with Computed", t.Name())),
					// Test email attribute with updated computed fields
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.description", "User's email with updated mappings"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.label", "Updated Email"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.profile_required", "true"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.#", "2"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.1", "http://schemas.microsoft.com/ws/2008/06/identity/claims/windowsaccountname"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.scim_mapping", "emails[primary eq true].value"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.oidc_mapping.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.oidc_mapping.0.mapping", "email_verified"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.oidc_mapping.0.display_name", "Email Verified Status"),
					// Test given_name attribute with updated computed fields
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.name", "given_name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.description", "Updated first name attribute"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.label", "Updated First Name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.profile_required", "false"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.scim_mapping", "name.familyName"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.saml_mapping.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.1.saml_mapping.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname"),
				),
			},
		},
	})
}

func TestAccUserAttributeProfile_ImportState(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileComplete, t.Name()),
			},
			{
				ResourceName:      "auth0_user_attribute_profile.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccUserAttributeProfile_FieldRemoval(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// Test user_attributes field removal
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileWithFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s With Fields", t.Name())),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.scim_mapping", "emails[primary eq true].value"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileFieldsRemoved, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s Fields Removed", t.Name())),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.name", "email"),
					// saml_mapping and scim_mapping should be cleared from user_attributes
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.saml_mapping.#", "0"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_attributes.0.scim_mapping", ""),
				),
			},
			// Test user_id field removal
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileUserIDWithFields, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s UserID With Fields", t.Name())),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.oidc_mapping", "sub"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.#", "1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.0", "urn:oid:0.9.2342.19200300.100.1.1"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.scim_mapping", "externalId"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccUserAttributeProfileUserIDFieldsRemoved, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "name", fmt.Sprintf("%s UserID Fields Removed", t.Name())),
					// Only oidc_mapping should remain as configured by user
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.oidc_mapping", "sub"),
					// API preserves saml_mapping and scim_mapping in user_id - cannot be cleared
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.#", "3"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.0", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/nameidentifier"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.1", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/upn"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.saml_mapping.2", "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name"),
					resource.TestCheckResourceAttr("auth0_user_attribute_profile.test", "user_id.0.scim_mapping", "externalId"),
				),
			},
		},
	})
}
