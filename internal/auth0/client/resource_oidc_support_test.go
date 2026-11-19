package client_test

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccClientOIDCSupportNoBlock = `
resource "auth0_client" "my_client" {
	name                      = "Acceptance Test - OIDC Support - {{.testName}}"
	app_type                  = "regular_web"
	is_first_party            = false
	third_party_security_mode = "strict"
}
`

const testAccClientOIDCSupportScopes = `
resource "auth0_client" "my_client" {
	name                      = "Acceptance Test - OIDC Support - {{.testName}}"
	app_type                  = "regular_web"
	is_first_party            = false
	third_party_security_mode = "strict"

	oidc_support {
		is_allowed     = true
		allowed_scopes = ["profile", "email"]
	}
}
`

const testAccClientOIDCSupportScopesUpdated = `
resource "auth0_client" "my_client" {
	name                      = "Acceptance Test - OIDC Support - {{.testName}}"
	app_type                  = "regular_web"
	is_first_party            = false
	third_party_security_mode = "strict"

	oidc_support {
		is_allowed     = true
		allowed_scopes = ["address", "phone"]
	}
}
`

const testAccClientOIDCSupportAllScopes = `
resource "auth0_client" "my_client" {
	name                      = "Acceptance Test - OIDC Support - {{.testName}}"
	app_type                  = "regular_web"
	is_first_party            = false
	third_party_security_mode = "strict"

	oidc_support {
		is_allowed       = true
		allow_all_scopes = true
	}
}
`

const testAccClientOIDCSupportDisabled = `
resource "auth0_client" "my_client" {
	name                      = "Acceptance Test - OIDC Support - {{.testName}}"
	app_type                  = "regular_web"
	is_first_party            = false
	third_party_security_mode = "strict"

	oidc_support {
		is_allowed = false
	}
}
`

const testAccClientOIDCSupportDataSources = testAccClientOIDCSupportScopes + `
data "auth0_client" "my_client" {
	client_id = auth0_client.my_client.client_id
}
`

func TestAccClientOIDCSupport(t *testing.T) {
	const name = "auth0_client.my_client"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			// The invalid configurations are rejected at plan time, so they make no write calls.
			{
				Config:      acctest.ParseTestName(testAccClientOIDCSupportInvalidBothScopeFields, t.Name()),
				ExpectError: regexp.MustCompile(`mutually\s+exclusive`),
			},
			{
				Config:      acctest.ParseTestName(testAccClientOIDCSupportInvalidOpenID, t.Name()),
				ExpectError: regexp.MustCompile(`to\s+be\s+one\s+of`),
			},
			{
				Config:      acctest.ParseTestName(testAccClientOIDCSupportInvalidDisabledWithScopes, t.Name()),
				ExpectError: regexp.MustCompile(`cannot\s+be\s+set\s+when`),
			},
			{
				Config:      acctest.ParseTestName(testAccClientOIDCSupportInvalidNonConformant, t.Name()),
				ExpectError: regexp.MustCompile(`oidc_conformant`),
			},
			{
				// A client without the block reads back a server-materialized default and plans no diff.
				Config: acctest.ParseTestName(testAccClientOIDCSupportNoBlock, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_conformant", "true"),
					resource.TestCheckResourceAttr(name, "oidc_support.#", "1"),
				),
			},
			{
				Config:   acctest.ParseTestName(testAccClientOIDCSupportNoBlock, t.Name()),
				PlanOnly: true,
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_support.#", "1"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.is_allowed", "true"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allowed_scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, "oidc_support.0.allowed_scopes.*", "profile"),
					resource.TestCheckTypeSetElemAttr(name, "oidc_support.0.allowed_scopes.*", "email"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allow_all_scopes", "false"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportScopesUpdated, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_support.0.allowed_scopes.#", "2"),
					resource.TestCheckTypeSetElemAttr(name, "oidc_support.0.allowed_scopes.*", "address"),
					resource.TestCheckTypeSetElemAttr(name, "oidc_support.0.allowed_scopes.*", "phone"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportAllScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_support.0.is_allowed", "true"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allow_all_scopes", "true"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allowed_scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportDisabled, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_support.0.is_allowed", "false"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allow_all_scopes", "false"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allowed_scopes.#", "0"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportScopes, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "oidc_support.0.is_allowed", "true"),
					resource.TestCheckResourceAttr(name, "oidc_support.0.allowed_scopes.#", "2"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccClientOIDCSupportDataSources, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.auth0_client.my_client", "oidc_support.0.is_allowed", "true"),
					resource.TestCheckResourceAttr("data.auth0_client.my_client", "oidc_support.0.allowed_scopes.#", "2"),
				),
			},
		},
	})
}

const testAccClientOIDCSupportInvalidBothScopeFields = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - OIDC Support - {{.testName}}"

	oidc_support {
		is_allowed       = true
		allowed_scopes   = ["profile"]
		allow_all_scopes = true
	}
}
`

const testAccClientOIDCSupportInvalidOpenID = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - OIDC Support - {{.testName}}"

	oidc_support {
		is_allowed     = true
		allowed_scopes = ["openid"]
	}
}
`

const testAccClientOIDCSupportInvalidDisabledWithScopes = `
resource "auth0_client" "my_client" {
	name = "Acceptance Test - OIDC Support - {{.testName}}"

	oidc_support {
		is_allowed     = false
		allowed_scopes = ["profile"]
	}
}
`

const testAccClientOIDCSupportInvalidNonConformant = `
resource "auth0_client" "my_client" {
	name            = "Acceptance Test - OIDC Support - {{.testName}}"
	oidc_conformant = false

	oidc_support {
		is_allowed       = true
		allow_all_scopes = true
	}
}
`
