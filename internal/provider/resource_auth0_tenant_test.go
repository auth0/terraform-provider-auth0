package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/recorder"
)

func TestAccTenant(t *testing.T) {
	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: ProviderTestFactories(httpRecorder),
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
					resource.TestCheckResourceAttrSet("auth0_tenant.my_tenant", "change_password.0.enabled"),
					resource.TestCheckResourceAttrSet("auth0_tenant.my_tenant", "guardian_mfa_page.0.enabled"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_audience", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_directory", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.html", "<html>Error Page</html>"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.show_log_link", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.url", "https://mycompany.org/error"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "friendly_name", "My Test Tenant"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "picture_url", "https://mycompany.org/logo.png"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "support_email", "support@mycompany.org"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "support_url", "https://mycompany.org/support"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.0", "https://mycompany.org/logoutCallback"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "720"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sandbox_version", "12"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.0", "en"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.1", "de"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.2", "fr"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.universal_login", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.disable_clickjack_protection_headers", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_public_signup_user_exists_error", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.use_scope_descriptions_for_consent", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "universal_login.0.colors.0.primary", "#0059d6"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "universal_login.0.colors.0.page_background", "#000000"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", "https://example.com/login"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "non-persistent"),
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
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "persistent"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", ""),
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
	default_directory = ""
	default_audience = ""
	error_page {
		html = "<html>Error Page</html>"
		show_log_link = false
		url = "https://mycompany.org/error"
	}
	friendly_name = "My Test Tenant"
	picture_url = "https://mycompany.org/logo.png"
	support_email = "support@mycompany.org"
	support_url = "https://mycompany.org/support"
	allowed_logout_urls = [
		"https://mycompany.org/logoutCallback"
	]
	session_lifetime = 720
	sandbox_version = "12"
	idle_session_lifetime = 72
	enabled_locales = ["en", "de", "fr"]
	flags {
		universal_login = true
		disable_clickjack_protection_headers = true
		enable_public_signup_user_exists_error = true
		use_scope_descriptions_for_consent = true
		no_disclose_enterprise_connections = false
		disable_management_api_sms_obfuscation = false
		disable_fields_map_fix = false
	}
	universal_login {
		colors {
			primary = "#0059d6"
			page_background = "#000000"
		}
	}
	default_redirection_uri = "https://example.com/login"
	session_cookie {
		mode = "non-persistent"
	}
}
`

const testAccTenantConfigUpdate = `
resource "auth0_tenant" "my_tenant" {
	change_password {
		enabled = true
		html = "<html>Change Password</html>"
	}
	guardian_mfa_page {
		enabled = true
		html = "<html>MFA</html>"
	}
	default_directory = ""
	error_page {
		html = "<html>Error Page</html>"
		show_log_link = false
		url = "https://mycompany.org/error"
	}
	friendly_name = "My Test Tenant"
	picture_url = "https://mycompany.org/logo.png"
	support_email = "support@mycompany.org"
	support_url = "https://mycompany.org/support"
	allowed_logout_urls = []
	session_lifetime = 720
	sandbox_version = "12"
	idle_session_lifetime = 72
	enabled_locales = ["de", "fr"]
	flags {
		universal_login = true
		enable_public_signup_user_exists_error = true
		disable_clickjack_protection_headers = false # <---- disable and test
		use_scope_descriptions_for_consent = false   #
		no_disclose_enterprise_connections = false
		disable_management_api_sms_obfuscation = true
		disable_fields_map_fix = true
	}
	universal_login {
		colors {
			primary = "#0059d6"
			page_background = "#000000"
		}
	}
	default_redirection_uri = ""
	session_cookie {
		mode = "persistent"
	}
}
`

const testAccEmptyTenant = `resource "auth0_tenant" "my_tenant" {}`

func TestAccTenantDefaults(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != recorder.RecordingsDomain {
		// Only run with recorded HTTP requests because  normal E2E tests will naturally configure the tenant
		// and this test will only pass when the tenant has not been configured yet (aka "fresh" tenants).
		t.Skip()
	}

	httpRecorder := recorder.New(t)

	resource.Test(t, resource.TestCase{
		ProviderFactories: ProviderTestFactories(httpRecorder),
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
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "168"),     // Auth0 default
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"), // Auth0 default
				),
			},
		},
	})
}
