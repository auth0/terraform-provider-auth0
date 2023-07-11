package email

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/assert"
)

func TestEmailProviderIsConfigured(t *testing.T) {
	t.Run("it returns true if the provider is configured", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/emails/provider" {
				w.WriteHeader(http.StatusOK)
				return
			}
			http.NotFound(w, r)
		})
		testServer := httptest.NewServer(testHandler)

		api, err := management.New(testServer.URL, management.WithInsecure())
		assert.NoError(t, err)

		actual := emailProviderIsConfigured(context.Background(), api)
		assert.True(t, actual)
	})

	t.Run("it returns false if the provider is not configured", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/v2/emails/provider" {
				http.NotFound(w, r)
				return
			}
			http.NotFound(w, r)
		})
		testServer := httptest.NewServer(testHandler)

		api, err := management.New(testServer.URL, management.WithInsecure())
		assert.NoError(t, err)

		actual := emailProviderIsConfigured(context.Background(), api)
		assert.False(t, actual)
	})
}
