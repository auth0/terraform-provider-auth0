package outboundips

import "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

// flattenOutboundIPs converts the API response into a list with region as key.
func flattenOutboundIPs(data *schema.ResourceData, response *outboundIPsResponse) error {
	if err := data.Set("last_updated_at", response.LastUpdatedAt); err != nil {
		return err
	}

	// Flatten regions.
	regions := make([]map[string]interface{}, 0, len(response.Regions))
	for code, region := range response.Regions {
		regions = append(regions, map[string]interface{}{
			"region":     code,
			"ipv4_cidrs": region.IPv4CIDRs,
		})
	}

	if err := data.Set("regions", regions); err != nil {
		return err
	}

	// Flatten changelog.
	changelog := make([]map[string]interface{}, 0, len(response.Changelog))
	for _, entry := range response.Changelog {
		changelogEntry := map[string]interface{}{
			"date":       entry.Date,
			"region":     entry.Region,
			"action":     entry.Action,
			"ipv4_cidrs": entry.IPv4CIDRs,
		}
		changelog = append(changelog, changelogEntry)
	}
	if err := data.Set("changelog", changelog); err != nil {
		return err
	}

	return nil
}
