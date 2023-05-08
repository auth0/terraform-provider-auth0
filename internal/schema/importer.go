package schema

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const separator = ":"

var (
	errEmptyID         = fmt.Errorf("ID cannot be empty")
	errInvalidIDFormat = "ID must be formatted as <%s>:<%s>"
)

// ImportResourcePairID deconstructs the given ID when terraform
// import runs, so the 2 pairs can be set within terraform state.
func ImportResourcePairID(resourceAID, resourceBID string) schema.StateContextFunc {
	return func(_ context.Context, data *schema.ResourceData, _ interface{}) ([]*schema.ResourceData, error) {
		givenRawID := data.Id()
		if givenRawID == "" {
			return nil, errEmptyID
		}

		if !strings.Contains(givenRawID, separator) {
			return nil, fmt.Errorf(errInvalidIDFormat, resourceAID, resourceBID)
		}

		idPair := strings.Split(givenRawID, separator)
		if len(idPair) != 2 {
			return nil, fmt.Errorf(errInvalidIDFormat, resourceAID, resourceBID)
		}

		result := multierror.Append(
			data.Set(resourceAID, idPair[0]),
			data.Set(resourceBID, idPair[1]),
		)

		return []*schema.ResourceData{data}, result.ErrorOrNil()
	}
}
