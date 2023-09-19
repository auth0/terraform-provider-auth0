package signingkey

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewDataSource will return a new auth0_signing_keys data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readSigningKeys,
		Description: "Data source to retrieve signing keys used by the applications in your tenant. [Learn more](https://auth0.com/docs/get-started/tenant-settings/signing-keys).",
		Schema: map[string]*schema.Schema{
			"signing_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All application signing keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"kid": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The key ID of the signing key.",
						},
						"cert": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public certificate of the signing key.",
						},
						"pkcs7": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The public certificate of the signing key in PKCS7 format.",
						},
						"current": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the key is the the current key.",
						},
						"next": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the key is the the next key.",
						},
						"previous": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the key is the the previous key.",
						},
						"revoked": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "True if the key is revoked.",
						},
						"fingerprint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cert fingerprint.",
						},
						"thumbprint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cert thumbprint.",
						},
					},
				},
			},
		},
	}
}

func readSigningKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	signingKeys, err := api.SigningKey.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id.UniqueId())

	return diag.FromErr(data.Set("signing_keys", flattenSigningKeys(signingKeys)))
}
