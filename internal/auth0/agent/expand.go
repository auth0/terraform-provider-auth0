package agent

import (
	"encoding/json"
	"fmt"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandAgentCreate(data *schema.ResourceData) (*managementv3.CreateAgentRequestContent, error) {
	cfg := data.GetRawConfig()

	request := &managementv3.CreateAgentRequestContent{
		Name:            data.Get("name").(string),
		ExternalAgentID: value.String(cfg.GetAttr("external_agent_id")),
	}

	metadata, err := expandMetadata(value.String(cfg.GetAttr("metadata")))
	if err != nil {
		return nil, err
	}

	// Never send a null metadata on create, the API rejects it.
	if metadata != nil {
		request.Metadata = &metadata
	}

	return request, nil
}

func expandAgentUpdate(data *schema.ResourceData) (*managementv3.PatchAgentRequestParameters, error) {
	cfg := data.GetRawConfig()

	name := data.Get("name").(string)
	request := &managementv3.PatchAgentRequestParameters{Name: &name}

	if !data.HasChange("metadata") {
		return request, nil
	}

	metadata, err := expandMetadata(value.String(cfg.GetAttr("metadata")))
	if err != nil {
		return nil, err
	}

	// An explicit null is needed to clear metadata, as omitempty would otherwise drop the removal.
	request.SetMetadata(metadata)

	return request, nil
}

// expandMetadata parses the metadata JSON object. It returns nil when the metadata is not set.
func expandMetadata(raw *string) (managementv3.AgentMetadata, error) {
	if raw == nil || *raw == "" {
		return nil, nil
	}

	var metadata managementv3.AgentMetadata
	if err := json.Unmarshal([]byte(*raw), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata as a JSON object: %w", err)
	}

	return metadata, nil
}
