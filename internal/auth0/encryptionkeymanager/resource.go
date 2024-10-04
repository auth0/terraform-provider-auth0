package encryptionkeymanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewEncryptionKeyManagerResource will return a new auth0_encryption_key_manager resource.
func NewEncryptionKeyManagerResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEncryptionKeyManager,
		UpdateContext: updateEncryptionKeyManager,
		ReadContext:   readEncryptionKeyManager,
		DeleteContext: deleteEncryptionKeyManager,
		Description:   "Resource to allow the rekeying of your tenant master key.",
		Schema: map[string]*schema.Schema{
			"key_rotation_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "If this value is changed, the encryption keys will be rotated. A UUID is recommended for the `key_rotation_id`.",
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

func createEncryptionKeyManager(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())

	return updateEncryptionKeyManager(ctx, data, meta)
}

func updateEncryptionKeyManager(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if !data.IsNewResource() && data.HasChange("key_rotation_id") {
		keyRotationID := data.GetRawConfig().GetAttr("key_rotation_id")
		if !keyRotationID.IsNull() && len(keyRotationID.AsString()) > 0 {
			if err := api.EncryptionKey.Rekey(ctx); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return readEncryptionKeyManager(ctx, data, meta)
}

func readEncryptionKeyManager(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	encryptionKeys, err := api.EncryptionKey.List(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(id.UniqueId())

	return diag.FromErr(data.Set("encryption_keys", flattenEncryptionKeys(encryptionKeys.Keys)))
}

func deleteEncryptionKeyManager(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
	return nil
}
