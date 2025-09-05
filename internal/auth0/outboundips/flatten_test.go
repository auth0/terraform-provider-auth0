package outboundips

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenOutboundIPs(t *testing.T) {
	response := &outboundIPsResponse{
		LastUpdatedAt: "2024-11-10T12:00:00Z",
		Regions: map[string]regionInfo{
			"US": {
				IPv4CIDRs: []string{"174.129.105.183/32", "18.116.79.126/32"},
			},
			"CA": {
				IPv4CIDRs: []string{"15.222.97.193/32", "3.97.144.31/32"},
			},
		},
		Changelog: []changelogEntry{
			{
				Date:      "2024-11-10",
				Region:    "CA",
				Action:    "add",
				IPv4CIDRs: []string{"15.222.97.193/32"},
			},
		},
	}

	// Create a test schema.ResourceData.
	resourceSchema := dataSourceSchema()
	data := schema.TestResourceDataRaw(t, resourceSchema, map[string]interface{}{})

	err := flattenOutboundIPs(data, response)
	require.NoError(t, err)

	// Verify last_updated_at.
	assert.Equal(t, "2024-11-10T12:00:00Z", data.Get("last_updated_at"))

	// Verify regions map structure.
	regions := data.Get("regions").([]interface{})
	assert.Len(t, regions, 2)

	assert.Equal(t, []string{"174.129.105.183/32", "18.116.79.126/32"}, ipv4ForRegion(regions, "US"))
	assert.Equal(t, []string{"15.222.97.193/32", "3.97.144.31/32"}, ipv4ForRegion(regions, "CA"))

	// Verify changelog.
	changelog := data.Get("changelog").([]interface{})
	assert.Len(t, changelog, 1)
	changelogEntry := changelog[0].(map[string]interface{})
	assert.Equal(t, "2024-11-10", changelogEntry["date"])
	assert.Equal(t, "CA", changelogEntry["region"])
	assert.Equal(t, "add", changelogEntry["action"])
}

func ipv4ForRegion(regions []interface{}, region string) []string {
	ipv4 := []string{}
	for _, value := range regions {
		regionMap := value.(map[string]interface{})

		if regionMap["region"].(string) != region {
			continue
		}

		for _, v := range regionMap["ipv4_cidrs"].([]interface{}) {
			// Individual type assertions because we have to go from interface{} to []interface{} first.
			if ip, ok := v.(string); ok {
				ipv4 = append(ipv4, ip)
			}
		}
	}
	return ipv4
}
