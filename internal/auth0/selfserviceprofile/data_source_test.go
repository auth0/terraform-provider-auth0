package selfserviceprofile_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAGivenSelfServiceProfile = `
resource "auth0_self_service_profile" "my_self_service_profile" {
	name = "my-sso-profile"
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

const testDataResourceWithoutID = testAGivenSelfServiceProfile + `
data "auth0_self_service_profile" "my_profile" {
	depends_on = [ auth0_self_service_profile.my_self_service_profile ]
}`

const testDataResourceWithValidID = testAGivenSelfServiceProfile + `
data "auth0_self_service_profile" "my_profile" {
	depends_on = [ auth0_self_service_profile.my_self_service_profile ]
    id = auth0_self_service_profile.my_self_service_profile.id
}
`

const testDataResourceWithInvalidID = testAGivenSelfServiceProfile + `
data "auth0_self_service_profile" "my_profile" {
	depends_on = [ auth0_self_service_profile.my_self_service_profile ]
    id = "ssp_bskks8aGbiq7qS13umnuvX"
}
`

const testDataResourceWithUserAttributeProfile = `
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
	name = "my-sso-profile"
	description = "sample description"
	user_attribute_profile_id = auth0_user_attribute_profile.test_profile.id
	allowed_strategies = ["oidc", "samlp"]
	branding {
		logo_url    = "https://mycompany.org/v2/logo.png"
		colors {
			primary = "#0059d6"
		}
	}
}

data "auth0_self_service_profile" "my_profile" {
	depends_on = [ auth0_self_service_profile.my_self_service_profile ]
    id = auth0_self_service_profile.my_self_service_profile.id
}
`

func TestSelfServiceDataSourceResourceRequiredId(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testDataResourceWithoutID,
				ExpectError: regexp.MustCompile("The argument \"id\" is required, but no definition was found."),
			},
		},
	})
}

func TestSelfServiceDataSourceResource(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testDataResourceWithInvalidID, t.Name()),
				ExpectError: regexp.MustCompile(
					`Error: 404 Not Found`,
				),
			},
			{
				Config: acctest.ParseTestName(testDataResourceWithValidID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_self_service_profile.my_profile",
						"user_attributes.*",
						map[string]string{
							"name":        fmt.Sprintf("sample-name-%s", t.Name()),
							"description": "sample-description",
							"is_optional": "true",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_self_service_profile.my_profile",
						"branding.*",
						map[string]string{
							"logo_url": "https://mycompany.org/v2/logo.png",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_self_service_profile.my_profile",
						"branding.*.colors.*",
						map[string]string{
							"primary": "#0059d6",
						},
					),
				),
			},
		},
	})
}

func TestSelfServiceDataSource_UserAttributeProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testDataResourceWithUserAttributeProfile, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_self_service_profile.my_profile", "name", "my-sso-profile"),
					resource.TestCheckResourceAttr("data.auth0_self_service_profile.my_profile", "description", "sample description"),
					resource.TestCheckResourceAttrSet("data.auth0_self_service_profile.my_profile", "user_attribute_profile_id"),
					resource.TestCheckResourceAttrPair("data.auth0_self_service_profile.my_profile", "user_attribute_profile_id", "auth0_user_attribute_profile.test_profile", "id"),
					resource.TestCheckResourceAttr("data.auth0_self_service_profile.my_profile", "allowed_strategies.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_self_service_profile.my_profile",
						"branding.*",
						map[string]string{
							"logo_url": "https://mycompany.org/v2/logo.png",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.auth0_self_service_profile.my_profile",
						"branding.*.colors.*",
						map[string]string{
							"primary": "#0059d6",
						},
					),
				),
			},
		},
	})
}
