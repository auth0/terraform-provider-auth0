package networkaclkey

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandNetworkACLKey(data *schema.ResourceData) *management.CreateKeysNetworkACLsRequestContent {
	req := &management.CreateKeysNetworkACLsRequestContent{}

	req.SetName(data.Get("name").(string))
	req.SetAlg(management.NetworkACLKeyAlgorithmEnum(data.Get("alg").(string)))

	// Value is WriteOnly — must be read from raw config, not state.
	if rawVal := data.GetRawConfig().GetAttr("value"); !rawVal.IsNull() && rawVal.IsKnown() {
		req.SetValue(rawVal.AsString())
	}

	return req
}
