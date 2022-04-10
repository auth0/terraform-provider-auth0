package wiremock

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/go-retryablehttp"
)

// Client provides communciation with the admin API of a WireMock server
// instance.
type Client struct {
	host string
	cl   *retryablehttp.Client
}

// NewManagementAPIClient creates a Go Auth0 management API client configured to
// talk to this WireMock server instance.
func (c *Client) NewManagementAPIClient() (*management.Management, error) {
	return management.New(c.host, management.WithInsecure())
}

// ResetAllMappings causes the WireMock server to remove transient mappings and
// reload static mappings from disk.
func (c *Client) ResetAllMappings(ctx context.Context) (err error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "__admin/mappings/reset")
	if err != nil {
		return fmt.Errorf("failed to reset WireMock mappings: %w", err)
	}
	defer func() {
		err = multierror.Append(err, resp.Body.Close()).ErrorOrNil()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the WireMock server returned %s when trying to reset mappings", resp.Status)
	}
	return nil
}

// ResetAllScenarios causes the WireMock server to put all currently defined
// mappings' scenarios back into their initial state.
func (c *Client) ResetAllScenarios(ctx context.Context) (err error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "__admin/scenarios/reset")
	if err != nil {
		return fmt.Errorf("failed to reset WireMock scenarios: %w", err)
	}
	defer func() {
		err = multierror.Append(err, resp.Body.Close()).ErrorOrNil()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the WireMock server returned %s when trying to reset scenarios", resp.Status)
	}
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, path string) (*http.Response, error) {
	u := &url.URL{
		Scheme: "http",
		Host:   c.host,
		Path:   path,
	}

	req, err := retryablehttp.NewRequest(method, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := c.cl.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to send HTTP request to WireMock: %w", err)
	}
	return resp, nil
}

// NewClient constructs a WireMock client that communicates with the WireMock
// server on the given host and port.
func NewClient(host string) *Client {
	cl := retryablehttp.NewClient()
	cl.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy

	return &Client{
		host: host,
		cl:   cl,
	}
}
