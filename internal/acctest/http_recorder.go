package acctest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"testing"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	recordingsDIR    = "test/data/recordings/"
	recordingsTenant = "terraform-provider-auth0-dev"

	// RecordingsDomain is used for testing with our recorded http interactions.
	RecordingsDomain = recordingsTenant + ".eu.auth0.com"
)

// NewHTTPRecorder creates a new instance of our http recorder used in tests.
func newHTTPRecorder(t *testing.T) *recorder.Recorder {
	t.Helper()

	recorderTransport, err := recorder.NewWithOptions(
		&recorder.Options{
			CassetteName:       cassetteName(t.Name()),
			Mode:               recorder.ModeRecordOnce,
			SkipRequestLatency: true,
		},
	)
	require.NoError(t, err)

	removeSensitiveDataFromRecordings(t, recorderTransport)

	t.Cleanup(func() {
		err := recorderTransport.Stop()
		require.NoError(t, err)
	})

	return recorderTransport
}

func cassetteName(testName string) string {
	_, file, _, _ := runtime.Caller(0)
	rootDir := path.Join(path.Dir(file), "../..")
	return path.Join(rootDir, recordingsDIR, testName)
}

func removeSensitiveDataFromRecordings(t *testing.T, recorderTransport *recorder.Recorder) {
	recorderTransport.AddHook(
		func(i *cassette.Interaction) error {
			skip429Response(i)
			redactHeaders(i)

			domain := os.Getenv("AUTH0_DOMAIN")
			require.NotEmpty(t, domain, "removeSensitiveDataFromRecordings(): AUTH0_DOMAIN is empty")

			redactSensitiveDataInSigningKeys(t, i, domain)
			redactSensitiveDataInClient(t, i, domain)
			redactDomain(i, domain)

			return nil
		},
		recorder.BeforeSaveHook,
	)
}

func skip429Response(i *cassette.Interaction) {
	if i.Response.Code == http.StatusTooManyRequests {
		i.DiscardOnSave = true
	}
}

func redactHeaders(i *cassette.Interaction) {
	allowedHeaders := map[string]bool{
		"Content-Type": true,
		"User-Agent":   true,
	}

	for header := range i.Request.Headers {
		if _, ok := allowedHeaders[header]; !ok {
			delete(i.Request.Headers, header)
		}
	}
	for header := range i.Response.Headers {
		if _, ok := allowedHeaders[header]; !ok {
			delete(i.Response.Headers, header)
		}
	}
}

func redactDomain(i *cassette.Interaction, domain string) {
	i.Request.Host = strings.ReplaceAll(i.Request.Host, domain, RecordingsDomain)
	i.Request.URL = strings.ReplaceAll(i.Request.URL, domain, RecordingsDomain)

	domainParts := strings.Split(domain, ".")

	i.Response.Body = strings.ReplaceAll(i.Response.Body, domainParts[0], recordingsTenant)
	i.Request.Body = strings.ReplaceAll(i.Request.Body, domainParts[0], recordingsTenant)
}

func redactSensitiveDataInClient(t *testing.T, i *cassette.Interaction, domain string) {
	baseURL := "https://" + domain + "/api/v2/clients"
	urlPath := strings.Split(i.Request.URL, "?")[0] // Strip query params.

	create := i.Request.URL == baseURL &&
		i.Request.Method == http.MethodPost

	readList := urlPath == baseURL &&
		i.Request.Method == http.MethodGet

	readOne := strings.Contains(i.Request.URL, baseURL+"/") &&
		!strings.Contains(i.Request.URL, "credentials") &&
		i.Request.Method == http.MethodGet

	update := strings.Contains(i.Request.URL, baseURL+"/") &&
		!strings.Contains(i.Request.URL, "credentials") &&
		i.Request.Method == http.MethodPatch

	if create || readList || readOne || update {
		if i.Response.Code == http.StatusNotFound {
			return
		}

		redacted := "[REDACTED]"

		// Handle list response.
		if readList {
			var response management.ClientList
			err := json.Unmarshal([]byte(i.Response.Body), &response)
			require.NoError(t, err)

			for _, client := range response.Clients {
				client.SigningKeys = []map[string]string{
					{"cert": redacted},
				}
				if client.GetClientSecret() != "" {
					client.ClientSecret = &redacted
				}
			}

			responseBody, err := json.Marshal(response)
			require.NoError(t, err)
			i.Response.Body = string(responseBody)
			return
		}

		// Handle single client response.
		var client management.Client
		err := json.Unmarshal([]byte(i.Response.Body), &client)
		require.NoError(t, err)

		client.SigningKeys = []map[string]string{
			{"cert": redacted},
		}

		if client.GetClientSecret() != "" {
			client.ClientSecret = &redacted
		}

		clientBody, err := json.Marshal(client)
		require.NoError(t, err)

		i.Response.Body = string(clientBody)
	}
}

func redactSensitiveDataInSigningKeys(t *testing.T, i *cassette.Interaction, domain string) {
	read := i.Request.URL == "https://"+domain+"/api/v2/keys/signing" && i.Request.Method == http.MethodGet
	if read {
		currentSigningKey := &management.SigningKey{
			KID:         auth0.String("111111111111111111111"),
			Cert:        auth0.String("-----BEGIN CERTIFICATE-----\\r\\n[REDACTED]\\r\\n-----END CERTIFICATE-----"),
			PKCS7:       auth0.String("-----BEGIN PKCS7-----\\r\\n[REDACTED]\\r\\n-----END PKCS7-----"),
			Current:     auth0.Bool(true),
			Next:        auth0.Bool(false),
			Previous:    auth0.Bool(true),
			Fingerprint: auth0.String("[REDACTED]"),
			Thumbprint:  auth0.String("[REDACTED]"),
			Revoked:     auth0.Bool(false),
		}
		previousSigningKey := &management.SigningKey{
			KID:         auth0.String("222222222222222222222"),
			Cert:        auth0.String("-----BEGIN CERTIFICATE-----\\r\\n[REDACTED]\\r\\n-----END CERTIFICATE-----"),
			PKCS7:       auth0.String("-----BEGIN PKCS7-----\\r\\n[REDACTED]\\r\\n-----END PKCS7-----"),
			Current:     auth0.Bool(false),
			Next:        auth0.Bool(true),
			Previous:    auth0.Bool(true),
			Fingerprint: auth0.String("[REDACTED]"),
			Thumbprint:  auth0.String("[REDACTED]"),
			Revoked:     auth0.Bool(false),
		}

		currentSigningKeyBody, err := json.Marshal(currentSigningKey)
		require.NoError(t, err)

		previousSigningKeyBody, err := json.Marshal(previousSigningKey)
		require.NoError(t, err)

		i.Response.Body = fmt.Sprintf(`[%s,%s]`, currentSigningKeyBody, previousSigningKeyBody)
	}
}
