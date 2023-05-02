package connection

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func importConnectionClient(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, errEmptyConnectionClientID
	}

	if !strings.Contains(rawID, ":") {
		return nil, errInvalidConnectionClientIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return nil, errInvalidConnectionClientIDFormat
	}

	result := multierror.Append(
		data.Set("connection_id", idPair[0]),
		data.Set("client_id", idPair[1]),
	)

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}
