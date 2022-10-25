package recorder

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/auth0/go-auth0/management"
	"github.com/stretchr/testify/require"
	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	recordingsDIR    = "./../../test/data/recordings/"
	recordingsTenant = "terraform-provider-auth0-dev"

	// RecordingsDomain is used for testing with our recorded http interactions.
	RecordingsDomain = recordingsTenant + ".eu.auth0.com"
)

// Recorder used for our http test mocks.
type Recorder struct {
	*recorder.Recorder
}

// New creates a new instance of our http recorder used in tests.
func New(t *testing.T) *Recorder {
	t.Helper()

	httpRecordings := os.Getenv("AUTH0_HTTP_RECORDINGS")
	httpRecordingsEnabled := httpRecordings == "true" || httpRecordings == "1" || httpRecordings == "on"
	if !httpRecordingsEnabled {
		return nil
	}

	recorderTransport, err := recorder.NewWithOptions(
		&recorder.Options{
			CassetteName:       recordingsDIR + t.Name(),
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

	return &Recorder{recorderTransport}
}

func removeSensitiveDataFromRecordings(t *testing.T, recorderTransport *recorder.Recorder) {
	recorderTransport.AddHook(
		func(i *cassette.Interaction) error {
			skip429Response(i)
			redactHeaders(i)

			domain := os.Getenv("AUTH0_DOMAIN")
			require.NotEmpty(t, domain, "removeSensitiveDataFromRecordings(): AUTH0_DOMAIN is empty")

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
	i.Request.Host = strings.Replace(i.Request.Host, domain, RecordingsDomain, -1)
	i.Request.URL = strings.Replace(i.Request.URL, domain, RecordingsDomain, -1)

	domainParts := strings.Split(domain, ".")

	i.Response.Body = strings.Replace(i.Response.Body, domainParts[0], recordingsTenant, -1)
	i.Request.Body = strings.Replace(i.Request.Body, domainParts[0], recordingsTenant, -1)
}

func redactSensitiveDataInClient(t *testing.T, i *cassette.Interaction, domain string) {
	create := i.Request.URL == "https://"+domain+"/api/v2/clients" && i.Request.Method == http.MethodPost
	read := strings.Contains(i.Request.URL, "https://"+domain+"/api/v2/clients/") && i.Request.Method == http.MethodGet
	update := strings.Contains(i.Request.URL, "https://"+domain+"/api/v2/clients/") && i.Request.Method == http.MethodPatch
	rotateSecret := strings.Contains(i.Request.URL, "clients") && strings.Contains(i.Request.URL, "/rotate-secret")

	if create || read || update || rotateSecret {
		var client management.Client
		err := json.Unmarshal([]byte(i.Response.Body), &client)
		require.NoError(t, err)

		redacted := "[REDACTED]"
		client.SigningKeys = []map[string]string{
			{"cert": redacted},
		}
		client.ClientSecret = &redacted

		clientBody, err := json.Marshal(client)
		require.NoError(t, err)

		i.Response.Body = string(clientBody)
	}
}
