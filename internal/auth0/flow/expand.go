package flow

import (
	"encoding/json"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandFlow(data *schema.ResourceData) (*management.Flow, error) {
	config := data.GetRawConfig()

	flow := &management.Flow{
		Name:    value.String(config.GetAttr("name")),
		Actions: expandInterfaceArray(data, "actions"),
	}

	return flow, nil
}

func expandInterfaceArray(d *schema.ResourceData, key string) []interface{} {
	oldMetadata, newMetadata := d.GetChange(key)
	result := make([]interface{}, 0)
	if oldMetadata == "" && newMetadata == "" {
		return result
	}

	if oldMetadata == "" {
		if newMetadataStr, ok := newMetadata.(string); ok {
			var newMetadataArr []interface{}
			if err := json.Unmarshal([]byte(newMetadataStr), &newMetadataArr); err != nil {
				return nil
			}
			return newMetadataArr
		}
		return result
	}

	if newMetadata == "" {
		return result
	}

	b, err := json.Marshal(newMetadata)
	if err != nil {
		return nil
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return nil
	}

	return result
}
