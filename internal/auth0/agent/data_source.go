package agent

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

// NewDataSource will return a new auth0_agent data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readAgentForDataSource,
		Description: "Data source to retrieve a specific Auth0 agent by its `agent_id` (EA only).",
		Schema:      dataSourceSchema(),
	}
}

func dataSourceSchema() map[string]*schema.Schema {
	dataSourceSchema := internalSchema.TransformResourceToDataSource(NewResource().Schema)
	dataSourceSchema["agent_id"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: internalValidation.IsAgentID,
		Description:  "The ID of the agent.",
	}

	return dataSourceSchema
}

func readAgentForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	agentID := data.Get("agent_id").(string)

	agent, err := apiv3.Agents.Read(ctx, agentID)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(agent.GetAgentID())

	return diag.FromErr(flattenAgent(data, agent))
}
