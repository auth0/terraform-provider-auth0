package networkaclkey

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenNetworkACLKey(data *schema.ResourceData, key *management.NetworkACLKey) error {
	if key == nil {
		return nil
	}

	return multierror.Append(
		data.Set("name", key.GetName()),
		data.Set("alg", string(key.GetAlg())),
		data.Set("fingerprint", key.GetFingerprint()),
		data.Set("created_at", key.GetCreatedAt()),
		data.Set("updated_at", key.GetUpdatedAt()),
	).ErrorOrNil()
}
