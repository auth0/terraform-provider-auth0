package provider

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func TestAccOrganizationMember(t *testing.T) {
	httpRecorder := recorder.New(t)

	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories: ProviderTestFactories(httpRecorder),
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

func TestImportOrganizationMember(t *testing.T) {
	var testCases = []struct {
		testName               string
		givenID                string
		expectedOrganizationID string
		expectedUserID         string
		expectedError          error
	}{
		{
			testName:               "it correctly parses the resource ID",
			givenID:                "org_1234:auth0|62d82",
			expectedOrganizationID: "org_1234",
			expectedUserID:         "auth0|62d82",
		},
		{
			testName:      "it fails when the given ID is empty",
			givenID:       "",
			expectedError: fmt.Errorf("ID cannot be empty"),
		},
		{
			testName:      "it fails when the given ID does not have \":\" as a separator",
			givenID:       "org_1234auth0|62d82",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<userID>"),
		},
		{
			testName:      "it fails when the given ID has too many separators",
			givenID:       "org_1234:auth0|62d82:",
			expectedError: fmt.Errorf("ID must be formated as <organizationID>:<userID>"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			data := schema.TestResourceDataRaw(t, newOrganizationMember().Schema, nil)
			data.SetId(testCase.givenID)

			actualData, err := importOrganizationMember(context.Background(), data, nil)

			if testCase.expectedError != nil {
				assert.EqualError(t, err, testCase.expectedError.Error())
				assert.Nil(t, actualData)
				return
			}

			assert.Equal(t, actualData[0].Get("organization_id").(string), testCase.expectedOrganizationID)
			assert.Equal(t, actualData[0].Get("user_id").(string), testCase.expectedUserID)
			assert.NotEqual(t, actualData[0].Id(), testCase.givenID)
		})
	}
}
