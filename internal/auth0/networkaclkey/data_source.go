package networkaclkey

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewDataSource returns a new auth0_network_acl_key data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readNetworkACLKeyDataSource,
		Description: "Retrieve a Network ACL key by its ID. The `value` field is not available " +
			"via this data source (it is write-only on the resource). (EA Only)",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Network ACL key to retrieve.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User-supplied label for the key.",
			},
			"alg": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Algorithm used for the key.",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "SHA-256 fingerprint of the decoded key material (lowercase hex).",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time when the key was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time when the key was last updated.",
			},
		},
	}
}

func readNetworkACLKeyDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	id := data.Get("id").(string)

	key, err := apiv3.Keys.NetworkACLs.Get(ctx, id)
	if err != nil {
		return internalError.HandleReadAPIError("auth0_network_acl_key", data, err)
	}

	data.SetId(id)

	return diag.FromErr(flattenNetworkACLKey(data, key))
}
