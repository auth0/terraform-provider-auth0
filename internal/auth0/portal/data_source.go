package portal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_portal data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readPortalForDataSource,
		Description: "Data source to retrieve a specific Universal Portal by `id`. (EA only)",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["id"] = &schema.Schema{
		Type:        schema.TypeString,
		Required:    true,
		Description: "The ID of the portal.",
	}

	// The API never returns the secret, so the data source does not expose the secret attributes.
	client := dataSourceSchema["client"].Elem.(*schema.Resource)
	secretPost := client.Schema["client_secret_post"].Elem.(*schema.Resource)
	delete(secretPost.Schema, "client_secret")
	delete(secretPost.Schema, "client_secret_wo")
	delete(secretPost.Schema, "client_secret_wo_version")

	return dataSourceSchema
}

func readPortalForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	portal, err := apiv3.Portals.Get(ctx, data.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(portal.GetID())

	return flattenPortal(data, portal)
}
