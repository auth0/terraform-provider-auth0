package action

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// NewTriggerBindingResource will return a new auth0_trigger_binding resource.
func NewTriggerBindingResource() *schema.Resource {
	// TODO: remove this resource for v1 release.
	resource := NewTriggerActionsResource()

	resource.DeprecationMessage = "This resource has been renamed to `auth0_trigger_actions`. The `auth0_trigger_binding` alias will be removed in the next major version release."

	return resource
}
