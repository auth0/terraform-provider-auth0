package portal_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAPortal = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }

  navigation {
    sidebar {
      component {
        internal_link_v2 {
          label     = "Account"
          page_slug = "account"
        }
      }
    }
  }

  pages {
    default = "account"

    standard {
      title = "Account"
      slug  = "account"

      component {
        structure_section {
          variant = "card"

          child {
            my_account_mfa_management {}
          }
        }
      }
    }
  }
}
`

const testAccDataSourcePortalWithoutID = testAccGivenAPortal + `
data "auth0_portal" "my_portal" {
  depends_on = [auth0_portal.my_portal]
}
`

const testAccDataSourcePortalWithInvalidID = testAccGivenAPortal + `
data "auth0_portal" "my_portal" {
  depends_on = [auth0_portal.my_portal]
  id         = "up_aaaaaaaaaaaaaaaaaaaaaa"
}
`

const testAccDataSourcePortal = testAccGivenAPortal + `
data "auth0_portal" "my_portal" {
  id = auth0_portal.my_portal.id
}
`

func TestAccDataSourcePortal(t *testing.T) {
	const name = "data.auth0_portal.my_portal"
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccDataSourcePortalWithoutID, testName),
				ExpectError: regexp.MustCompile(`The argument "id" is required`),
			},
			{
				Config:      acctest.ParseTestName(testAccDataSourcePortalWithInvalidID, testName),
				ExpectError: regexp.MustCompile(`404|[Nn]ot [Ff]ound`),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourcePortal, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(name, "id", "auth0_portal.my_portal", "id"),
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test - "+testName),
					resource.TestCheckResourceAttr(name, "slug", testName),
					resource.TestCheckResourceAttrPair(name, "client.0.client_secret_post.0.client_id", "auth0_client.portal", "client_id"),
					resource.TestCheckNoResourceAttr(name, "client.0.client_secret_post.0.client_secret"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.internal_link_v2.0.label", "Account"),
					resource.TestCheckResourceAttr(name, "pages.0.default", "account"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.variant", "card"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.child.0.my_account_mfa_management.#", "1"),
					resource.TestCheckResourceAttrSet(name, "created_at"),
					resource.TestCheckResourceAttrSet(name, "updated_at"),
				),
			},
		},
	})
}
