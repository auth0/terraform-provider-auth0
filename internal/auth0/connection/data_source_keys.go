package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewKeysDataSource will return a new auth0_connection_keys data source.
func NewKeysDataSource() *schema.Resource {
	return &schema.Resource{
		Description: "Retrieve the list of signing keys for a specific Auth0 connection. (Okta/OIDC only)",
		ReadContext: readConnectionKeysDataSource,

		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the connection to retrieve keys for.",
			},
			"keys": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of signing keys associated with the connection.",
				Elem: &schema.Resource{
					Schema: internalSchema.TransformResourceToDataSource(NewKeysResource().Schema),
				},
			},
		},
	}
}

func readConnectionKeysDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connectionID := data.Get("connection_id").(string)

	keys, err := api.Connection.ReadKeys(ctx, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	var keyList []map[string]interface{}
	for _, key := range keys {
		keyList = append(keyList, flattenConnectionKeyMap(connectionID, key))
	}

	if err := data.Set("keys", keyList); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)
	return nil
}
