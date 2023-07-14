package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const separator = "::"

var errEmptyID = fmt.Errorf("ID cannot be empty")

// ImportResourceGroupID deconstructs the given ID when terraform import
// runs, so the attribute groups can be set within the terraform state.
func ImportResourceGroupID(resourceGroup ...string) schema.StateContextFunc {
	return func(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
		givenRawID := data.Id()
		if givenRawID == "" {
			return nil, errEmptyID
		}

		if !strings.Contains(givenRawID, separator) {
			return nil, errInvalidID(resourceGroup...)
		}

		idGroup := strings.Split(givenRawID, separator)
		if len(idGroup) != len(resourceGroup) {
			return nil, errInvalidID(resourceGroup...)
		}

		var result *multierror.Error
		for index, attribute := range resourceGroup {
			result = multierror.Append(result, data.Set(attribute, idGroup[index]))
		}

		return []*schema.ResourceData{data}, result.ErrorOrNil()
	}
}

// SetResourceGroupID sets the ID of the resource when the ID is a combination of
// multiple resource IDs. If the value is blank, then the resource is destroyed.
func SetResourceGroupID(data *schema.ResourceData, resourceGroup ...string) {
	data.SetId(strings.Join(resourceGroup, separator))
}

func errInvalidID(resourceGroup ...string) error {
	var formattedErrorMessage []string
	for _, s := range resourceGroup {
		formattedErrorMessage = append(formattedErrorMessage, "<"+s+">")
	}

	return fmt.Errorf("ID must be formatted as %s", strings.Join(formattedErrorMessage, separator))
}
