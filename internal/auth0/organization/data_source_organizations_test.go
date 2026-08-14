package organization_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceOrganizationsConfig = `
resource "auth0_organization" "my_organization" {
	name                      = "test-org-{{.testName}}"
	display_name              = "Test Org {{.testName}}"
	is_app_entitlement_active = true
}

resource "auth0_client" "my_test_client" {
	name = "test-client-{{.testName}}"
}

resource "auth0_organization_client" "my_organization_client" {
	depends_on = [ auth0_organization.my_organization, auth0_client.my_test_client ]

	organization_id       = auth0_organization.my_organization.id
	client_id             = auth0_client.my_test_client.id
	use_for_member_access = true
}

data "auth0_organizations" "my_organizations" {
	depends_on = [ auth0_organization_client.my_organization_client ]
}

data "auth0_organizations" "my_organizations_with_client_association" {
	depends_on = [ auth0_organization_client.my_organization_client ]

	include_client_association_for = auth0_client.my_test_client.id
}
`

func TestAccDataSourceOrganizations(t *testing.T) {
	testName := strings.ToLower(t.Name())
	organizationName := "test-org-" + testName

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceOrganizationsConfig, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_organizations.my_organizations", "id", "organizations"),
					resource.TestCheckResourceAttrSet("data.auth0_organizations.my_organizations", "organizations.#"),
					// Without `include_client_association_for` no organization carries a `client` block,
					// not even the one that is associated with the client created above.
					checkOrganizationInList(
						"data.auth0_organizations.my_organizations",
						organizationName,
						map[string]string{
							"is_app_entitlement_active": "true",
							"third_party_client_access": "block",
							"client.#":                  "0",
						},
					),
					resource.TestCheckResourceAttrSet("data.auth0_organizations.my_organizations_with_client_association", "organizations.#"),
					// With the filter set, the associated organization gains the `client` block.
					checkOrganizationInList(
						"data.auth0_organizations.my_organizations_with_client_association",
						organizationName,
						map[string]string{
							"is_app_entitlement_active":      "true",
							"client.#":                       "1",
							"client.0.use_for_member_access": "true",
						},
					),
				),
			},
		},
	})
}

// organizationNameAttribute matches the flatmap key holding an organization's name
// within the `organizations` list, capturing the index of the list element.
var organizationNameAttribute = regexp.MustCompile(`^organizations\.(\d+)\.name$`)

// checkOrganizationInList finds the organization with the given name in a data source's
// `organizations` list and asserts on it. The list is tenant wide, so neither its size nor its
// ordering can be relied upon.
func checkOrganizationInList(dataSourceName, organizationName string, expected map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		resourceState, ok := state.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("%s not found in state", dataSourceName)
		}

		attributes := resourceState.Primary.Attributes

		index := ""
		for key, value := range attributes {
			matches := organizationNameAttribute.FindStringSubmatch(key)
			if matches != nil && value == organizationName {
				index = matches[1]
				break
			}
		}
		if index == "" {
			return fmt.Errorf("%s: no organization named %q in the organizations list", dataSourceName, organizationName)
		}

		for attribute, want := range expected {
			key := fmt.Sprintf("organizations.%s.%s", index, attribute)
			if got := attributes[key]; got != want {
				return fmt.Errorf("%s: %s is %q, expected %q", dataSourceName, key, got, want)
			}
		}

		return nil
	}
}
