package tenant_test

//// import (
//// 	"fmt"
//// 	"os"
//// 	"testing"
//// 
//// 	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
//// 
//// 	"github.com/auth0/terraform-provider-auth0/internal/acctest"
//// )
//// 
//// const testAccDataTenantConfig = `
//// resource "auth0_tenant" "my_tenant" {
//// 	default_directory            = ""
//// 	default_audience             = ""
//// 	default_redirection_uri      = "https://example.com/login"
//// 	friendly_name                = "My Test Tenant"
//// 	picture_url                  = "https://mycompany.org/logo.png"
//// 	support_email                = "support@mycompany.org"
//// 	support_url                  = "https://mycompany.org/support"
//// 	allowed_logout_urls          = [ "https://mycompany.org/logoutCallback" ]
//// 	session_lifetime             = 720
//// 	sandbox_version              = "16"
//// 	idle_session_lifetime        = 72
//// 	enabled_locales              = ["en", "de", "fr"]
//// 	disable_acr_values_supported = true
//// 
//// 	flags {
//// 		disable_clickjack_protection_headers   = true
//// 		enable_public_signup_user_exists_error = true
//// 		use_scope_descriptions_for_consent     = true
//// 		no_disclose_enterprise_connections     = false
//// 		disable_management_api_sms_obfuscation = false
//// 		disable_fields_map_fix                 = false
//// 	}
//// 
//// 	session_cookie {
//// 		mode = "non-persistent"
//// 	}
//// 
//// 	mtls {
//// 		enable_endpoint_aliases = true
//// 	}
//// }
//// 
//// data "auth0_tenant" "current" {
//// 	depends_on = [ auth0_tenant.my_tenant ]
//// }
//// `
//// 
//// const testAccDataTenantConfigUpdate = `
//// resource "auth0_tenant" "my_tenant" {
//// 	default_directory            = ""
//// 	default_audience             = ""
//// 	default_redirection_uri      = "https://example.com/login"
//// 	friendly_name                = "My Test Tenant"
//// 	picture_url                  = "https://mycompany.org/logo.png"
//// 	support_email                = "support@mycompany.org"
//// 	support_url                  = "https://mycompany.org/support"
//// 	allowed_logout_urls          = [ "https://mycompany.org/logoutCallback" ]
//// 	session_lifetime             = 720
//// 	sandbox_version              = "16"
//// 	idle_session_lifetime        = 72
//// 	enabled_locales              = ["en", "de", "fr"]
//// 	acr_values_supported         = ["foo", "bar"]
//// 
//// 	flags {
//// 		disable_clickjack_protection_headers   = true
//// 		enable_public_signup_user_exists_error = true
//// 		use_scope_descriptions_for_consent     = true
//// 		no_disclose_enterprise_connections     = false
//// 		disable_management_api_sms_obfuscation = false
//// 		disable_fields_map_fix                 = false
//// 	}
//// 
//// 	session_cookie {
//// 		mode = "non-persistent"
//// 	}
//// 
//// 	mtls {
//// 		disable = true
//// 	}
//// }
//// 
//// data "auth0_tenant" "current" {
//// 	depends_on = [ auth0_tenant.my_tenant ]
//// }
//// `
//// 
//// func TestAccDataSourceTenant(t *testing.T) {
//// 	acctest.Test(t, resource.TestCase{
//// 		Steps: []resource.TestStep{
//// 			{
//// 				Config: acctest.ParseTestName(testAccDataTenantConfig, t.Name()),
//// 				Check: resource.ComposeTestCheckFunc(
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "domain", os.Getenv("AUTH0_DOMAIN")),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "management_api_identifier", fmt.Sprintf("https://%s/api/v2/", os.Getenv("AUTH0_DOMAIN"))),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "default_audience", ""),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "default_directory", ""),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "friendly_name", "My Test Tenant"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "picture_url", "https://mycompany.org/logo.png"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "support_email", "support@mycompany.org"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "support_url", "https://mycompany.org/support"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "allowed_logout_urls.0", "https://mycompany.org/logoutCallback"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "session_lifetime", "720"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "sandbox_version", "16"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "idle_session_lifetime", "72"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "enabled_locales.0", "en"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "enabled_locales.1", "de"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "enabled_locales.2", "fr"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "flags.0.disable_clickjack_protection_headers", "true"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "flags.0.enable_public_signup_user_exists_error", "true"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "flags.0.use_scope_descriptions_for_consent", "true"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "default_redirection_uri", "https://example.com/login"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "session_cookie.0.mode", "non-persistent"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "disable_acr_values_supported", "true"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "acr_values_supported.#", "0"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "mtls.#", "1"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "mtls.0.enable_endpoint_aliases", "true"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "mtls.0.disable", "false"),
//// 				),
//// 			},
//// 			{
//// 				Config: acctest.ParseTestName(testAccDataTenantConfigUpdate, t.Name()),
//// 				Check: resource.ComposeTestCheckFunc(
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "disable_acr_values_supported", "false"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "acr_values_supported.#", "2"),
//// 					resource.TestCheckTypeSetElemAttr("data.auth0_tenant.current", "acr_values_supported.*", "foo"),
//// 					resource.TestCheckTypeSetElemAttr("data.auth0_tenant.current", "acr_values_supported.*", "bar"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "mtls.#", "1"),
//// 					resource.TestCheckResourceAttr("data.auth0_tenant.current", "mtls.0.disable", "true"),
//// 				),
//// 			},
//// 		},
//// 	})
//// }
