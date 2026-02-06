package selfserviceprofile_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testSelfServiceProfileCreate = `
resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "my-sso-profile-{{.testName}}"
	description = "sample description"
	allowed_strategies = ["oidc", "samlp"]
	user_attributes	{
		name		= "sample-name-{{.testName}}"
		description = "sample-description"
		is_optional = true
	}
	branding {
		logo_url    = "https://mycompany.org/v2/logo.png"
		colors {
			primary = "#0059d6"
		}
	}
}
`

const testSelfServiceProfileUpdate = `
resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "updated-my-sso-profile-{{.testName}}"
	description = "updated sample description"
	allowed_strategies = ["oidc"]
	user_attributes	{
		name		= "updated-sample-name-{{.testName}}"
		description = "updated-sample-description"
		is_optional = true
	}
	branding {
		logo_url    = "https://newcompany.org/v2/logo.png"
		colors {
			primary = "#000000"
		}
	}
}
`

const testSelfServiceProfileUpdateWithNewAllowedStrategies = `
resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "updated-my-sso-profile-{{.testName}}"
	description = "updated sample description"
	allowed_strategies = ["oidc", "auth0-samlp", "okta-samlp"]
	user_attributes	{
		name		= "updated-sample-name-{{.testName}}"
		description = "updated-sample-description"
		is_optional = true
	}
	branding {
		logo_url    = "https://newcompany.org/v2/logo.png"
		colors {
			primary = "#000000"
		}
	}
}
`

const testSelfServiceProfileWithUserAttributeProfile = `
resource "auth0_user_attribute_profile" "test_profile" {
	name = "Test User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}

resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "my-sso-profile-with-uap-{{.testName}}"
	description = "profile with user attribute profile"
	user_attribute_profile_id = auth0_user_attribute_profile.test_profile.id
	allowed_strategies = ["oidc", "samlp"]
}
`

const testSelfServiceProfileUpdateUserAttributeProfile = `
resource "auth0_user_attribute_profile" "test_profile" {
	name = "Test User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}

resource "auth0_user_attribute_profile" "test_profile_2" {
	name = "Second User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "given_name"
		description      = "User's first name"
		label            = "First Name"
		profile_required = true
		auth0_mapping    = "given_name"
	}
}

resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "updated-sso-profile-with-uap-{{.testName}}"
	description = "updated profile with different user attribute profile"
	user_attribute_profile_id = auth0_user_attribute_profile.test_profile_2.id
	allowed_strategies = ["oidc"]
}
`

const testSelfServiceProfileRemoveUserAttributeProfile = `
resource "auth0_user_attribute_profile" "test_profile" {
	name = "Test User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}

resource "auth0_user_attribute_profile" "test_profile_2" {
	name = "Second User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "given_name"
		description      = "User's first name"
		label            = "First Name"
		profile_required = true
		auth0_mapping    = "given_name"
	}
}

resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "sso-profile-no-uap-{{.testName}}"
	description = "profile with user attribute profile removed"
	# user_attribute_profile_id removed
	allowed_strategies = ["samlp"]
	user_attributes	{
		name		= "final-sample-name-{{.testName}}"
		description = "final-sample-description"
		is_optional = true
	}
}
`

func TestSelfServiceProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testSelfServiceProfileCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "name", fmt.Sprintf("my-sso-profile-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "description", "sample description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "2"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.name", fmt.Sprintf("sample-name-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.description", "sample-description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.is_optional", "true"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "branding.0.logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "branding.0.colors.0.primary", "#0059d6"),
				),
			},
			{
				Config: acctest.ParseTestName(testSelfServiceProfileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.name", fmt.Sprintf("updated-sample-name-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.description", "updated-sample-description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "1"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.is_optional", "true"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "branding.0.logo_url", "https://newcompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "branding.0.colors.0.primary", "#000000"),
				),
			},
			{
				Config: acctest.ParseTestName(testSelfServiceProfileUpdateWithNewAllowedStrategies, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "3"),
				),
			},
		},
	})
}

const testSelfServiceProfileConflictingFields = `
resource "auth0_user_attribute_profile" "test_profile" {
	name = "Test User Attribute Profile {{.testName}}"

	user_attributes {
		name             = "email"
		description      = "User's email address"
		label            = "Email"
		profile_required = false
		auth0_mapping    = "email"
	}
}

resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "conflicting-fields-{{.testName}}"
	description = "profile with conflicting fields"
	user_attribute_profile_id = auth0_user_attribute_profile.test_profile.id
	allowed_strategies = ["oidc"]
	# Both user_attribute_profile_id and user_attributes are specified - should fail
	user_attributes	{
		name		= "sample-name-{{.testName}}"
		description = "sample-description"
		is_optional = true
	}
}
`

func TestSelfServiceProfile_ConflictingFields(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testSelfServiceProfileConflictingFields, t.Name()),
				ExpectError: regexp.MustCompile("\"user_attributes\": conflicts with user_attribute_profile_id"),
			},
		},
	})
}

func TestSelfServiceProfile_UserAttributeProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testSelfServiceProfileWithUserAttributeProfile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "name", fmt.Sprintf("my-sso-profile-with-uap-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "description", "profile with user attribute profile"),
					resource.TestCheckResourceAttrSet("auth0_self_service_profile.my_self_service_profile", "user_attribute_profile_id"),
					resource.TestCheckResourceAttrPair("auth0_self_service_profile.my_self_service_profile", "user_attribute_profile_id", "auth0_user_attribute_profile.test_profile", "id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testSelfServiceProfileUpdateUserAttributeProfile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "name", fmt.Sprintf("updated-sso-profile-with-uap-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "description", "updated profile with different user attribute profile"),
					resource.TestCheckResourceAttrSet("auth0_self_service_profile.my_self_service_profile", "user_attribute_profile_id"),
					resource.TestCheckResourceAttrPair("auth0_self_service_profile.my_self_service_profile", "user_attribute_profile_id", "auth0_user_attribute_profile.test_profile_2", "id"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testSelfServiceProfileRemoveUserAttributeProfile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "name", fmt.Sprintf("sso-profile-no-uap-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "description", "profile with user attribute profile removed"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attribute_profile_id", ""),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "allowed_strategies.#", "1"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.name", fmt.Sprintf("final-sample-name-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.description", "final-sample-description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.my_self_service_profile", "user_attributes.0.is_optional", "true"),
				),
			},
		},
	})
}
