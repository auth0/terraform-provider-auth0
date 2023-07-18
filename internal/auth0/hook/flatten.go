package hook

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenHook(data *schema.ResourceData, hook *management.Hook) error {
	result := multierror.Append(
		data.Set("name", hook.GetName()),
		data.Set("dependencies", hook.GetDependencies()),
		data.Set("script", hook.GetScript()),
		data.Set("trigger_id", hook.GetTriggerID()),
		data.Set("enabled", hook.GetEnabled()),
	)

	return result.ErrorOrNil()
}
