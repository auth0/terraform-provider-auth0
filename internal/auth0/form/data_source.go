package form

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_form data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readFormDataSource,
		Description: "Data source to retrieve a specific Auth0 Form by `id`",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the Form.",
	}
	dataSourceSchema["id"].Description = "The id of the Form."
	return dataSourceSchema
}

func readFormDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)
	data.SetId(id)
	form, err := api.Form.Read(ctx, id)
	if err != nil {
		return diag.FromErr(err)
	}

	err = flattenForm(data, form)
	return diag.FromErr(err)
}
