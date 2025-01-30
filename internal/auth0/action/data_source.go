package action

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewDataSource will return a new auth0_action data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readActionForDataSource,
		Description: "Data source to retrieve a specific Auth0 action by `name`.",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(internalSchema.Clone(NewResource().Schema))
	dataSourceSchema["id"] = &schema.Schema{
		Type:         schema.TypeString,
		Optional:     true,
		Description:  "The ID of the action. If not provided, `name` must be set.",
		AtLeastOneOf: []string{"id", "name"},
	}

	internalSchema.SetExistingAttributesAsOptional(dataSourceSchema, "name")
	dataSourceSchema["name"].Description = "The name of the action. If not provided, `id` must be set."
	dataSourceSchema["name"].AtLeastOneOf = []string{"id", "name"}

	return dataSourceSchema
}

func readActionForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	id := data.Get("id").(string)

	// If action_id is provided, use Get an Action API.
	if id != "" {
		data.SetId(id)
		action, err := api.Action.Read(ctx, id)
		if err != nil {
			return diag.FromErr(err)
		}
		return diag.FromErr(flattenAction(data, action))
	}

	// Else use Get Actions API and filter by name.
	name := data.Get("name").(string)

	// The Actions List API works on an exact name match.
	// Therefore, no need to create a List of Actions.
	// Either it's an exact match, or no match.
	// There are other params like deployed and triggerId to use as query param,
	// but they are not part of this implementation.
	actions, err := api.Action.List(ctx, management.Parameter("actionName", name))
	if err != nil {
		return diag.FromErr(err)
	}
	if len(actions.Actions) == 1 {
		data.SetId(*(actions.Actions[0].ID))
		return diag.FromErr(flattenAction(data, actions.Actions[0]))
	}
	return diag.Errorf("No action found with \"name\" = %q", name)
}
