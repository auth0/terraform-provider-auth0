package tokenexchangeprofile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/value"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandTokenExchangeProfiles(data *schema.ResourceData) *management.TokenExchangeProfile {
	cfg := data.GetRawConfig()

	return &management.TokenExchangeProfile{

		Name:             value.String(cfg.GetAttr("name")),
		SubjectTokenType: value.String(cfg.GetAttr("subject_token_type")),
		ActionID:         value.String(cfg.GetAttr("action_id")),
		Type:             value.String(cfg.GetAttr("type")),
	}
}
