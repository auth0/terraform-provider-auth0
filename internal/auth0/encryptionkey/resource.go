package encryptionkey

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewEncryptionKeyResource will return a new auth0_encryption_keys resource.
func NewEncryptionKeyResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEncryptionKeys,
		UpdateContext: updateEncryptionKeys,
		ReadContext:   readEncryptionKeys,
		DeleteContext: deleteEncryptionKeys,
		Description:   "Resource to allow the rekeying of your tenant master key.",
		Schema: map[string]*schema.Schema{
			"rekey": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "If set to `true`, the encryption keys will be rotated.",
			},
			"encryption_keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "All encryption keys.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The key ID of the encryption key.",
						},
						"type": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "The type of the encryption key. One of " +
								"`customer-provided-root-key`, `environment-root-key`, " +
								"or `tenant-master-key`.",
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
							Description: "The state of the encryption key. One of " +
								"`pre-activation`, `active`, `deactivated`, or `destroyed`.",
						},
						"parent_key_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The key ID of the parent wrapping key.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ISO 8601 formatted date the encryption key was created.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ISO 8601 formatted date the encryption key was updated.",
						},
					},
				},
			},
		},
	}
}

func createEncryptionKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())

	return updateEncryptionKeys(ctx, data, meta)
}

func updateEncryptionKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if data.IsNewResource() || data.HasChange("rekey") {
		rekey := data.GetRawConfig().GetAttr("rekey")
		if !rekey.IsNull() && rekey.True() {
			if err := api.EncryptionKey.Rekey(ctx); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return readEncryptionKeys(ctx, data, meta)
}

func readEncryptionKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	encryptionKeys, err := api.EncryptionKey.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id.UniqueId())

	return diag.FromErr(data.Set("encryption_keys", flattenEncryptionKeys(encryptionKeys.Keys)))
}

func deleteEncryptionKeys(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
