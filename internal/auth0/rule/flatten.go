package rule

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenRule(data *schema.ResourceData, rule *management.Rule) error {
	result := multierror.Append(
		data.Set("name", rule.GetName()),
		data.Set("script", rule.GetScript()),
		data.Set("order", rule.GetOrder()),
		data.Set("enabled", rule.GetEnabled()),
	)

	return result.ErrorOrNil()
}
