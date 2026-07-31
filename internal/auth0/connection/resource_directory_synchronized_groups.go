package connection

import (
	"context"
	"errors"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	managementv2client "github.com/auth0/go-auth0/v2/management/client"
	"github.com/auth0/go-auth0/v2/management/core"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewDirectorySynchronizedGroupsResource will return a new auth0_connection_directory_synchronized_groups resource.
func NewDirectorySynchronizedGroupsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: upsertDirectorySynchronizedGroups,
		ReadContext:   readDirectorySynchronizedGroups,
		UpdateContext: upsertDirectorySynchronizedGroups,
		DeleteContext: deleteDirectorySynchronizedGroups,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage the set of Google Workspace group IDs " +
			"synchronized via directory provisioning for an Auth0 connection. (EA only)",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection for which to manage synchronized groups. (EA only)",
			},
			"group_ids": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Required:    true,
				Description: "List of Google Workspace Directory group IDs to synchronize. (EA only)",
			},
		},
	}
}

func upsertDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()
	connectionID := data.Get("connection_id").(string)

	groupIDs := value.Strings(data.GetRawConfig().GetAttr("group_ids"))

	if err := putSynchronizedGroups(ctx, apiv2, connectionID, *groupIDs); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(connectionID)

	return readDirectorySynchronizedGroups(ctx, data, meta)
}

func readDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	groupIDs, err := getAllSynchronizedGroups(ctx, apiv2, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_connection_directory_synchronized_groups", data, err)
	}

	result := multierror.Append(
		data.Set("connection_id", data.Id()),
		data.Set("group_ids", groupIDs),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := putSynchronizedGroups(ctx, apiv2, data.Id(), []string{}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func getAllSynchronizedGroups(ctx context.Context, apiv2 *managementv2client.Management, connectionID string) ([]string, error) {
	var groupIDs []string

	page, err := apiv2.Connections.DirectoryProvisioning.ListSynchronizedGroups(ctx, connectionID,
		&managementv2.ListSynchronizedGroupsRequestParameters{},
	)
	if err != nil {
		return nil, err
	}

	for _, g := range page.Results {
		groupIDs = append(groupIDs, g.GetID())
	}

	for {
		page, err = page.GetNextPage(ctx)
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return nil, err
		}
		for _, g := range page.Results {
			groupIDs = append(groupIDs, g.GetID())
		}
	}

	return groupIDs, nil
}

func putSynchronizedGroups(ctx context.Context, apiv2 *managementv2client.Management, connectionID string, groupIDs []string) error {
	payloadGroups := make([]*managementv2.SynchronizedGroupPayload, len(groupIDs))
	for i, id := range groupIDs {
		payloadGroups[i] = &managementv2.SynchronizedGroupPayload{
			ID: id,
		}
	}

	return apiv2.Connections.DirectoryProvisioning.Set(ctx, connectionID,
		&managementv2.ReplaceSynchronizedGroupsRequestContent{
			Groups: payloadGroups,
		},
	)
}
