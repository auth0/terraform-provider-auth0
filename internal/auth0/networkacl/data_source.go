package networkacl

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_network_acl data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readNetworkACLForDataSource,
		Description: "Data source to retrieve a specific Auth0 Network ACL by ID.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Network ACL.",
	}

	return dataSourceSchema
}

func readNetworkACLForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	networkACLID := data.Get("id").(string)
	data.SetId(networkACLID)

	networkACL, err := api.NetworkACL.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenNetworkACL(data, networkACL))
}
