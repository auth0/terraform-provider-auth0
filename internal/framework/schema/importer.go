package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

const separator = "::"

var errEmptyID = "ID cannot be empty"

// ImportStateCompositeID is a helper function to set the import
// identifier to a group of state attribute paths. The attributes must accept a
// string value.
func ImportStateCompositeID(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse, attrPaths ...path.Path) {
	givenRawID := request.ID
	if givenRawID == "" {
		response.Diagnostics.AddError("Error", errEmptyID)
		return
	}

	if !strings.Contains(givenRawID, separator) {
		response.Diagnostics.Append(diagInvalidID(attrPaths...)...)
		return
	}

	idGroup := strings.Split(givenRawID, separator)
	if len(idGroup) != len(attrPaths) {
		response.Diagnostics.Append(diagInvalidID(attrPaths...)...)
		return
	}

	for index, attrPath := range attrPaths {
		response.Diagnostics.Append(response.State.SetAttribute(ctx, attrPath, idGroup[index])...)
	}
}

func diagInvalidID(resourceGroup ...path.Path) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	var formattedErrorMessage []string
	for _, s := range resourceGroup {
		formattedErrorMessage = append(formattedErrorMessage, fmt.Sprintf("<%s>", s.String()))
	}

	diagnostics.AddError("Error", fmt.Sprintf("ID must be formatted as %s", strings.Join(formattedErrorMessage, separator)))

	return diagnostics
}
