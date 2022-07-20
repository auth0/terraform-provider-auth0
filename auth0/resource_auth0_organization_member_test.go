package auth0

import (
	"log"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/template"
)

func init() {
	resource.AddTestSweepers("auth0_organization_member", &resource.Sweeper{
		Name: "auth0_organization_member",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			var page int
			var result *multierror.Error
			for {
				organizationList, err := api.Organization.List(management.Page(page))
				if err != nil {
					result = multierror.Append(result, err)
					break
				}

				for _, organization := range organizationList.Organizations {
					log.Printf("[DEBUG] ➝ %s", organization.GetName())

					if strings.Contains(organization.GetName(), "test") {
						result = multierror.Append(
							result,
							api.Organization.Delete(organization.GetID()),
						)

						log.Printf("[DEBUG] ✗ %s", organization.GetName())
					}
				}
				if !organizationList.HasNext() {
					break
				}
				page++
			}

			users, err := api.User.List()
			if err != nil {
				return err
			}
			for _, user := range users.Users {
				if strings.Contains(*user.Email, "test") {
					err = api.User.Delete(*user.ID)
					result = multierror.Append(result, err)
				}
			}

			roles, err := api.Role.List()
			if err != nil {
				return err
			}
			for _, role := range roles.Roles {
				if strings.Contains(*role.Name, "test") {
					err = api.Role.Delete(*role.ID)
					result = multierror.Append(result, err)
				}
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccOrganizationMember(t *testing.T) {
	httpRecorder := configureHTTPRecorder(t)

	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviders(httpRecorder),
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
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.0", "auth0_role.reader", "id"),
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
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.0", "auth0_role.reader", "id"),
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.1", "auth0_role.admin", "id"),
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
					resource.TestCheckTypeSetElemAttrPair("auth0_organization_member.test_member", "roles.0", "auth0_role.admin", "id"),
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
	name = "Admin - {{.testName}}"
}

resource auth0_user user {
	email = "{{.testName}}@auth0.com"
	connection_name = "Username-Password-Authentication"
	email_verified = true
	password = "MyPass123$"
}

resource auth0_organization some_org{
	name = "some-org-{{.testName}}"
	display_name = "{{.testName}}"
}
`
