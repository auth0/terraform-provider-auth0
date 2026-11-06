package portal_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAPortalClient = `
resource "auth0_client" "portal" {
  name      = "Acceptance Test - Portal - {{.testName}}"
  app_type  = "regular_web"
  callbacks = ["https://example.com/portals/auth/callback"]
}
`

const testAccPortalMinimal = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }
}
`

const testAccPortalFull = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }

  custom_code {
    head = "<style>body { margin: 0; }</style>"
  }

  navigation {
    sidebar {
      component {
        internal_link_v2 {
          label     = "Account"
          page_slug = "account"
          icon      = "user"
        }
      }

      component {
        external_link_v2 {
          label = "Help"
          href  = "https://help.example.com"
          icon  = "circle-help"
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
        typography_heading {
          title       = "Your account"
          description = "Manage how you sign in."
        }
      }

      component {
        typography_rich_text {
          content = "<p>Keep your sign-in methods up to date.</p>"
        }
      }

      component {
        media_image {
          src      = "https://example.com/banner.png"
          position = "center"
          height   = 120
        }
      }

      component {
        structure_separator {
          variant = "solid"
          text    = "Security"
        }
      }

      component {
        form {
          form_id            = "ap_aaaaaaaaaaaaaaaaaaaaaa"
          completion_message = "Your profile is saved."
        }
      }

      component {
        my_account_passkey_management {}
      }

      component {
        my_account_mfa_management {}
      }

      component {
        structure_section {
          variant     = "card"
          title       = "More"
          description = "Other settings."

          child {
            typography_heading {
              title = "Profile"
            }
          }

          child {
            structure_separator {}
          }

          child {
            my_account_mfa_management {}
          }
        }
      }
    }

    standard {
      title = "Organization"
      slug  = "organization"

      component {
        my_organization_details_edit {}
      }

      component {
        my_organization_domain_table {}
      }

      component {
        my_organization_member_management {}
      }
    }
  }
}
`

const testAccPortalUpdated = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}} - Updated"
  slug = "{{.testName}}"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }

  custom_code {
    head = "<style>body { margin: 1px; }</style>"
  }

  navigation {
    sidebar {
      component {
        external_link_v2 {
          label = "Support"
          href  = "https://support.example.com"
        }
      }
    }
  }

  pages {
    default = "organization"

    standard {
      title = "Organization"
      slug  = "organization"

      component {
        structure_section {
          variant = "none"

          child {
            my_organization_member_management {}
          }
        }
      }
    }
  }
}
`

const testAccPortalEmptyBlocks = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}-renamed"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }

  custom_code {}

  navigation {}

  pages {}
}
`

const testAccPortalRemovedBlocks = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}-renamed"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }
}
`

const testAccPortalWriteOnlySecretVersion1 = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}-renamed"

  client {
    client_secret_post {
      client_id                = auth0_client.portal.client_id
      client_secret_wo         = "acceptance-test-secret-wo"
      client_secret_wo_version = 1
    }
  }
}
`

const testAccPortalWriteOnlySecretVersion2 = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}-renamed"

  client {
    client_secret_post {
      client_id                = auth0_client.portal.client_id
      client_secret_wo         = "acceptance-test-secret-wo-rotated"
      client_secret_wo_version = 2
    }
  }
}
`

const testAccPortalTwoTypesInOneComponent = testAccGivenAPortalClient + `
resource "auth0_portal" "my_portal" {
  name = "Acceptance Test - {{.testName}}"
  slug = "{{.testName}}"

  client {
    client_secret_post {
      client_id     = auth0_client.portal.client_id
      client_secret = "acceptance-test-secret"
    }
  }

  pages {
    standard {
      title = "Account"
      slug  = "account"

      component {
        my_account_passkey_management {}
        my_account_mfa_management {}
      }
    }
  }
}
`

func TestAccPortal(t *testing.T) {
	const name = "auth0_portal.my_portal"
	const secretPost = "client.0.client_secret_post.0."
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccPortalTwoTypesInOneComponent, testName),
				ExpectError: regexp.MustCompile("`pages.standard\\[0\\].component\\[0\\]` must set exactly one of"),
			},
			{
				Config: acctest.ParseTestName(testAccPortalMinimal, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(name, "id", regexp.MustCompile(`^up_`)),
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test - "+testName),
					resource.TestCheckResourceAttr(name, "slug", testName),
					resource.TestCheckResourceAttrPair(name, secretPost+"client_id", "auth0_client.portal", "client_id"),
					resource.TestCheckResourceAttr(name, secretPost+"client_secret", "acceptance-test-secret"),
					resource.TestCheckResourceAttr(name, "custom_code.#", "0"),
					resource.TestCheckResourceAttr(name, "navigation.#", "0"),
					resource.TestCheckResourceAttr(name, "pages.#", "0"),
					resource.TestCheckResourceAttrSet(name, "created_at"),
					resource.TestCheckResourceAttrSet(name, "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPortalFull, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "custom_code.0.head", "<style>body { margin: 0; }</style>"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.#", "2"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.internal_link_v2.0.label", "Account"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.internal_link_v2.0.page_slug", "account"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.internal_link_v2.0.icon", "user"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.1.external_link_v2.0.label", "Help"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.1.external_link_v2.0.href", "https://help.example.com"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.1.external_link_v2.0.icon", "circle-help"),
					resource.TestCheckResourceAttr(name, "pages.0.default", "account"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.#", "2"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.title", "Account"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.slug", "account"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.#", "8"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.typography_heading.0.title", "Your account"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.typography_heading.0.description", "Manage how you sign in."),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.1.typography_rich_text.0.content", "<p>Keep your sign-in methods up to date.</p>"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.2.media_image.0.src", "https://example.com/banner.png"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.2.media_image.0.position", "center"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.2.media_image.0.height", "120"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.3.structure_separator.0.variant", "solid"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.3.structure_separator.0.text", "Security"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.4.form.0.form_id", "ap_aaaaaaaaaaaaaaaaaaaaaa"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.4.form.0.completion_message", "Your profile is saved."),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.5.my_account_passkey_management.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.6.my_account_mfa_management.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.variant", "card"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.title", "More"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.description", "Other settings."),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.child.#", "3"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.child.0.typography_heading.0.title", "Profile"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.child.1.structure_separator.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.7.structure_section.0.child.2.my_account_mfa_management.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.1.component.#", "3"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.1.component.0.my_organization_details_edit.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.1.component.1.my_organization_domain_table.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.1.component.2.my_organization_member_management.#", "1"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					secretPost + "client_secret",
				},
			},
			{
				Config: acctest.ParseTestName(testAccPortalUpdated, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test - "+testName+" - Updated"),
					resource.TestCheckResourceAttr(name, "custom_code.0.head", "<style>body { margin: 1px; }</style>"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.#", "1"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.internal_link_v2.#", "0"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.external_link_v2.0.label", "Support"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.0.component.0.external_link_v2.0.icon", ""),
					resource.TestCheckResourceAttr(name, "pages.0.default", "organization"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.variant", "none"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.title", ""),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.child.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.0.component.0.structure_section.0.child.0.my_organization_member_management.#", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPortalEmptyBlocks, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "slug", testName+"-renamed"),
					resource.TestCheckResourceAttr(name, "custom_code.#", "1"),
					resource.TestCheckResourceAttr(name, "custom_code.0.head", ""),
					resource.TestCheckResourceAttr(name, "navigation.#", "1"),
					resource.TestCheckResourceAttr(name, "navigation.0.sidebar.#", "0"),
					resource.TestCheckResourceAttr(name, "pages.#", "1"),
					resource.TestCheckResourceAttr(name, "pages.0.standard.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPortalRemovedBlocks, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "custom_code.#", "0"),
					resource.TestCheckResourceAttr(name, "navigation.#", "0"),
					resource.TestCheckResourceAttr(name, "pages.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPortalWriteOnlySecretVersion1, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, secretPost+"client_secret", ""),
					resource.TestCheckNoResourceAttr(name, secretPost+"client_secret_wo"),
					resource.TestCheckResourceAttr(name, secretPost+"client_secret_wo_version", "1"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccPortalWriteOnlySecretVersion2, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(name, secretPost+"client_secret_wo"),
					resource.TestCheckResourceAttr(name, secretPost+"client_secret_wo_version", "2"),
				),
			},
		},
	})
}
