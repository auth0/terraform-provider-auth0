package acctest

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ExtractResourceAttributeFromState extracts the value of a resource's attribute into the target.
func ExtractResourceAttributeFromState(state *terraform.State, name, key string) (string, error) {
	tfResource, ok := state.RootModule().Resources[name]
	if !ok {
		return "", fmt.Errorf("extract resource attribute: failed to find resource with name: %q", name)
	}

	attribute, ok := tfResource.Primary.Attributes[key]
	if !ok {
		return "", fmt.Errorf("extract resource attribute: failed to find attribute with name: %q", key)
	}

	return attribute, nil
}
