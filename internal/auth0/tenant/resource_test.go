package tenant_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

func TestAccTenant(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccEmptyTenant,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "168"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
				),
			},
			{
				Config: testAccTenantConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_audience", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_directory", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "friendly_name", "My Test Tenant"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "picture_url", "https://mycompany.org/logo.png"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "support_email", "support@mycompany.org"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "support_url", "https://mycompany.org/support"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.0", "https://mycompany.org/logoutCallback"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "720"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sandbox_version", "16"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.#", "3"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.0", "en"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.1", "de"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.2", "fr"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.disable_clickjack_protection_headers", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_public_signup_user_exists_error", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.use_scope_descriptions_for_consent", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.mfa_show_factor_list_on_enrollment", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", "https://example.com/login"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "non-persistent"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sessions.0.oidc_logout_prompt_enabled", "false"),
				),
			},
			{
				Config: testAccTenantConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.0", "de"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.1", "fr"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_audience", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.disable_clickjack_protection_headers", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_public_signup_user_exists_error", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.use_scope_descriptions_for_consent", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.mfa_show_factor_list_on_enrollment", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "persistent"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sessions.0.oidc_logout_prompt_enabled", "true"),
				),
			},
			{
				Config: testAccEmptyTenant,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.0", "de"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.1", "fr"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_audience", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.disable_clickjack_protection_headers", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_public_signup_user_exists_error", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.use_scope_descriptions_for_consent", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "persistent"),
				),
			},
		},
	})
}

const testAccTenantConfigCreate = `
resource "auth0_tenant" "my_tenant" {
	default_directory       = ""
	default_audience        = ""
	friendly_name           = "My Test Tenant"
	picture_url             = "https://mycompany.org/logo.png"
	support_email           = "support@mycompany.org"
	support_url             = "https://mycompany.org/support"
	default_redirection_uri = "https://example.com/login"
	allowed_logout_urls     = [ "https://mycompany.org/logoutCallback" ]
	session_lifetime        = 720
	sandbox_version         = "16"
	idle_session_lifetime   = 72
	enabled_locales         = ["en", "de", "fr"]

	flags {
		disable_clickjack_protection_headers   = true
		enable_public_signup_user_exists_error = true
		use_scope_descriptions_for_consent     = true
		no_disclose_enterprise_connections     = false
		disable_management_api_sms_obfuscation = false
		disable_fields_map_fix                 = false
		mfa_show_factor_list_on_enrollment     = false
	}

	session_cookie {
		mode = "non-persistent"
	}

	sessions {
		oidc_logout_prompt_enabled = false
	}
}
`

const testAccTenantConfigUpdate = `
resource "auth0_tenant" "my_tenant" {
	default_directory       = ""
	default_redirection_uri = ""
	friendly_name           = "My Test Tenant"
	picture_url             = "https://mycompany.org/logo.png"
	support_email           = "support@mycompany.org"
	support_url             = "https://mycompany.org/support"
	allowed_logout_urls     = []
	session_lifetime        = 720
	sandbox_version         = "16"
	idle_session_lifetime   = 72
	enabled_locales         = ["de", "fr"]

	flags {
		enable_public_signup_user_exists_error = true
		disable_clickjack_protection_headers   = false # <---- disable and test
		use_scope_descriptions_for_consent     = false
		no_disclose_enterprise_connections     = false
		disable_management_api_sms_obfuscation = true
		disable_fields_map_fix                 = true
		mfa_show_factor_list_on_enrollment     = true
	}

	session_cookie {
		mode = "persistent"
	}

	sessions {
		oidc_logout_prompt_enabled = true
	}
}
`

const testAccEmptyTenant = `resource "auth0_tenant" "my_tenant" {}`

func TestAccTenantDefaults(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// Only run with recorded HTTP requests because  normal E2E tests will naturally configure the tenant
		// and this test will only pass when the tenant has not been configured yet (aka "fresh" tenants).
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:        testAccEmptyTenant,
				ImportState:   true,
				ImportStateId: "some-arbitrary-identifier",
				ResourceName:  "auth0_tenant.my_tenant",
			},
			{
				Config: testAccEmptyTenant,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "168"),     // Auth0 default.
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"), // Auth0 default.
				),
			},
		},
	})
}
