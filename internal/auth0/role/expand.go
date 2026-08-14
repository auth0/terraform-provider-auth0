package role

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// expandRole builds the payload for both create and update. Type and OwnerID
// are only accepted on create, and the SDK strips them from update requests.
func expandRole(data *schema.ResourceData) *management.Role {
	cfg := data.GetRawConfig()

	return &management.Role{
		Name:        value.String(cfg.GetAttr("name")),
		Description: value.String(cfg.GetAttr("description")),
		Type:        value.String(cfg.GetAttr("type")),
		OwnerID:     value.String(cfg.GetAttr("owner_id")),
	}
}
