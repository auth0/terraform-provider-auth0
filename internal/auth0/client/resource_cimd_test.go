package client_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientCIMDCreate = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
}
`

const testAccClientCIMDWithEditableFields = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
    description        = "CIMD test client"
    app_type           = "native"
    allowed_origins    = ["https://example.com"]
    web_origins        = ["https://example.com"]
    grant_types        = ["authorization_code", "refresh_token"]
    oidc_conformant    = true

    jwt_configuration {
        lifetime_in_seconds = 7200
        alg                 = "RS256"
    }

    refresh_token {
        rotation_type               = "rotating"
        expiration_type             = "expiring"
        leeway                      = 30
        token_lifetime              = 2592000
        infinite_token_lifetime     = false
        idle_token_lifetime         = 1296000
        infinite_idle_token_lifetime = false
    }
}
`

const testAccClientCIMDUpdate = `
resource "auth0_client_cimd" "test" {
    external_client_id = "https://tinywiki.xyz/client.json"
    description        = "Updated CIMD test client"
    app_type           = "native"
    allowed_origins    = ["https://example.com", "https://other.com"]
    web_origins        = ["https://example.com"]
    grant_types        = ["authorization_code"]
    oidc_conformant    = true

    jwt_configuration {
        lifetime_in_seconds = 3600
        alg                 = "RS256"
    }

    refresh_token {
        rotation_type               = "non-rotating"
        expiration_type             = "expiring"
        leeway                      = 0
        token_lifetime              = 1296000
        infinite_token_lifetime     = false
        idle_token_lifetime         = 648000
        infinite_idle_token_lifetime = false
    }
}
`

func TestAccClientCIMD(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccClientCIMDCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "client_id"),
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "name"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_client_id", "https://tinywiki.xyz/client.json"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_created_by", "admin"),
				),
			},
			{
				Config: testAccClientCIMDWithEditableFields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "description", "CIMD test client"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "app_type", "native"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.0", "https://example.com"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "web_origins.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "grant_types.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "oidc_conformant", "true"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "is_first_party", "false"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "jwt_configuration.0.lifetime_in_seconds", "7200"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "jwt_configuration.0.alg", "RS256"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "refresh_token.0.rotation_type", "rotating"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "refresh_token.0.expiration_type", "expiring"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "refresh_token.0.leeway", "30"),
					resource.TestCheckResourceAttrSet("auth0_client_cimd.test", "name"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
				),
			},
			{
				Config: testAccClientCIMDUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "description", "Updated CIMD test client"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "app_type", "native"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "allowed_origins.#", "2"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "grant_types.#", "1"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "jwt_configuration.0.lifetime_in_seconds", "3600"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "refresh_token.0.rotation_type", "non-rotating"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "refresh_token.0.leeway", "0"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr("auth0_client_cimd.test", "external_metadata_created_by", "admin"),
				),
			},
			{
				ResourceName:            "auth0_client_cimd.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"external_client_id_version"},
			},
		},
	})
}
