package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewManagementV3DisablesSDKRetries(t *testing.T) {
	var apiCalls int32
	testServer := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&apiCalls, 1)
		writer.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(testServer.Close)

	config := ProviderConfig{
		Domain:   strings.TrimPrefix(testServer.URL, "https://"),
		APIToken: "test-token",
	}

	// The test server's client has no retry transport, so any retry seen here
	// would come from the SDK itself.
	terraformVersion := "1.0.0"
	client, err := newManagementV3(config, &terraformVersion, testServer.Client())
	require.NoError(t, err)

	_, err = client.Organizations.Clients.Get(context.Background(), "org_123", "client_123")
	require.Error(t, err)

	assert.Equal(t, int32(1), atomic.LoadInt32(&apiCalls), "the v3 SDK must not retry; retries belong to the provider's HTTP client")
}
