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
		assert.Equal(t, mockResourceData.Get("idle_session_lifetime"), 72.00)
		assert.Equal(t, mockResourceData.Get("session_lifetime"), 168.00)
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

}
