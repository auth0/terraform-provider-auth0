package tokenexchangeprofile

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenTokenExchangeProfile(data *schema.ResourceData, tokenExchangeProfile *management.TokenExchangeProfile) error {
	result := multierror.Append(
		data.Set("name", tokenExchangeProfile.Name),
		data.Set("subject_token_type", tokenExchangeProfile.SubjectTokenType),
		data.Set("action_id", tokenExchangeProfile.GetActionID()),
		data.Set("type", tokenExchangeProfile.GetType()),
		//data.Set("created_at", tokenExchangeProfile.GetCreatedAt()),
		//data.Set("updated_at", tokenExchangeProfile.GetUpdatedAt()),
	)
	return result.ErrorOrNil()
}
