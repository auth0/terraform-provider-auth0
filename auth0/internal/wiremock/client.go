package wiremock

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
)

const (
	resetMappingsURI  = "/__admin/mappings/reset"
	resetScenariosURI = "/__admin/scenarios/reset"
)

// Client provides communication with the
// admin API of a WireMock server instance.
type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

// NewClient constructs a WireMock httpClient
// that communicates with the WireMock
// server on the given baseURL and port.
func NewClient(host string) *Client {
	return &Client{
		baseURL: &url.URL{
			Scheme: "http",
			Host:   host,
		},
		httpClient: http.DefaultClient,
	}
}

// NewManagementAPIClient creates a Go Auth0 management API httpClient configured to
// talk to this WireMock server instance.
func (c *Client) NewManagementAPIClient() (*management.Management, error) {
	return management.New(c.baseURL.String(), management.WithInsecure())
}

// Reset calls ResetAllMappings and ResetAllScenarios
// to provide us with a clean WireMock state.
func (c *Client) Reset(ctx context.Context) error {
	return multierror.Append(
		c.ResetAllMappings(ctx),
		c.ResetAllScenarios(ctx),
	).ErrorOrNil()
}

// ResetAllMappings causes the WireMock server to remove transient mappings and
// reload static mappings from disk.
func (c *Client) ResetAllMappings(ctx context.Context) error {
	response, err := c.doRequest(ctx, http.MethodPost, resetMappingsURI)
	if err != nil {
		return fmt.Errorf("failed to reset WireMock mappings: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("failed to close response body when resetting mappings: %+v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("the WireMock server returned %s when trying to reset mappings", response.Status)
	}

	return nil
}

// ResetAllScenarios causes the WireMock server to put all currently defined
// mappings' scenarios back into their initial state.
func (c *Client) ResetAllScenarios(ctx context.Context) error {
	response, err := c.doRequest(ctx, http.MethodPost, resetScenariosURI)
	if err != nil {
		return fmt.Errorf("failed to reset WireMock scenarios: %w", err)
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Printf("failed to close response body when resetting scenarios: %+v", err)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("the WireMock server returned %s when trying to reset scenarios", response.Status)
	}

	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.baseURL.String()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to WireMock: %w", err)
	}

	return response, nil
}
