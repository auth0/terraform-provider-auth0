package action

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewModuleDataSource will return a new auth0_action_module data source.
func NewModuleDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readActionModuleForDataSource,
		Description: "Data source to retrieve a specific Auth0 action module by `id`.",
		Schema:      moduleDataSourceSchema(),
	}
}

func moduleDataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewModuleResource().Schema))
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the action module.",
	}

	return dataSourceSchema
}

func readActionModuleForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	id := data.Get("id").(string)

	data.SetId(id)

	module, err := apiv2.Actions.Modules.Get(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenActionModule(data, module))
}
