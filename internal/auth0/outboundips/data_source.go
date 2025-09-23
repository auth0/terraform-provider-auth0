package outboundips

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const outboundIPsURL = "https://cdn.auth0.com/ip-ranges.json"

// NewDataSource will return a new auth0_outbound_ips resource.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readOutboundIPs,
		Description: "Use this data source to retrieve Auth0's outbound IP ranges for allowlisting purposes.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"last_updated_at": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "RFC3339 timestamp when the IP ranges were last updated.",
		},
		"regions": {
			Description: "A list of regions and their corresponding IP CIDR blocks.",
			Type:        schema.TypeList,
			Computed:    true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"region": {
						Description: "The code for the region (e.g., 'US', 'CA').",
						Type:        schema.TypeString,
						Computed:    true,
					},
					"ipv4_cidrs": {
						Description: "A list of IPv4 CIDR blocks for the region.",
						Type:        schema.TypeList,
						Computed:    true,
						Elem: &schema.Schema{
							Type: schema.TypeString,
						},
					},
				},
			},
		},
		"changelog": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "List of recent changes to IP ranges.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"date": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Date of the change (YYYY-MM-DD format).",
					},
					"region": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Region affected by the change.",
					},
					"action": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Type of change (add or remove).",
					},
					"ipv4_cidrs": {
						Type:        schema.TypeList,
						Computed:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "List of IPv4 CIDR blocks affected by this change.",
					},
				},
			},
		},
	}
}

func readOutboundIPs(ctx context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())

	response, err := fetchOutboundIPs(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := flattenOutboundIPs(data, response); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// fetchOutboundIPs retrieves the outbound IP ranges from Auth0's CDN endpoint.
func fetchOutboundIPs(ctx context.Context) (*outboundIPsResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", outboundIPsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "terraform-provider-auth0")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outbound IP ranges: %w", err)
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch outbound IP ranges: HTTP %d - the API endpoint may not be available yet", resp.StatusCode)
	}

	var result outboundIPsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate the response structure.
	if result.LastUpdatedAt == "" {
		return nil, fmt.Errorf("invalid response: missing last_updated_at field")
	}

	if len(result.Regions) == 0 {
		return nil, fmt.Errorf("invalid response: no regions found")
	}

	return &result, nil
}
