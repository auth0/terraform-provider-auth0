package selfserviceprofile_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testSelfServiceProfileCreate = `
resource "auth0_self_service_profile" "my_self_service_profile" {
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

func TestSelfServiceProfile(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testSelfServiceProfileCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.name", fmt.Sprintf("sample-name-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.description", "sample-description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.is_optional", "true"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "branding.0.logo_url", "https://mycompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "branding.0.colors.0.primary", "#0059d6"),
				),
			},
			{
				Config: acctest.ParseTestName(testSelfServiceProfileUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.name", fmt.Sprintf("updated-sample-name-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.description", "updated-sample-description"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "user_attributes.0.is_optional", "true"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "branding.0.logo_url", "https://newcompany.org/v2/logo.png"),
					resource.TestCheckResourceAttr("auth0_self_service_profile.ssp", "branding.0.colors.0.primary", "#000000"),
				),
			},
		},
	})
}
