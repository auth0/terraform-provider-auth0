package connection

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewKeysResource will return a new auth0_connection_keys resource.
func NewKeysResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: rotateConnectionKeys,
		ReadContext:   readConnectionKeys,
		UpdateContext: rotateConnectionKeys,
		DeleteContext: deleteConnectionKeys,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Client Assertion JWT is a more secure alternative to client secret authentication " +
			"for OIDC and Okta Workforce connections. It uses a signed JWT instead of " +
			"a shared secret to authenticate the client. The resource only supports key rotation. " +
			"Use the auth0_connection_keys data source to read existing keys. Removing the resource from " +
			"configuration will NOT DELETE the key.",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"triggers": {
				Type:     schema.TypeMap,
				Required: true,
				Description: "This is an arbitrary map, which when edited shall perform rotation of keys for corresponding connection. " +
					"It can host keys like version, timestamp of last rotation etc.",
				Elem: &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "The trigger key which when changed will perform rotation",
				},
			},
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
			"pkcs": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public certificate of the signing key in PKCS7 format.",
			},
			"current": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the key is the current key.",
			},
			"next": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the key is the next key.",
			},
			"previous": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the key is the previous key.",
			},
			"current_since": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The date and time when the key became the current key.",
			},
			"fingerprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The certificate fingerprint.",
			},
			"thumbprint": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The certificate thumbprint.",
			},
			"algorithm": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The signing key algorithm.",
			},
			"key_use": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The signing key use, whether for encryption or signing.",
			},
			"subject_dn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The subject distinguished name (DN) of the certificate.",
			},
		},
	}
}

func rotateConnectionKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)

	key, err := api.Connection.RotateKeys(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	internalSchema.SetResourceGroupID(data, connectionID, key.GetKID())

	return flattenConnectionKey(data, connectionID, key)
}

func readConnectionKeys(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)
	kid := data.Get("kid").(string)

	keys, err := api.Connection.ReadKeys(ctx, connectionID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	for _, key := range keys {
		if key.KID != nil && *key.KID == kid {
			return flattenConnectionKey(data, connectionID, key)
		}
	}

	// Key no longer exists.
	data.SetId("")
	return nil
}

func deleteConnectionKeys(_ context.Context, data *schema.ResourceData, _ interface{}) diag.Diagnostics {
	// No API call, maybe log something.
	log.Printf("[INFO] No delete endpoint for resource %s - Connection Keys, removing from state", data.Id())
	data.SetId("")
	return nil
}
