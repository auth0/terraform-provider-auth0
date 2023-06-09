package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const (
	// SeparatorColon used for imports that use ":".
	SeparatorColon = ":"

	// SeparatorDoubleColon used for imports that use "::".
	SeparatorDoubleColon = "::"
)

var errEmptyID = fmt.Errorf("ID cannot be empty")

// ImportResourceGroupID deconstructs the given ID when terraform import
// runs, so the attribute groups can be set within the terraform state.
func ImportResourceGroupID(separator string, resourceGroup ...string) schema.StateContextFunc {
	return func(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
		givenRawID := data.Id()
		if givenRawID == "" {
			return nil, errEmptyID
		}

		if !strings.Contains(givenRawID, separator) {
			return nil, errInvalidID(separator, resourceGroup...)
		}

		idGroup := strings.Split(givenRawID, separator)
		if len(idGroup) != len(resourceGroup) {
			return nil, errInvalidID(separator, resourceGroup...)
		}

		var result *multierror.Error
		for index, attribute := range resourceGroup {
			result = multierror.Append(result, data.Set(attribute, idGroup[index]))
		}

		return []*schema.ResourceData{data}, result.ErrorOrNil()
	}
}

func errInvalidID(separator string, resourceGroup ...string) error {
	var formattedErrorMessage []string
	for _, s := range resourceGroup {
		formattedErrorMessage = append(formattedErrorMessage, "<"+s+">")
	}

	return fmt.Errorf("ID must be formatted as %s", strings.Join(formattedErrorMessage, separator))
}
