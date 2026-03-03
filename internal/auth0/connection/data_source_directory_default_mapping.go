package connection

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewDirectoryDefaultMappingDataSource will return a new auth0_connection_directory_default_mapping data source.
func NewDirectoryDefaultMappingDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readDirectoryDefaultMappingDataSource,
		Description: "Data source to retrieve the default attribute mapping for directory provisioning on an Auth0 connection by `connection_id`. " +
			"This shows the standard mapping that would be used if no custom mapping is specified.",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the connection to retrieve the default directory provisioning mapping.",
			},
			"mapping": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Default mapping between Auth0 attributes and IDP user attributes for this connection type.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"auth0": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The field location in the Auth0 schema.",
						},
						"idp": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The field location in the IDP schema.",
						},
					},
				},
			},
		},
	}
}

func readDirectoryDefaultMappingDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	connectionID := data.Get("connection_id").(string)
	defaultMapping, err := apiv2.Connections.DirectoryProvisioning.GetDefaultMapping(ctx, connectionID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connectionID)

	return diag.FromErr(data.Set("mapping", flattenDirectoryMappings(defaultMapping.GetMapping())))
}
