package flow

import (
	"fmt"
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/value"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandFlow(data *schema.ResourceData) (*management.Flow, error) {
	flow := &management.Flow{}

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

	if data.HasChange("ready") {
		vaultConnection.Ready = value.Bool(cfg.GetAttr("ready"))
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

	return nil, fmt.Errorf("expected map[string]interface{} for %s, got %T", key, raw)
}
