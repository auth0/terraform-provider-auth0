package tenant

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestFlattenTenant(t *testing.T) {
	mockResourceData := schema.TestResourceDataRaw(t, NewResource().Schema, map[string]interface{}{})

	t.Run("it sets default values if remote tenant does not have them set", func(t *testing.T) {
		tenant := management.Tenant{
			IdleSessionLifetime: nil,
			SessionLifetime:     nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, mockResourceData.Get("idle_session_lifetime"), idleSessionLifetimeDefault)
		assert.Equal(t, mockResourceData.Get("session_lifetime"), sessionLifetimeDefault)
	})

	t.Run("it does not set default values if remote tenant has values set", func(t *testing.T) {
		tenant := management.Tenant{
			IdleSessionLifetime: auth0.Float64(73.5),
			SessionLifetime:     auth0.Float64(169.5),
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, mockResourceData.Get("idle_session_lifetime"), 73.5)
		assert.Equal(t, mockResourceData.Get("session_lifetime"), 169.5)
	})

	t.Run("it sets acr_values_supported if remote tenant has valid value set", func(t *testing.T) {
		tenant := management.Tenant{
			ACRValuesSupported: &[]string{"foo"},
		}

		assert.NoError(t, mockResourceData.Set("disable_acr_values_supported", true))
		assert.True(t, mockResourceData.Get("disable_acr_values_supported").(bool))
		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.False(t, mockResourceData.Get("disable_acr_values_supported").(bool))
		assert.Equal(t, mockResourceData.Get("acr_values_supported").(*schema.Set).Len(), 1)
		assert.Equal(t, mockResourceData.Get("acr_values_supported").(*schema.Set).List()[0].(string), "foo")
	})

	t.Run("it sets disable_acr_values_supported if remote tenant has null value set", func(t *testing.T) {
		tenant := management.Tenant{
			ACRValuesSupported: nil,
		}

		assert.NoError(t, mockResourceData.Set("acr_values_supported", &[]string{"foo"}))
		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.True(t, mockResourceData.Get("disable_acr_values_supported").(bool))
		assert.Equal(t, mockResourceData.Get("acr_values_supported").(*schema.Set).Len(), 0)
	})

	t.Run("it sets enable_endpoint_aliases if remote tenant has valid value set", func(t *testing.T) {
		tenant := management.Tenant{
			MTLS: &management.TenantMTLSConfiguration{
				EnableEndpointAliases: auth0.Bool(true),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, len(mockResourceData.Get("mtls").([]interface{})), 1)
		assert.True(t, mockResourceData.Get("mtls").([]interface{})[0].(map[string]interface{})["enable_endpoint_aliases"].(bool))
	})

	t.Run("it disables mtls if remote tenant has no value set", func(t *testing.T) {
		tenant := management.Tenant{
			MTLS: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.True(t, mockResourceData.Get("mtls").([]interface{})[0].(map[string]interface{})["disable"].(bool))
	})

	t.Run("it sets the error_page to nil if there is nothing set", func(t *testing.T) {
		tenant := management.Tenant{
			ErrorPage: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, mockResourceData.Get("error_page"), []interface{}{})
	})

	t.Run("it defaults enable_client_connections to true when API returns nil", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				EnableClientConnections: nil,
				EnableSSO:               auth0.Bool(true),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, enableClientConnectionsDefault, flags["enable_client_connections"])
		assert.Equal(t, true, flags["enable_sso"])
	})

	t.Run("it preserves enable_client_connections false when API returns explicit false", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				EnableClientConnections: auth0.Bool(false),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, false, flags["enable_client_connections"])
	})

	t.Run("it defaults oidc_logout_prompt_enabled to true when API returns nil sessions", func(t *testing.T) {
		tenant := management.Tenant{
			Sessions: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		sessions := mockResourceData.Get("sessions").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, oidcLogoutPromptEnabledDefault, sessions["oidc_logout_prompt_enabled"])
	})

	t.Run("it preserves oidc_logout_prompt_enabled false when API returns explicit false", func(t *testing.T) {
		tenant := management.Tenant{
			Sessions: &management.TenantSessions{
				OIDCLogoutPromptEnabled: auth0.Bool(false),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		sessions := mockResourceData.Get("sessions").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, false, sessions["oidc_logout_prompt_enabled"])
	})

	t.Run("it sets ephemeral session values correctly when returned by the API", func(t *testing.T) {
		tenant := management.Tenant{
			EphemeralSessionLifetime:     auth0.Float64(1.5),
			IdleEphemeralSessionLifetime: auth0.Float64(0.25),
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, 1.5, mockResourceData.Get("ephemeral_session_lifetime"))
		assert.Equal(t, 0.25, mockResourceData.Get("idle_ephemeral_session_lifetime"))
	})

	t.Run("it defaults ephemeral_session_lifetime to 72 when API returns nil", func(t *testing.T) {
		tenant := management.Tenant{
			EphemeralSessionLifetime:     nil,
			IdleEphemeralSessionLifetime: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, ephemeralSessionLifetimeDefault, mockResourceData.Get("ephemeral_session_lifetime"))
		assert.Equal(t, idleEphemeralSessionLifetimeDefault, mockResourceData.Get("idle_ephemeral_session_lifetime"))
	})

	t.Run("it defaults enable_pipeline2 to true when API returns nil", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				EnablePipeline2: nil,
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, enablePipeline2Default, flags["enable_pipeline2"])
	})

	t.Run("it preserves enable_pipeline2 false when API returns explicit false", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				EnablePipeline2: auth0.Bool(false),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, false, flags["enable_pipeline2"])
	})

	t.Run("it defaults disable_management_api_sms_obfuscation to true when API returns nil", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				DisableManagementAPISMSObfuscation: nil,
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, disableManagementAPISMSObfuscationDefault, flags["disable_management_api_sms_obfuscation"])
	})

	t.Run("it preserves disable_management_api_sms_obfuscation false when API returns explicit false", func(t *testing.T) {
		tenant := management.Tenant{
			Flags: &management.TenantFlags{
				DisableManagementAPISMSObfuscation: auth0.Bool(false),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		flags := mockResourceData.Get("flags").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, false, flags["disable_management_api_sms_obfuscation"])
	})

	t.Run("it defaults rp_logout_end_session_endpoint_discovery to true when API returns nil", func(t *testing.T) {
		tenant := management.Tenant{
			OIDCLogout: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		oidcLogout := mockResourceData.Get("oidc_logout").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, rpLogoutEndSessionEndpointDiscoveryDefault, oidcLogout["rp_logout_end_session_endpoint_discovery"])
	})

	t.Run("it preserves rp_logout_end_session_endpoint_discovery false when API returns explicit false", func(t *testing.T) {
		tenant := management.Tenant{
			OIDCLogout: &management.TenantOIDCLogout{
				OIDCResourceProviderLogoutEndSessionEndpointDiscovery: auth0.Bool(false),
			},
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		oidcLogout := mockResourceData.Get("oidc_logout").([]interface{})[0].(map[string]interface{})
		assert.Equal(t, false, oidcLogout["rp_logout_end_session_endpoint_discovery"])
	})

	t.Run("it sets dynamic_client_registration_security_mode when returned by API", func(t *testing.T) {
		tenant := management.Tenant{
			DynamicClientRegistrationSecurityMode: auth0.String("strict"),
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, "strict", mockResourceData.Get("dynamic_client_registration_security_mode"))
	})

	t.Run("it handles nil dynamic_client_registration_security_mode for newer tenants", func(t *testing.T) {
		tenant := management.Tenant{
			DynamicClientRegistrationSecurityMode: nil,
		}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, "", mockResourceData.Get("dynamic_client_registration_security_mode"))
	})

	t.Run("it handles missing dynamic_client_registration_security_mode for newer tenants", func(t *testing.T) {
		tenant := management.Tenant{}

		err := flattenTenant(mockResourceData, &tenant)

		assert.NoError(t, err)
		assert.Equal(t, "", mockResourceData.Get("dynamic_client_registration_security_mode"))
	})
}
