package connection

import (
	"context"
	"errors"
	"fmt"

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
		CustomizeDiff: rejectDuplicateGroupsByID,
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
			"group_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"groups"},
				Deprecated:    "Use `groups` instead, which exposes each group's name, email and member count alongside its ID.",
				Description:   "IDs of the Google Workspace Directory groups to synchronize.",
			},
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
							Optional:    true,
							Description: "Google Workspace Directory group name.",
						},
						"email": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Google Workspace Directory group email.",
						},
						"direct_members_count": {
							Type:        schema.TypeInt,
							Optional:    true,
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

// rejectDuplicateGroupsByID fails the plan on a repeated group ID with different metadata.
func rejectDuplicateGroupsByID(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	seenGroupIDs := make(map[string]struct{})

	for _, groupID := range groupIDsIn(diff.Get("groups")) {
		if _, duplicate := seenGroupIDs[groupID]; duplicate {
			return fmt.Errorf("group %q is declared more than once in `groups`", groupID)
		}

		seenGroupIDs[groupID] = struct{}{}
	}

	return nil
}

// createDirectorySynchronizedGroups replaces the whole set.
func createDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()
	connectionID := data.Get("connection_id").(string)

	if err := putGroups(ctx, apiv3, connectionID, expandDirectorySynchronizedGroupsCreate(data)); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(connectionID)

	return readDirectorySynchronizedGroups(ctx, data, meta)
}

// updateDirectorySynchronizedGroups writes the delta, removing before adding so that a group whose
// metadata changed is gone by the time the add rewrites it.
func updateDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()
	connectionID := data.Get("connection_id").(string)

	update := expandDirectorySynchronizedGroupsUpdate(data)

	if err := removeGroups(ctx, apiv3, connectionID, update.remove); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	if err := addGroups(ctx, apiv3, connectionID, update.add); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readDirectorySynchronizedGroups(ctx, data, meta)
}

// metadataDiffers compares one group's declared metadata against what the last read stored. A field
// the configuration omits counts as empty, so deleting one is a change: leaving it out of the add
// payload is what clears it. A group state does not hold differs as soon as anything is declared for
// it, since indexing the absent entry yields the zero values.
func metadataDiffers(group *managementv3.SynchronizedGroupPayload, priorGroup map[string]interface{}) bool {
	priorName, _ := priorGroup["name"].(string)
	priorEmail, _ := priorGroup["email"].(string)
	priorDirectMembersCount, _ := priorGroup["direct_members_count"].(int)

	return group.GetName() != priorName ||
		group.GetEmail() != priorEmail ||
		group.GetDirectMembersCount() != priorDirectMembersCount
}

// existedGroupsInState builds a payload for each `groups` block existing in the state.
func existedGroupsInState(data *schema.ResourceData) map[string]map[string]interface{} {
	priorGroups, _ := data.GetChange("groups")

	groupsByID := make(map[string]map[string]interface{})

	priorGroupSet, ok := priorGroups.(*schema.Set)
	if !ok {
		return groupsByID
	}

	for _, rawGroup := range priorGroupSet.List() {
		group, ok := rawGroup.(map[string]interface{})
		if !ok {
			continue
		}

		if groupID, ok := group["id"].(string); ok {
			groupsByID[groupID] = group
		}
	}

	return groupsByID
}

// groupIDsChange returns the group IDs in state and the ones the configuration asks for.
func groupIDsChange(data *schema.ResourceData) (prior []string, desired []string) {
	priorGroupIDs, desiredGroupIDs := data.GetChange("group_ids")
	priorGroups, desiredGroups := data.GetChange("groups")

	return mergeGroupIDs(priorGroupIDs, priorGroups),
		mergeGroupIDs(desiredGroupIDs, desiredGroups)
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
		flattenDirectorySynchronizedGroups(data, groups),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteDirectorySynchronizedGroups(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	if err := putGroups(ctx, apiv3, data.Id(), []*managementv3.SynchronizedGroupPayload{}); err != nil {
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

func putGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, groups []*managementv3.SynchronizedGroupPayload) error {
	return apiv3.Connections.DirectoryProvisioning.Set(ctx, connectionID,
		&managementv3.ReplaceSynchronizedGroupsRequestContent{
			Groups: groups,
		},
	)
}

// addGroups writes each group's metadata alongside its ID, since add stores what the payload holds
// and clears what it leaves out. Re-adding a synchronized group creates no duplicate, so a retry
// after a partial failure is safe.
func addGroups(ctx context.Context, apiv3 *managementv3client.Management, connectionID string, groups []*managementv3.SynchronizedGroupPayload) error {
	for _, chunk := range chunkGroups(groups) {
		if err := apiv3.Connections.DirectoryProvisioning.AddSynchronizedGroupSelections(ctx, connectionID,
			&managementv3.AddSynchronizedGroupsRequestContent{
				Groups: chunk,
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
	for _, chunk := range chunkGroups(groupIDs) {
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

// chunkGroups splits a write into calls the add and remove endpoints accept, whether it carries
// whole groups or the IDs alone.
func chunkGroups[T any](groups []T) [][]T {
	var chunks [][]T

	for start := 0; start < len(groups); start += groupsWriteChunkSize {
		chunks = append(chunks, groups[start:min(start+groupsWriteChunkSize, len(groups))])
	}

	return chunks
}
