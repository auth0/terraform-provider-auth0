package role_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
	"github.com/auth0/terraform-provider-auth0/internal/recorder"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

const testAccGivenAResourceServer = `
resource "auth0_resource_server" "matrix" {
    name       = "Role - Acceptance Test - {{.testName}}"
    identifier = "https://{{.testName}}.matrix.com/"

    scopes {
        value = "stop:bullets"
        description = "Stop bullets"
    }

    scopes {
        value = "bring:peace"
        description = "Bring peace"
    }
}

resource "auth0_role" "the_one" {
	name        = "The One - Acceptance Test - {{.testName}}"
	description = "The One - Acceptance Test"

	permissions {
		name = "stop:bullets"
		resource_server_identifier = auth0_resource_server.matrix.identifier
	}
	permissions {
		name = "bring:peace"
		resource_server_identifier = auth0_resource_server.matrix.identifier
	}
}
`

const testAccDataSourceRoleByName = testAccGivenAResourceServer + `
data "auth0_role" "test" {
	name = auth0_role.the_one.name
}
`

const testAccDataSourceRoleByID = testAccGivenAResourceServer + `
data "auth0_role" "test" {
	role_id = auth0_role.the_one.id
}
`

func TestAccDataSourceRoleRequiredArguments(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: provider.TestFactories(nil),
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_role" "test" { }`,
				ExpectError: regexp.MustCompile("one of `name,role_id` must be specified"),
			},
		},
	})
}

func TestAccDataSourceRoleByName(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccDataSourceRoleByName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.auth0_role.test", "role_id"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "name", fmt.Sprintf("The One - Acceptance Test - %s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_role.test", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "permissions.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceRoleByID(t *testing.T) {
	httpRecorder := recorder.New(t)
	testName := strings.ToLower(t.Name())

	resource.Test(t, resource.TestCase{
		ProviderFactories:         provider.TestFactories(httpRecorder),
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccDataSourceRoleByID, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_role.test", "role_id"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "name", fmt.Sprintf("The One - Acceptance Test - %s", strings.ToLower(t.Name()))),
					resource.TestCheckResourceAttr("data.auth0_role.test", "description", "The One - Acceptance Test"),
					resource.TestCheckResourceAttr("data.auth0_role.test", "permissions.#", "2"),
				),
			},
		},
	})
}
