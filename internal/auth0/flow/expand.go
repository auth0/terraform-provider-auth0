package flow

import (
	"encoding/json"
	"fmt"

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

func expandVaultConnection(data *schema.ResourceData) (*management.FlowVaultConnection, error) {
	cfg := data.GetRawConfig()

	vaultConnection := &management.FlowVaultConnection{}

	vaultConnection.Name = value.String(cfg.GetAttr("name"))

	if data.HasChange("app_id") {
		vaultConnection.AppID = value.String(cfg.GetAttr("app_id"))
	}

	if data.HasChange("environment") {
		vaultConnection.Environment = value.String(cfg.GetAttr("environment"))
	}

	if data.HasChange("setup") {
		setup, err := expandStringInterfaceMap(data, "setup")
		if err != nil {
			return nil, err
		}
		vaultConnection.Setup = &setup
	}

	if data.HasChange("account_name") {
		vaultConnection.AccountName = value.String(cfg.GetAttr("account_name"))
	}

	return vaultConnection, nil
}

func expandStringInterfaceMap(data *schema.ResourceData, key string) (map[string]interface{}, error) {
	raw := data.Get(key)
	if raw == nil {
		return nil, nil
	}

	if m, ok := raw.(map[string]interface{}); ok {
		return m, nil
	}

	return nil, fmt.Errorf("expected map for %s, got %T", key, raw)
}

func expandInterfaceArray(d *schema.ResourceData, key string) []interface{} {
	oldMetadata, newMetadata := d.GetChange(key)
	result := make([]interface{}, 0)
	if oldMetadata == "" && newMetadata == "" {
		return result
	}

	if newMetadata == "" {
		return result
	}

	if newMetadataStr, ok := newMetadata.(string); ok {
		var newMetadataArr []interface{}
		if err := json.Unmarshal([]byte(newMetadataStr), &newMetadataArr); err != nil {
			return nil
		}
		return newMetadataArr
	}

	if newMetadataArr, ok := newMetadata.([]interface{}); ok {
		return newMetadataArr
	}

	return result
}
