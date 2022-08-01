package provider

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/dnaeon/go-vcr/v2/cassette"
	"github.com/dnaeon/go-vcr/v2/recorder"
	"github.com/stretchr/testify/require"
)

const (
	recordingsDIR    = "./../../test/data/recordings/"
	recordingsDomain = "terraform-provider-auth0-dev.eu.auth0.com"
)

func configureHTTPRecorder(t *testing.T) *recorder.Recorder {
	t.Helper()

	httpRecordings := os.Getenv("AUTH0_HTTP_RECORDINGS")
	httpRecordingsEnabled := httpRecordings == "true" || httpRecordings == "1" || httpRecordings == "on"
	if !httpRecordingsEnabled {
		return nil
	}

	recorderTransport, err := recorder.New(recordingsDIR + t.Name())
	require.NoError(t, err)

	removeSensitiveDataFromRecordings(t, recorderTransport)

	t.Cleanup(func() {
		err := recorderTransport.Stop()
		require.NoError(t, err)
	})

	return recorderTransport
}

func removeSensitiveDataFromRecordings(t *testing.T, recorderTransport *recorder.Recorder) {
	requestHeaders := []string{"Authorization"}
	responseHeaders := []string{
		"Alt-Svc",
		"Cache-Control",
		"Cf-Cache-Status",
		"Cf-Ray",
		"Date",
		"Expect-Ct",
		"Ot-Baggage-Auth0-Request-Id",
		"Ot-Tracer-Sampled",
		"Ot-Tracer-Spanid",
		"Ot-Tracer-Traceid",
		"Server",
		"Set-Cookie",
		"Strict-Transport-Security",
		"Traceparent",
		"Tracestate",
		"Vary",
		"X-Content-Type-Options",
		"X-Ratelimit-Limit",
		"X-Ratelimit-Remaining",
		"X-Ratelimit-Reset",
	}

	recorderTransport.AddFilter(func(i *cassette.Interaction) error {
		for _, header := range requestHeaders {
			delete(i.Request.Headers, header)
		}
		for _, header := range responseHeaders {
			delete(i.Response.Headers, header)
		}
		return nil
	})
	recorderTransport.AddSaveFilter(func(i *cassette.Interaction) error {
		domain := os.Getenv("AUTH0_DOMAIN")
		require.NotEmpty(t, domain, "removeSensitiveDataFromRecordings(): AUTH0_DOMAIN is empty")

		redactSensitiveDataInClient(t, i, domain)

		i.URL = strings.Replace(i.URL, domain, recordingsDomain, -1)
		i.Duration = time.Millisecond

		return nil
	})
}

func redactSensitiveDataInClient(t *testing.T, i *cassette.Interaction, domain string) {
	create := i.URL == "https://"+domain+"/api/v2/clients" && i.Method == http.MethodPost
	read := strings.Contains(i.URL, "https://"+domain+"/api/v2/clients/") && i.Method == http.MethodGet
	update := strings.Contains(i.URL, "https://"+domain+"/api/v2/clients/") && i.Method == http.MethodPatch
	rotateSecret := strings.Contains(i.URL, "clients") && strings.Contains(i.URL, "/rotate-secret")

	if create || read || update || rotateSecret {
		var client management.Client
		err := json.Unmarshal([]byte(i.Response.Body), &client)
		require.NoError(t, err)

		client.SigningKeys = []map[string]string{
			{"cert": "[REDACTED]"},
		}
		client.ClientSecret = auth0.String("[REDACTED]")

		clientBody, err := json.Marshal(client)
		require.NoError(t, err)

		i.Response.Body = string(clientBody)
	}
}
