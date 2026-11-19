package clientprofile_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientProfileMinimal = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile"
  description = "Created by the acceptance tests."
}
`

const testAccClientProfileFull = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile"
  description = "Created by the acceptance tests."

  apply_policy {
    external_metadata_type       = "cimd"
    external_metadata_created_by = "client"
  }

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 2592000
      }
      idle_token_lifetime {
        default_value = 1296000
      }
      leeway {
        default_value = 0
      }
    }

    organization_usage {
      default_value = "require"
    }
    organization_require_behavior {
      default_value = "pre_login_prompt"
    }
    organization_discovery_methods {
      default_value = ["email", "organization_name"]
    }
    require_proof_of_possession {
      default_value = false
    }
  }
}
`

const testAccClientProfileUpdated = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile Updated"
  description = "Updated by the acceptance tests."

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 5184000
      }
      idle_token_lifetime {
        default_value = 2592000
      }
      leeway {
        default_value = 10
      }
    }

    organization_usage {
      default_value = "allow"
    }
    organization_require_behavior {
      default_value = "no_prompt"
    }
    require_proof_of_possession {
      default_value = true
    }
  }
}
`

const testAccClientProfileSingleLifetime = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile Updated"
  description = "Updated by the acceptance tests."

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 5184000
      }
    }
  }
}
`

const testAccClientProfileBlockRemovedDescriptionCleared = `
resource "auth0_client_profile" "my_profile" {
  name = "Acceptance Test Profile Updated"
}
`

const testAccClientProfileDuplicate = `
resource "auth0_client_profile" "my_profile" {
  name = "Acceptance Test Profile Updated"
}

resource "auth0_client_profile" "duplicate" {
  name = "Acceptance Test Duplicate Profile"
}
`

const testAccClientProfileInvalidLifetimes = `
resource "auth0_client_profile" "my_profile" {
  name = "Acceptance Test Profile"

  client_configuration {
    refresh_token {
      token_lifetime {
        default_value = 1000
      }
      idle_token_lifetime {
        default_value = 2000
      }
    }
  }
}
`

const testAccClientProfileInvalidDiscovery = `
resource "auth0_client_profile" "my_profile" {
  name = "Acceptance Test Profile"

  client_configuration {
    organization_require_behavior {
      default_value = "no_prompt"
    }
    organization_discovery_methods {
      default_value = ["email"]
    }
  }
}
`

func TestAccClientProfile(t *testing.T) {
	const name = "auth0_client_profile.my_profile"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccClientProfileInvalidLifetimes,
				ExpectError: regexp.MustCompile(
					"must not exceed",
				),
			},
			{
				Config: testAccClientProfileInvalidDiscovery,
				ExpectError: regexp.MustCompile(
					"requires\\s+`client_configuration.organization_require_behavior`",
				),
			},
			{
				Config: testAccClientProfileMinimal,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test Profile"),
					resource.TestCheckResourceAttr(name, "description", "Created by the acceptance tests."),
					resource.TestCheckResourceAttr(name, "apply_policy.0.external_metadata_type", "cimd"),
					resource.TestCheckResourceAttr(name, "apply_policy.0.external_metadata_created_by", "client"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.#", "0"),
					resource.TestCheckResourceAttrSet(name, "created_at"),
					resource.TestCheckResourceAttrSet(name, "updated_at"),
				),
			},
			{
				Config: testAccClientProfileFull,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.token_lifetime.0.default_value", "2592000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.idle_token_lifetime.0.default_value", "1296000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.leeway.0.default_value", "0"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_usage.0.default_value", "require"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_require_behavior.0.default_value", "pre_login_prompt"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_discovery_methods.0.default_value.#", "2"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.require_proof_of_possession.0.default_value", "false"),
				),
			},
			{
				Config: testAccClientProfileUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "name", "Acceptance Test Profile Updated"),
					resource.TestCheckResourceAttr(name, "description", "Updated by the acceptance tests."),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.token_lifetime.0.default_value", "5184000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.idle_token_lifetime.0.default_value", "2592000"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.leeway.0.default_value", "10"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_usage.0.default_value", "allow"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_require_behavior.0.default_value", "no_prompt"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_discovery_methods.#", "0"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.require_proof_of_possession.0.default_value", "true"),
				),
			},
			{
				Config: testAccClientProfileSingleLifetime,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.0.token_lifetime.0.default_value", "5184000"),
					resource.TestCheckResourceAttrSet(name, "client_configuration.0.refresh_token.0.idle_token_lifetime.0.default_value"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.organization_usage.#", "0"),
					resource.TestCheckResourceAttr(name, "client_configuration.0.require_proof_of_possession.#", "0"),
				),
			},
			{
				Config: testAccClientProfileBlockRemovedDescriptionCleared,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "description", ""),
					resource.TestCheckResourceAttr(name, "client_configuration.0.refresh_token.#", "0"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientProfileDuplicate,
				ExpectError: regexp.MustCompile(
					"a client profile already exists for this tenant",
				),
			},
		},
	})
}

const testAccClientProfileOIDCSupportScopes = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile OIDC"
  description = "Created by the acceptance tests."

  client_configuration {
    oidc_support {
      default_value {
        is_allowed     = true
        allowed_scopes = ["profile", "email"]
      }
    }
  }
}
`

const testAccClientProfileOIDCSupportAllScopes = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile OIDC"
  description = "Created by the acceptance tests."

  client_configuration {
    oidc_support {
      default_value {
        is_allowed       = true
        allow_all_scopes = true
      }
    }
  }
}
`

const testAccClientProfileOIDCSupportDisabled = `
resource "auth0_client_profile" "my_profile" {
  name        = "Acceptance Test Profile OIDC"
  description = "Created by the acceptance tests."

  client_configuration {
    oidc_support {
      default_value {
        is_allowed = false
      }
    }
  }
}
`

const testAccClientProfileOIDCSupportInvalid = `
resource "auth0_client_profile" "my_profile" {
  name = "Acceptance Test Profile OIDC"

  client_configuration {
    oidc_support {
      default_value {
        is_allowed       = true
        allowed_scopes   = ["profile"]
        allow_all_scopes = true
      }
    }
  }
}
`

func TestAccClientProfileOIDCSupport(t *testing.T) {
	const (
		name       = "auth0_client_profile.my_profile"
		defaultKey = "client_configuration.0.oidc_support.0.default_value.0."
	)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccClientProfileOIDCSupportInvalid,
				ExpectError: regexp.MustCompile(`mutually\s+exclusive`),
			},
			{
				Config: testAccClientProfileOIDCSupportScopes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, defaultKey+"is_allowed", "true"),
					resource.TestCheckResourceAttr(name, defaultKey+"allowed_scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, defaultKey+"allowed_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr(name, defaultKey+"allowed_scopes.*", "email"),
				),
			},
			{
				Config: testAccClientProfileOIDCSupportAllScopes,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, defaultKey+"is_allowed", "true"),
					resource.TestCheckResourceAttr(name, defaultKey+"allow_all_scopes", "true"),
					resource.TestCheckResourceAttr(name, defaultKey+"allowed_scopes.#", "0"),
				),
			},
			{
				Config: testAccClientProfileOIDCSupportDisabled,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, defaultKey+"is_allowed", "false"),
					resource.TestCheckResourceAttr(name, defaultKey+"allowed_scopes.#", "0"),
				),
			},
			{
				ResourceName:      name,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccClientProfileOIDCSupportScopes + `
data "auth0_client_profile" "my_profile" {
  id = auth0_client_profile.my_profile.id
}
`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_client_profile.my_profile", defaultKey+"is_allowed", "true"),
					resource.TestCheckResourceAttr("data.auth0_client_profile.my_profile", defaultKey+"allowed_scopes.#", "2"),
				),
			},
		},
	})
}
