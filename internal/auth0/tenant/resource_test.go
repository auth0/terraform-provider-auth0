package tenant_test

import (
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccEmptyTenant = `resource "auth0_tenant" "my_tenant" {}`

const testAccTenantEnableSSOConfigCreate = `
resource "auth0_tenant" "my_tenant" {
	flags {
		enable_sso = false
	}
}
`

const testAccTenantEnableSSOConfigUpdate = `
resource "auth0_tenant" "my_tenant" {
	flags {
		enable_sso = true
	}
}
`

const testAccTenantConfigCreate = `
resource "auth0_tenant" "my_tenant" {
	default_directory                             = ""
	default_audience                              = ""
	friendly_name                                 = "My Test Tenant"
	picture_url                                   = "https://mycompany.org/logo.png"
	support_email                                 = "support@mycompany.org"
	support_url                                   = "https://mycompany.org/support"
	default_redirection_uri                       = "https://example.com/login"
	allowed_logout_urls                           = [ "https://mycompany.org/logoutCallback" ]
	session_lifetime                              = 720
	sandbox_version                               = "18"
	idle_session_lifetime                         = 72
	ephemeral_session_lifetime      			  = 48
	idle_ephemeral_session_lifetime 			  = 36
	enabled_locales                               = ["en", "de", "fr"]
	disable_acr_values_supported                  = true

	allow_organization_name_in_authentication_api = false
	customize_mfa_in_postlogin_action             = false
	pushed_authorization_requests_supported       = true

	flags {
		disable_clickjack_protection_headers   = true
		enable_public_signup_user_exists_error = true
		use_scope_descriptions_for_consent     = true
		remove_alg_from_jwks                   = true
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

	mtls {
		enable_endpoint_aliases = true
	}

	oidc_logout {
		rp_logout_end_session_endpoint_discovery = true
	}
}
`

const testAccTenantConfigInvalidACRValuesSupported = `
resource "auth0_tenant" "my_tenant" {
	default_directory                             = ""
	default_audience                              = ""
	friendly_name                                 = "My Test Tenant"
	picture_url                                   = "https://mycompany.org/logo.png"
	support_email                                 = "support@mycompany.org"
	support_url                                   = "https://mycompany.org/support"
	default_redirection_uri                       = "https://example.com/login"
	allowed_logout_urls                           = [ "https://mycompany.org/logoutCallback" ]
	session_lifetime                              = 720
	sandbox_version                               = "16"
	idle_session_lifetime                         = 72
	enabled_locales                               = ["en", "de", "fr"]
	disable_acr_values_supported                  = true
	acr_values_supported                          = ["foo", "bar"]

	allow_organization_name_in_authentication_api = false
	customize_mfa_in_postlogin_action             = false
	pushed_authorization_requests_supported       = true

	flags {
		disable_clickjack_protection_headers   = true
		enable_public_signup_user_exists_error = true
		use_scope_descriptions_for_consent     = true
		remove_alg_from_jwks                   = true
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

const testAccTenantConfigInvalidMTLS = `
resource "auth0_tenant" "my_tenant" {
	default_directory                             = ""
	default_audience                              = ""
	friendly_name                                 = "My Test Tenant"
	picture_url                                   = "https://mycompany.org/logo.png"
	support_email                                 = "support@mycompany.org"
	support_url                                   = "https://mycompany.org/support"
	default_redirection_uri                       = "https://example.com/login"
	allowed_logout_urls                           = [ "https://mycompany.org/logoutCallback" ]
	session_lifetime                              = 720
	sandbox_version                               = "16"
	idle_session_lifetime                         = 72
	enabled_locales                               = ["en", "de", "fr"]
	disable_acr_values_supported                  = true

	allow_organization_name_in_authentication_api = false
	customize_mfa_in_postlogin_action             = false
	pushed_authorization_requests_supported       = true

	flags {
		disable_clickjack_protection_headers   = true
		enable_public_signup_user_exists_error = true
		use_scope_descriptions_for_consent     = true
		remove_alg_from_jwks                   = true
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

	mtls {
		enable_endpoint_aliases = true
		disable                 = true
	}
}
`

const testAccTenantConfigUpdate = `
resource "auth0_tenant" "my_tenant" {
	default_directory                             = ""
	default_redirection_uri                       = ""
	friendly_name                                 = "My Test Tenant"
	picture_url                                   = "https://mycompany.org/logo.png"
	support_email                                 = "support@mycompany.org"
	support_url                                   = "https://mycompany.org/support"
	allowed_logout_urls                           = []
	session_lifetime                              = 720
	sandbox_version                               = "18"
	idle_session_lifetime                         = 72
	enabled_locales                               = ["de", "fr"]

	allow_organization_name_in_authentication_api = true
	customize_mfa_in_postlogin_action             = true
	pushed_authorization_requests_supported       = false

	flags {
		enable_public_signup_user_exists_error = true
		disable_clickjack_protection_headers   = false # <---- disable and test
		use_scope_descriptions_for_consent     = false
		no_disclose_enterprise_connections     = false
		remove_alg_from_jwks                   = false
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

	acr_values_supported                     = ["foo", "bar"]
	mtls {
		disable = true
	}

	error_page {
		html          = "<html></html>"
		show_log_link = false
		url           = "https://mycompany.org/error"
	}

	oidc_logout {
		rp_logout_end_session_endpoint_discovery = false
	}
}
`

const testAccTenantConfigUpdateBack = `
resource "auth0_tenant" "my_tenant" {
	default_directory                             = ""
	default_redirection_uri                       = ""
	friendly_name                                 = "My Test Tenant"
	picture_url                                   = "https://mycompany.org/logo.png"
	support_email                                 = "support@mycompany.org"
	support_url                                   = "https://mycompany.org/support"
	allowed_logout_urls                           = []
	session_lifetime                              = 720
	sandbox_version                               = "18"
	idle_session_lifetime                         = 72

	allow_organization_name_in_authentication_api = true
	customize_mfa_in_postlogin_action             = true
	pushed_authorization_requests_supported       = false

	flags {
		enable_public_signup_user_exists_error = true
		disable_clickjack_protection_headers   = false # <---- disable and test
		use_scope_descriptions_for_consent     = false
		no_disclose_enterprise_connections     = false
		remove_alg_from_jwks                   = false
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


	acr_values_supported  = []
    mtls {
		enable_endpoint_aliases = true
	}
}
`

const testAccTenantWithDefaultTokenQuota = `
resource "auth0_tenant" "my_tenant" {
	friendly_name = "My Test Tenant with Token Quota"
	default_token_quota {
		clients {
			client_credentials {
				enforce = true
				per_hour = 100
				per_day = 2000
			}
		}
		organizations {
			client_credentials {
				enforce = true
				per_hour = 200
				per_day = 4000
			}
		}
	}
}
`

const testAccTenantWithDefaultTokenQuotaUpdated = `
resource "auth0_tenant" "my_tenant" {
	friendly_name = "My Test Tenant with Token Quota"
	default_token_quota {
		clients {
			client_credentials {
				enforce = false
				per_hour = 50
				per_day = 1000
			}
		}
		organizations {
			client_credentials {
				enforce = false
				per_hour = 150
				per_day = 3000
			}
		}
	}
}
`

const testAccTenantWithDefaultTokenQuotaRemoved = `
resource "auth0_tenant" "my_tenant" {
	friendly_name = "My Test Tenant with Token Quota"
}
`

func TestAccTenant_Main(t *testing.T) {
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
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sandbox_version", "18"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "ephemeral_session_lifetime", "48"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_ephemeral_session_lifetime", "36"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.#", "3"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.0", "en"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.1", "de"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "enabled_locales.2", "fr"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.disable_clickjack_protection_headers", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_public_signup_user_exists_error", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.use_scope_descriptions_for_consent", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.remove_alg_from_jwks", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.mfa_show_factor_list_on_enrollment", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", "https://example.com/login"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "non-persistent"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sessions.0.oidc_logout_prompt_enabled", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allow_organization_name_in_authentication_api", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "customize_mfa_in_postlogin_action", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "pushed_authorization_requests_supported", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "disable_acr_values_supported", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "acr_values_supported.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.0.enable_endpoint_aliases", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.0.disable", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "oidc_logout.0.rp_logout_end_session_endpoint_discovery", "true"),
				),
			},
			{
				Config:      acctest.ParseTestName(testAccTenantConfigInvalidACRValuesSupported, t.Name()),
				ExpectError: regexp.MustCompile(`only one of disable_acr_values_supported and acr_values_supported should be set`),
			},
			{
				Config:      acctest.ParseTestName(testAccTenantConfigInvalidMTLS, t.Name()),
				ExpectError: regexp.MustCompile(`only one of disable and enable_endpoint_aliases should be set in the mtls block`),
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
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.remove_alg_from_jwks", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.mfa_show_factor_list_on_enrollment", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allowed_logout_urls.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_cookie.0.mode", "persistent"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_redirection_uri", ""),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "sessions.0.oidc_logout_prompt_enabled", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "allow_organization_name_in_authentication_api", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "customize_mfa_in_postlogin_action", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "pushed_authorization_requests_supported", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "disable_acr_values_supported", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "acr_values_supported.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_tenant.my_tenant", "acr_values_supported.*", "foo"),
					resource.TestCheckTypeSetElemAttr("auth0_tenant.my_tenant", "acr_values_supported.*", "bar"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.0.disable", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.html", "<html></html>"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.show_log_link", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.0.url", "https://mycompany.org/error"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "oidc_logout.0.rp_logout_end_session_endpoint_discovery", "false"),
				),
			},
			{
				Config: testAccTenantConfigUpdateBack,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "acr_values_supported.#", "0"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "mtls.0.enable_endpoint_aliases", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "error_page.#", "0"),
				),
			},
		},
	})
}

// TestAccTenant_EnableSSO tests the enable_sso flag. This test is added separately because it can only be tested on existing tenants.
// For new tenants, this flag is always set to true.
func TestAccTenant_EnableSSO(t *testing.T) {
	t.Skip() // I have no tenants on which this works now.
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
				Config: testAccTenantEnableSSOConfigCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "168"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_sso", "false"),
				),
			},
			{
				Config: testAccTenantEnableSSOConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "session_lifetime", "168"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "idle_session_lifetime", "72"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "flags.0.enable_sso", "true"),
				),
			},
		},
	})
}

func TestAccTenantDefaultTokenQuota(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccTenantWithDefaultTokenQuota,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "friendly_name", "My Test Tenant with Token Quota"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.enforce", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.per_hour", "100"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.per_day", "2000"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.enforce", "true"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.per_hour", "200"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.per_day", "4000"),
				),
			},
			{
				Config: testAccTenantWithDefaultTokenQuotaUpdated,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "friendly_name", "My Test Tenant with Token Quota"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.enforce", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.per_hour", "50"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.clients.0.client_credentials.0.per_day", "1000"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.#", "1"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.enforce", "false"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.per_hour", "150"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.0.organizations.0.client_credentials.0.per_day", "3000"),
				),
			},
			{
				Config: testAccTenantWithDefaultTokenQuotaRemoved,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "friendly_name", "My Test Tenant with Token Quota"),
					resource.TestCheckResourceAttr("auth0_tenant.my_tenant", "default_token_quota.#", "0"),
				),
			},
		},
	})
}

func TestAccTenantDefaults(t *testing.T) {
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		// Only run with recorded HTTP requests because  normal E2E tests will naturally configure the tenant
		// and this test will only pass when the tenant has not been configured yet (aka "fresh" tenants).
		// In this test, just re-recording it should work, although you may need to comment out the `t.Skip()`.
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
