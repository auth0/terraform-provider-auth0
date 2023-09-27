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
}
