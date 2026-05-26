package user_test

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceUserConnectedAccountsEmpty = `
resource "auth0_user" "user" {
	connection_name = "Username-Password-Authentication"
	user_id         = "{{.testName}}"
	username        = "{{.testName}}"
	password        = "passpass$12$12"
	email           = "{{.testName}}@acceptance.test.com"
}

data "auth0_user_connected_accounts" "test" {
	depends_on = [auth0_user.user]
	user_id = auth0_user.user.id
}
`

const testAccDataSourceUserConnectedAccountsTokenVault = `
data "auth0_user_connected_accounts" "test" {
	user_id = "{{.userID}}"
}
`

func TestAccDataSourceUserConnectedAccounts(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceUserConnectedAccountsEmpty, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_user_connected_accounts.test", "connected_accounts.#", "0"),
				),
			},
		},
	})
}

func TestAccDataSourceUserConnectedAccountsTokenVault(t *testing.T) {
	if os.Getenv("AUTH0_ENABLE_TOKEN_VAULT_TESTS") != "1" {
		t.Skip("AUTH0_ENABLE_TOKEN_VAULT_TESTS is not set to 1")
	}

	userID := os.Getenv("AUTH0_TOKEN_VAULT_USER_ID")
	if userID == "" {
		t.Fatal("AUTH0_TOKEN_VAULT_USER_ID must be set when AUTH0_ENABLE_TOKEN_VAULT_TESTS=1")
	}

	expectedOrgID := os.Getenv("AUTH0_TOKEN_VAULT_ORG_ID")
	if expectedOrgID == "" {
		t.Fatal("AUTH0_TOKEN_VAULT_ORG_ID must be set when AUTH0_ENABLE_TOKEN_VAULT_TESTS=1")
	}

	config := acctest.ParseParametersInTemplate(
		testAccDataSourceUserConnectedAccountsTokenVault,
		map[string]interface{}{"userID": userID},
	)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.#"),
					resource.TestCheckResourceAttr("data.auth0_user_connected_accounts.test", "connected_accounts.0.organization_id", expectedOrgID),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.id"),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.connection"),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.connection_id"),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.strategy"),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.access_type"),
					resource.TestCheckResourceAttrSet("data.auth0_user_connected_accounts.test", "connected_accounts.0.created_at"),
					resource.TestCheckResourceAttr("data.auth0_user_connected_accounts.test", "user_id", userID),
				),
			},
		},
	})
}
