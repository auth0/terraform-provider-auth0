package resourceserver

import (
	"context"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_resource_server data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readResourceServerForDataSource,
		Description: "Data source to retrieve a specific Auth0 resource server by `resource_server_id` or `identifier`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))

	dataSourceSchema["resource_server_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the resource server. If not provided, `identifier` must be set.",
		AtLeastOneOf: []string{"resource_server_id", "identifier"},
	}

	dataSourceSchema["identifier"] = &schema.Schema{
		Type:     schema.TypeString,
		Optional: true,
		Description: "Unique identifier for the resource server. Used as the audience parameter " +
			"for authorization calls. If not provided, `resource_server_id` must be set. ",
		AtLeastOneOf: []string{"resource_server_id", "identifier"},
	}

	dataSourceSchema["scopes"] = &schema.Schema{
		Type:        schema.TypeSet,
		Computed:    true,
		Description: "List of permissions (scopes) used by this resource server.",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Name of the permission (scope). Examples include `read:appointments` or `delete:appointments`.",
				},
				"description": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: "Description of the permission (scope).",
				},
			},
		},
	}

	return dataSourceSchema
}

func readResourceServerForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	resourceServerID := data.Get("resource_server_id").(string)
	if resourceServerID == "" {
		resourceServerID = url.PathEscape(data.Get("identifier").(string))
	}

	api := meta.(*config.Config).GetAPI()
	resourceServer, err := api.ResourceServer.Read(ctx, resourceServerID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Ensuring the ID is the resource server ID and not the identifier,
	// as both can be used to find a resource server with the Read() func.
	data.SetId(resourceServer.GetID())

	return diag.FromErr(flattenResourceServerForDataSource(data, resourceServer))
}
