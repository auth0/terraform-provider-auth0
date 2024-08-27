package selfserviceprofile

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_self_service_profile data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readSelfServiceProfileForDataSource,
		Description: "Data source to retrieve a specific Auth0 Self-Service Profile by `id`",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Self Service Profile.",
	}
	dataSourceSchema["id"].Description = "The id of the Self Service Profile "
	return dataSourceSchema
}

func readSelfServiceProfileForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)
	data.SetId(id)
	ssp, err := api.SelfServiceProfile.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = flattenSelfServiceProfile(data, ssp)
	return diag.FromErr(err)
}