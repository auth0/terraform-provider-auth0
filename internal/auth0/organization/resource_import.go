package organization

import (
	"context"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func importOrganizationConnection(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, errEmptyOrganizationConnectionID
	}

	if !strings.Contains(rawID, ":") {
		return nil, errInvalidOrganizationConnectionIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return nil, errInvalidOrganizationConnectionIDFormat
	}

	result := multierror.Append(
		data.Set("organization_id", idPair[0]),
		data.Set("connection_id", idPair[1]),
	)

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}

func importOrganizationMember(
	_ context.Context,
	data *schema.ResourceData,
	_ interface{},
) ([]*schema.ResourceData, error) {
	rawID := data.Id()
	if rawID == "" {
		return nil, errEmptyOrganizationMemberID
	}

	if !strings.Contains(rawID, ":") {
		return nil, errInvalidOrganizationMemberIDFormat
	}

	idPair := strings.Split(rawID, ":")
	if len(idPair) != 2 {
		return nil, errInvalidOrganizationMemberIDFormat
	}

	result := multierror.Append(
		data.Set("organization_id", idPair[0]),
		data.Set("user_id", idPair[1]),
	)

	return []*schema.ResourceData{data}, result.ErrorOrNil()
}
