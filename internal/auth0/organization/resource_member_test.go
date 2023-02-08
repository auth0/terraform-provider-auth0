package organization_test

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func TestAccOrganizationMember(t *testing.T) {
	httpRecorder := recorder.New(t)

	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(httpRecorder),
		Steps: []resource.TestStep{{
			Config: template.ParseTestName(testAccOrganizationMembersAux+`
			resource auth0_organization_member test_member {
				depends_on = [ auth0_user.user, auth0_organization.some_org ]
				organization_id = auth0_organization.some_org.id
				user_id = auth0_user.user.id
			}
			`, testName),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr("auth0_organization_member.test_member", "roles.#", "0"),
				resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "organization_id", "auth0_organization.some_org", "id"),
				resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "user_id", "auth0_user.user", "id"),
			),
		},
			{
				Config: template.ParseTestName(testAccOrganizationMembersAux+`
				resource auth0_organization_member test_member {
					depends_on = [ auth0_user.user, auth0_organization.some_org, auth0_role.reader ]
					organization_id = auth0_organization.some_org.id
					user_id = auth0_user.user.id
					roles = [ auth0_role.reader.id ] // Adding role
				}
			`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "organization_id", "auth0_organization.some_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "user_id", "auth0_user.user", "id"),
					resource.TestCheckResourceAttr("auth0_organization_member.test_member", "roles.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.*", "auth0_role.reader", "id"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationMembersAux+`
				resource auth0_organization_member test_member {
					depends_on = [ auth0_user.user, auth0_organization.some_org, auth0_role.reader, auth0_role.admin ]
					organization_id = auth0_organization.some_org.id
					user_id = auth0_user.user.id
					roles = [ auth0_role.reader.id, auth0_role.admin.id ] // Adding an additional role
				}
			`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "organization_id", "auth0_organization.some_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "user_id", "auth0_user.user", "id"),
					resource.TestCheckResourceAttr("auth0_organization_member.test_member", "roles.#", "2"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.*", "auth0_role.reader", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.*", "auth0_role.admin", "id"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationMembersAux+`
				resource auth0_organization_member test_member {
					depends_on = [ auth0_user.user, auth0_organization.some_org, auth0_role.reader, auth0_role.admin ]
					organization_id = auth0_organization.some_org.id
					user_id = auth0_user.user.id
					roles = [ auth0_role.admin.id ] // Removing a role
				}
			`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "organization_id", "auth0_organization.some_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "user_id", "auth0_user.user", "id"),
					resource.TestCheckResourceAttr("auth0_organization_member.test_member", "roles.#", "1"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.*", "auth0_role.admin", "id"),
				),
			},
			{
				Config: template.ParseTestName(testAccOrganizationMembersAux+
					`
			resource auth0_organization_member test_member {
				depends_on = [ auth0_user.user, auth0_organization.some_org, auth0_role.reader, auth0_role.admin ]
				organization_id = auth0_organization.some_org.id
				user_id = auth0_user.user.id
				// Removing roles entirely
			}`, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_organization_member.test_member", "roles.#", "0"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "organization_id", "auth0_organization.some_org", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "user_id", "auth0_user.user", "id"),
				),
			},
		},
	})
}

const testAccOrganizationMembersAux = `
resource auth0_role reader {
	name = "Reader - {{.testName}}"
}

resource auth0_role admin {
	depends_on = [ auth0_role.reader ]
	name = "Admin - {{.testName}}"
}

resource auth0_user user {
	username = "testusername"
	email = "{{.testName}}@auth0.com"
	connection_name = "Username-Password-Authentication"
	email_verified = true
	password = "MyPass123$"
}

resource auth0_organization some_org {
	name = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`
