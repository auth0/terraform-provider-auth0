package role_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccRoleEmpty = `
resource "auth0_role" "the_one" {
	name = "The One - Acceptance Test - {{.testName}}"
}
`

const testAccRoleCreate = `
resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One - Acceptance Test"
}
`

const testAccRoleUpdate = `
resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One who will bring peace - Acceptance Test"
}
`

const testAccGivenAnOrganization = `
resource "auth0_organization" "zion" {
	name         = "test-org-role-{{.testName}}"
	display_name = "Zion - Acceptance Test - {{.testName}}"
}
`

const testAccRoleOrganizationLevel = testAccGivenAnOrganization + `
resource "auth0_role" "the_operator" {
	name        = "The Operator - Acceptance Test - {{.testName}}"
	description = "The Operator - Acceptance Test"
	type        = "organization"
	owner_id    = auth0_organization.zion.id
}
`

const testAccRoleOrganizationLevelUpdate = testAccGivenAnOrganization + `
resource "auth0_role" "the_operator" {
	name        = "The Operator - Acceptance Test - {{.testName}}"
	description = "The Operator who runs the ship - Acceptance Test"
	type        = "organization"
	owner_id    = auth0_organization.zion.id
}
`

const testAccRoleOrganizationLevelWithoutOwnerID = `
resource "auth0_role" "the_operator" {
	name = "The Operator - Acceptance Test - {{.testName}}"
	type = "organization"
}
`

const testAccRoleTenantLevelWithOwnerID = `
resource "auth0_role" "the_operator" {
	name     = "The Operator - Acceptance Test - {{.testName}}"
	type     = "tenant"
	owner_id = "org_XXXXXXXXXXXXXXXX"
}
`

const testAccRoleMutabilityWithoutType = `
resource "auth0_role" "mutability" {
	name = "The Architect - Acceptance Test - {{.testName}}"
}
`

const testAccRoleMutabilityWithExplicitTenantType = `
resource "auth0_role" "mutability" {
	name = "The Architect - Acceptance Test - {{.testName}}"
	type = "tenant"
}
`

const testAccGivenAnOrganizationForMutability = `
resource "auth0_organization" "zion" {
	name         = "org-{{.testName}}"
	display_name = "Zion - Acceptance Test - {{.testName}}"
}
`

const testAccRoleMutabilityFlippedToOrganization = testAccGivenAnOrganizationForMutability + `
resource "auth0_role" "mutability" {
	name     = "The Architect - Acceptance Test - {{.testName}}"
	type     = "organization"
	owner_id = auth0_organization.zion.id
}
`

const testAccRoleMutabilityWithFieldsRemoved = testAccGivenAnOrganizationForMutability + `
resource "auth0_role" "mutability" {
	name = "The Architect - Acceptance Test - {{.testName}}"
}
`

const testAccRoleMutabilityFlippedBackToTenant = testAccGivenAnOrganizationForMutability + `
resource "auth0_role" "mutability" {
	name = "The Architect - Acceptance Test - {{.testName}}"
	type = "tenant"
}
`

func TestAccRole(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccRoleEmpty, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One - Acceptance Test"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_one", "description", "The One who will bring peace - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "type", "tenant"),
					resource.TestCheckResourceAttr("auth0_role.the_one", "owner_id", ""),
				),
			},
		},
	})
}

func TestAccRoleOrganizationLevel(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testAccRoleOrganizationLevelWithoutOwnerID, testName),
				ExpectError: regexp.MustCompile("owner_id is required when type is set to organization"),
			},
			{
				Config:      acctest.ParseTestName(testAccRoleTenantLevelWithOwnerID, testName),
				ExpectError: regexp.MustCompile("owner_id can only be set when type is set to organization"),
			},
			{
				Config: acctest.ParseTestName(testAccRoleOrganizationLevel, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_operator", "name", fmt.Sprintf("The Operator - Acceptance Test - %s", testName)),
					resource.TestCheckResourceAttr("auth0_role.the_operator", "description", "The Operator - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_operator", "type", "organization"),
					resource.TestCheckResourceAttrPair("auth0_role.the_operator", "owner_id", "auth0_organization.zion", "id"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccRoleOrganizationLevelUpdate, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.the_operator", "description", "The Operator who runs the ship - Acceptance Test"),
					resource.TestCheckResourceAttr("auth0_role.the_operator", "type", "organization"),
					resource.TestCheckResourceAttrPair("auth0_role.the_operator", "owner_id", "auth0_organization.zion", "id"),
				),
			},
			{
				Config:            acctest.ParseTestName(testAccRoleOrganizationLevelUpdate, testName),
				ResourceName:      "auth0_role.the_operator",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return state.RootModule().Resources["auth0_role.the_operator"].Primary.ID, nil
				},
			},
		},
	})
}

// The roles endpoint only accepts type and owner_id when the role is created, so
// changing either has to replace the role instead of updating it in place. Roles
// created before these fields existed keep working, as both are computed.
func TestAccRoleTypeAndOwnerIDAreCreateOnly(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				// How a role provisioned by an older provider version looks.
				Config: acctest.ParseTestName(testAccRoleMutabilityWithoutType, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.mutability", "type", "tenant"),
					resource.TestCheckResourceAttr("auth0_role.mutability", "owner_id", ""),
				),
			},
			{
				// Adding the type the API already returned is not a change.
				Config:   acctest.ParseTestName(testAccRoleMutabilityWithExplicitTenantType, testName),
				PlanOnly: true,
			},
			{
				// Switching from tenant to organization replaces the role.
				Config: acctest.ParseTestName(testAccRoleMutabilityFlippedToOrganization, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_role.mutability", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.mutability", "type", "organization"),
					resource.TestCheckResourceAttrPair("auth0_role.mutability", "owner_id", "auth0_organization.zion", "id"),
				),
			},
			{
				// Dropping both fields does not plan a replacement back to tenant.
				Config:   acctest.ParseTestName(testAccRoleMutabilityWithFieldsRemoved, testName),
				PlanOnly: true,
			},
			{
				// And the other way around.
				Config: acctest.ParseTestName(testAccRoleMutabilityFlippedBackToTenant, testName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_role.mutability", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_role.mutability", "type", "tenant"),
					resource.TestCheckResourceAttr("auth0_role.mutability", "owner_id", ""),
				),
			},
		},
	})
}
