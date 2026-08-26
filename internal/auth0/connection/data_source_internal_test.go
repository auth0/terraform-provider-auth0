package connection

import (
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/assert"
)

func TestFilterEnabledClients(t *testing.T) {
	t.Run("keeps clients without a status", func(t *testing.T) {
		// The Management API omits the status field and only lists the clients
		// the connection is enabled for, so those must be kept.
		result := filterEnabledClients([]management.ConnectionEnabledClient{
			{ClientID: auth0.String("client-1")},
			{ClientID: auth0.String("client-2")},
		})

		assert.Len(t, result, 2)
		assert.Equal(t, "client-1", result[0].GetClientID())
		assert.Equal(t, "client-2", result[1].GetClientID())
	})

	t.Run("keeps clients with status set to true", func(t *testing.T) {
		result := filterEnabledClients([]management.ConnectionEnabledClient{
			{ClientID: auth0.String("client-1"), Status: auth0.Bool(true)},
		})

		assert.Len(t, result, 1)
		assert.Equal(t, "client-1", result[0].GetClientID())
	})

	t.Run("drops clients with status set to false", func(t *testing.T) {
		result := filterEnabledClients([]management.ConnectionEnabledClient{
			{ClientID: auth0.String("client-1"), Status: auth0.Bool(true)},
			{ClientID: auth0.String("client-2"), Status: auth0.Bool(false)},
			{ClientID: auth0.String("client-3")},
		})

		assert.Len(t, result, 2)
		assert.Equal(t, "client-1", result[0].GetClientID())
		assert.Equal(t, "client-3", result[1].GetClientID())
	})

	t.Run("returns an empty list for no clients", func(t *testing.T) {
		assert.Empty(t, filterEnabledClients(nil))
		assert.Empty(t, filterEnabledClients([]management.ConnectionEnabledClient{}))
	})
}
