package flow

import (
	"encoding/json"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenFlow(data *schema.ResourceData, form *management.Flow) error {
	result := multierror.Append(
		data.Set("name", form.GetName()),
		data.Set("actions", flattenFlowAction(form.Actions)),
	)
	return result.ErrorOrNil()
}

func flattenFlowAction(formNodes []interface{}) string {
	if formNodes == nil {
		return ""
	}

	nodeBytes, err := json.Marshal(formNodes)
	if err != nil {
		return ""
	}

	return string(nodeBytes)
}

func flattenVaultConnection(data *schema.ResourceData, vaultConnection *management.FlowVaultConnection) error {

	result := multierror.Append(
		data.Set("name", vaultConnection.Name),
		data.Set("app_id", vaultConnection.AppID),
		data.Set("environment", vaultConnection.Environment),
		data.Set("setup", vaultConnection.Setup),
		data.Set("account_name", vaultConnection.AccountName),
		data.Set("ready", vaultConnection.Ready),
		data.Set("fingerprint", vaultConnection.Fingerprint),
	)

	return result.ErrorOrNil()
}
