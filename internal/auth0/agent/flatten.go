package agent

import (
	"encoding/json"
	"time"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenAgent(data *schema.ResourceData, agent *managementv3.AgentResponseContent) error {
	metadata, err := flattenMetadata(agent.GetMetadata())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("name", agent.GetName()),
		data.Set("external_agent_id", agent.GetExternalAgentID()),
		data.Set("metadata", metadata),
		data.Set("created_at", agent.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", agent.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

// flattenMetadata serializes the metadata to JSON. The API always returns an
// empty object for unset metadata, which is flattened to an empty string.
func flattenMetadata(metadata managementv3.AgentMetadata) (string, error) {
	if len(metadata) == 0 {
		return "", nil
	}

	b, err := json.Marshal(metadata)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
