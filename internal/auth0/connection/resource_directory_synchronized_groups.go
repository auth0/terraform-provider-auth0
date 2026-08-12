package connection

import (
	"context"
	"errors"

	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/auth0/go-auth0/v3/management/core"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewDirectorySynchronizedGroupsResource will return a new auth0_connection_directory_synchronized_groups resource.
func NewDirectorySynchronizedGroupsResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createDirectorySynchronizedGroups,
		ReadContext:   readDirectorySynchronizedGroups,
		UpdateContext: updateDirectorySynchronizedGroups,
		DeleteContext: deleteDirectorySynchronizedGroups,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage the set of Google Workspace groups " +
			"synchronized via directory provisioning for an Auth0 connection.",
		Schema: map[string]*schema.Schema{
			"connection_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the connection for which to manage synchronized groups.",
			},
			// Superseded by `groups`, but kept Optional so existing configurations keep working.
			// `ConflictsWith` rather than `ExactlyOneOf`, since setting neither is how every group
			// is unsynchronized.
			"group_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"groups"},
				Deprecated:    "Use `groups` instead, which exposes each group's name, email and member count alongside its ID.",
				Description:   "IDs of the Google Workspace Directory groups to synchronize.",
			},
			// `TypeSet` because the API treats this as a selection set: adding an
			// already-synchronized group yields one entry, and reads are unordered.
			//
			// TODO: the metadata is Computed-only because it is not established whether a directory
			// sync overwrites what a write stored. The feature team has been asked; revisit whether
			// it should also be settable once they answer.
			"groups": {
				Type:          schema.TypeSet,
				Optional:      true,
				ConflictsWith: []string{"group_ids"},
				Description:   "Google Workspace Directory groups to synchronize.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Google Workspace Directory group ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Google Workspace Directory group name.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Google Workspace Directory group email.",
						},
						"direct_members_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of direct members in the Google Workspace Directory group.",
						},
					},
				},
			},
		},
	}
}

// groupsWriteChunkSize is the most groups the add and remove endpoints accept in one call.
const groupsWriteChunkSize = 100

// createDirectorySynchronizedGroups replaces the whole set.
func createDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()
	connectionID := data.Get("connection_id").(string)

	if err := putGroups(ctx, apiv3, connectionID, configuredGroupIDs(data)); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(connectionID)

	return readDirectorySynchronizedGroups(ctx, data, meta)
}

// updateDirectorySynchronizedGroups touches only the groups that changed, so appending one group to
// a directory of 800 sends one rather than re-sending all 801 through a replace.
func updateDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()
	connectionID := data.Get("connection_id").(string)

	groupIDsToAdd, groupIDsToRemove := diffGroupIDs(groupIDsChange(data))

	if err := removeGroups(ctx, apiv3, connectionID, groupIDsToRemove); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if err := addGroups(ctx, apiv3, connectionID, groupIDsToAdd); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readDirectorySynchronizedGroups(ctx, data, meta)
}

func configuredGroupIDs(data *schema.ResourceData) []string {
	return mergeGroupIDs(data.Get("group_ids"), data.Get("groups"))
}

// groupIDsChange returns the group IDs in state and the ones the configuration asks for.
func groupIDsChange(data *schema.ResourceData) (prior []string, desired []string) {
	priorGroupIDs, desiredGroupIDs := data.GetChange("group_ids")
	priorGroups, desiredGroups := data.GetChange("groups")

	return mergeGroupIDs(priorGroupIDs, priorGroups),
		mergeGroupIDs(desiredGroupIDs, desiredGroups)
}

// mergeGroupIDs collects the IDs held by both args. The two conflict in
// configuration, so at most one ever contributes.
func mergeGroupIDs(rawGroupIDs interface{}, rawGroups interface{}) []string {
	groupIDs := make([]string, 0)

	if groupIDSet, ok := rawGroupIDs.(*schema.Set); ok {
		for _, rawGroupID := range groupIDSet.List() {
			if groupID, ok := rawGroupID.(string); ok && groupID != "" {
				groupIDs = append(groupIDs, groupID)
			}
		}
	}

	if groupSet, ok := rawGroups.(*schema.Set); ok {
		for _, rawGroup := range groupSet.List() {
			group, ok := rawGroup.(map[string]interface{})
			if !ok {
				continue
			}
			if groupID, ok := group["id"].(string); ok && groupID != "" {
				groupIDs = append(groupIDs, groupID)
			}
		}
	}

	return groupIDs
}

func diffGroupIDs(prior []string, desired []string) (toAdd []string, toRemove []string) {
	priorSet := make(map[string]struct{}, len(prior))
	for _, groupID := range prior {
		priorSet[groupID] = struct{}{}
	}

	desiredSet := make(map[string]struct{}, len(desired))
	for _, groupID := range desired {
		desiredSet[groupID] = struct{}{}
	}

	for _, groupID := range desired {
		if _, alreadySynchronized := priorSet[groupID]; !alreadySynchronized {
			toAdd = append(toAdd, groupID)
		}
	}

	for _, groupID := range prior {
		if _, stillWanted := desiredSet[groupID]; !stillWanted {
			toRemove = append(toRemove, groupID)
		}
	}

	return toAdd, toRemove
}

func readDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	groups, err := getAllGroups(ctx, apiv3, data.Id(), "")
	if err != nil {
		return internalError.HandleReadAPIError("auth0_connection_directory_synchronized_groups", data, err)
	}

	result := multierror.Append(
		data.Set("connection_id", data.Id()),
		setGroups(data, groups),
	)

	return diag.FromErr(result.ErrorOrNil())
}

// setGroups writes to whichever of `group_ids` and `groups` the configuration uses, never both.
// `groups` is the default, so an import lands on the attribute that is not deprecated.
func setGroups(data *schema.ResourceData, groups []*managementv3.SynchronizedGroupPayload) error {
	if priorGroupIDs, ok := data.Get("group_ids").(*schema.Set); ok && priorGroupIDs.Len() > 0 {
		return data.Set("group_ids", flattenGroupIDs(groups))
	}

	return data.Set("groups", flattenGroups(groups))
}

func flattenGroupIDs(groups []*managementv3.SynchronizedGroupPayload) []string {
	groupIDs := make([]string, 0, len(groups))

	for _, group := range groups {
		groupIDs = append(groupIDs, group.GetID())
	}

	return groupIDs
}

func flattenGroups(groups []*managementv3.SynchronizedGroupPayload) []interface{} {
	flattenedGroups := make([]interface{}, 0, len(groups))

	for _, group := range groups {
		flattenedGroups = append(flattenedGroups, map[string]interface{}{
			"id":                   group.GetID(),
			"name":                 group.GetName(),
			"email":                group.GetEmail(),
			"direct_members_count": group.GetDirectMembersCount(),
		})
	}

	return flattenedGroups
}

func deleteDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	if err := putGroups(ctx, apiv3, data.Id(), []string{}); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func getAllGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, query string) ([]*managementv3.SynchronizedGroupPayload, error) {
	var groups []*managementv3.SynchronizedGroupPayload

	requestParameters := &managementv3.ListSynchronizedGroupsRequestParameters{}
	if query != "" {
		requestParameters.Q = &query
	}

	page, err := apiv3.Connections.DirectoryProvisioning.ListSynchronizedGroups(ctx, connectionID,
		requestParameters,
	)
	if err != nil {
		return nil, err
	}

	groups = append(groups, page.Results...)

	for {
		page, err = page.GetNextPage(ctx)
		if err != nil {
			if errors.Is(err, core.ErrNoPages) {
				break
			}
			return nil, err
		}
		groups = append(groups, page.Results...)
	}

	return groups, nil
}

func putGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, groupIDs []string) error {
	payloadGroups := make([]*managementv3.SynchronizedGroupPayload, len(groupIDs))
	for i, id := range groupIDs {
		payloadGroups[i] = &managementv3.SynchronizedGroupPayload{
			ID: id,
		}
	}

	return apiv3.Connections.DirectoryProvisioning.Set(ctx, connectionID,
		&managementv3.ReplaceSynchronizedGroupsRequestContent{
			Groups: payloadGroups,
		},
	)
}

// addGroups is safe to retry after a partial failure: adding an already-synchronized group is a
// no-op.
func addGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, groupIDs []string) error {
	for _, chunk := range chunkGroupIDs(groupIDs) {
		payloadGroups := make([]*managementv3.SynchronizedGroupPayload, len(chunk))
		for i, id := range chunk {
			payloadGroups[i] = &managementv3.SynchronizedGroupPayload{
				ID: id,
			}
		}

		if err := apiv3.Connections.DirectoryProvisioning.AddSynchronizedGroupSelections(ctx, connectionID,
			&managementv3.AddSynchronizedGroupsRequestContent{
				Groups: payloadGroups,
			},
		); err != nil {
			return err
		}
	}

	return nil
}

// removeGroups is safe to retry after a partial failure: removing a group that is not synchronized
// is a no-op rather than a 404.
func removeGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, groupIDs []string) error {
	for _, chunk := range chunkGroupIDs(groupIDs) {
		selections := make([]*managementv3.SynchronizedGroupSelectionID, len(chunk))
		for i, id := range chunk {
			selections[i] = &managementv3.SynchronizedGroupSelectionID{
				ID: id,
			}
		}

		if err := apiv3.Connections.DirectoryProvisioning.DeleteSynchronizedGroupSelections(ctx, connectionID,
			&managementv3.DeleteSynchronizedGroupsRequestContent{
				Groups: selections,
			},
		); err != nil {
			return err
		}
	}

	return nil
}

func chunkGroupIDs(groupIDs []string) [][]string {
	var chunks [][]string

	for start := 0; start < len(groupIDs); start += groupsWriteChunkSize {
		chunks = append(chunks, groupIDs[start:min(start+groupsWriteChunkSize, len(groupIDs))])
	}

	return chunks
}
