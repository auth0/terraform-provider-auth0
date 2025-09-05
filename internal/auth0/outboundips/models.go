package outboundips

// OutboundIPsResponse represents the response from Auth0's outbound IP ranges API.
type OutboundIPsResponse struct {
	LastUpdatedAt string                `json:"last_updated_at"`
	Regions       map[string]RegionInfo `json:"regions"`
	Changelog     []ChangelogEntry      `json:"changelog"`
}

// RegionInfo contains the IP ranges for a specific region.
type RegionInfo struct {
	IPv4CIDRs []string `json:"ipv4_cidrs"`
}

// ChangelogEntry represents a change in the IP ranges.
type ChangelogEntry struct {
	Date      string   `json:"date"`
	Region    string   `json:"region"`
	Action    string   `json:"action"`
	IPv4CIDRs []string `json:"ipv4_cidrs"`
}
