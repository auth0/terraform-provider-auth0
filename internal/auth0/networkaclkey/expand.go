package networkaclkey

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandNetworkACLKey(data *schema.ResourceData) *management.CreateKeysNetworkACLsRequestContent {
	req := &management.CreateKeysNetworkACLsRequestContent{}

	req.SetName(data.Get("name").(string))
	req.SetAlg(management.NetworkACLKeyAlgorithmEnum(data.Get("alg").(string)))
	req.SetValue(data.Get("value").(string))

	return req
}
