package organization

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/auth0/go-auth0"
	managementv3 "github.com/auth0/go-auth0/v3/management"
	managementv3client "github.com/auth0/go-auth0/v3/management/client"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// organizationClientsBatchSize is the maximum number of clients the batch association
// endpoints accept in a single request.
const organizationClientsBatchSize = 10

// NewClientsResource will return a new auth0_organization_clients (1:many) resource (EA only).
func NewClientsResource() *schema.Resource {
	return &schema.Resource{
		Description: "With this resource, you can manage all of the client (application) associations of an " +
			"organization, controlling those applications' entitlement to the organization (EA only). " +
			"This resource is authoritative: it manages the full set of associations, so it must not be " +
			"used together with the `auth0_organization_client` resource on the same organization. " +
			"An organization accepts up to 100 associated clients.",
		CreateContext: createOrganizationClients,
		ReadContext:   readOrganizationClients,
		UpdateContext: updateOrganizationClients,
		DeleteContext: deleteOrganizationClients,
		CustomizeDiff: validateOrganizationClientsDiff,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to associate the clients (applications) with.",
			},
			"clients": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The clients (applications) associated with the organization.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"client_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the client (application) to associate with the organization.",
						},
						"use_for_member_access": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether this client is used for member access to the organization.",
						},
					},
				},
			},
		},
	}
}

func createOrganizationClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Get("organization_id").(string)

	desiredClients := expandOrganizationClients(data.Get("clients"))
	if err := validateOrganizationClientsAreUnique(desiredClients); err != nil {
		return diag.FromErr(err)
	}

	// Not HandleAPIError: there is nothing in state to remove at create time, so swallowing a
	// 404 would end the apply with no error and no ID, which Terraform blames on the provider.
	alreadyAssociatedClients, err := fetchAllOrganizationClients(ctx, apiv3, organizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	if diagnostics := guardAgainstErasingUnwantedOrganizationClients(
		organizationID,
		alreadyAssociatedClients,
		desiredClients,
	); diagnostics.HasError() {
		return diagnostics
	}

	data.SetId(organizationID)

	// Past the guard the organization is empty or already holds exactly the configured client
	// IDs, but those associations can still carry the wrong use_for_member_access, so
	// reconcile rather than skip the call.
	return readOrganizationClientsAfterApply(ctx, data, meta, applyOrganizationClientsDiff(
		ctx,
		apiv3,
		organizationID,
		diffOrganizationClients(alreadyAssociatedClients, desiredClients),
	))
}

func readOrganizationClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationClients, err := fetchAllOrganizationClients(ctx, apiv3, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_organization_clients", data, err)
	}

	return diag.FromErr(flattenOrganizationClients(data, organizationClients))
}

func updateOrganizationClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	organizationID := data.Id()

	desiredClients := expandOrganizationClients(data.Get("clients"))
	if err := validateOrganizationClientsAreUnique(desiredClients); err != nil {
		return diag.FromErr(err)
	}

	// Diffing against the API rather than prior state also reconciles out-of-band changes.
	// State would leave a rogue association for the read below to put back, and would re-post
	// a client already associated out-of-band, which is a 409 for the whole batch. As on the
	// create path, a 404 is reported instead of dropping the resource from state.
	currentClients, err := fetchAllOrganizationClients(ctx, apiv3, organizationID)
	if err != nil {
		return diag.FromErr(err)
	}

	return readOrganizationClientsAfterApply(ctx, data, meta, applyOrganizationClientsDiff(
		ctx,
		apiv3,
		organizationID,
		diffOrganizationClients(currentClients, desiredClients),
	))
}

func deleteOrganizationClients(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	clientsInState := expandOrganizationClients(data.Get("clients"))

	clientIDs := make([]string, 0, len(clientsInState))
	for _, client := range clientsInState {
		clientIDs = append(clientIDs, client.ClientID)
	}

	if err := removeOrganizationClients(ctx, apiv3, data.Id(), clientIDs); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

// validateOrganizationClientsDiff reports a duplicated client at plan time, when it can.
//
// The raw config is read instead of data.Get, which renders every client_id only known after
// apply as the same placeholder string, making distinct clients look like one listed twice.
// Unknown IDs are skipped here and left to the checks in create and update.
func validateOrganizationClientsDiff(_ context.Context, data *schema.ResourceDiff, _ interface{}) error {
	// Absent where there is no configuration at all, and GetAttr panics on such a value.
	rawConfig := data.GetRawConfig()
	if rawConfig.IsNull() || !rawConfig.IsKnown() {
		return nil
	}

	clients := rawConfig.GetAttr("clients")
	if clients.IsNull() || !clients.IsKnown() {
		return nil
	}

	var knownClients []*managementv3.CreateOrganizationClientRequestItem

	for iterator := clients.ElementIterator(); iterator.Next(); {
		_, client := iterator.Element()

		clientID := client.GetAttr("client_id")
		if clientID.IsNull() || !clientID.IsKnown() {
			continue
		}

		knownClients = append(knownClients, &managementv3.CreateOrganizationClientRequestItem{
			ClientID: clientID.AsString(),
		})
	}

	return validateOrganizationClientsAreUnique(knownClients)
}

// validateOrganizationClientsAreUnique rejects a `clients` set naming the same client twice,
// which only survives the set semantics when the entries disagree on use_for_member_access.
// The endpoint dedupes the payload and keeps the last item, so the organization would come
// back with fewer associations than configured and fail with an inconsistent-result error.
func validateOrganizationClientsAreUnique(clients []*managementv3.CreateOrganizationClientRequestItem) error {
	seen := make(map[string]struct{}, len(clients))

	for _, client := range clients {
		if _, duplicate := seen[client.ClientID]; duplicate {
			return fmt.Errorf(
				"client_id %q is listed more than once in `clients` with conflicting "+
					"`use_for_member_access` values, declare each client only once",
				client.ClientID,
			)
		}

		seen[client.ClientID] = struct{}{}
	}

	return nil
}

// readOrganizationClientsAfterApply refreshes state and, when the apply failed, reports that
// failure alongside it. Each batch request is atomic but the sequence of them is not, so a
// failure part way through needs the read to keep state truthful for the next plan.
func readOrganizationClientsAfterApply(
	ctx context.Context,
	data *schema.ResourceData,
	meta interface{},
	err error,
) diag.Diagnostics {
	if err == nil {
		return readOrganizationClients(ctx, data, meta)
	}

	return append(diag.FromErr(err), readOrganizationClients(ctx, data, meta)...)
}

// organizationClientsDiff holds the calls needed to take an organization's associations from
// what the API currently reports to what the configuration asks for.
type organizationClientsDiff struct {
	toAdd    []*managementv3.CreateOrganizationClientRequestItem
	toUpdate []*managementv3.CreateOrganizationClientRequestItem
	toRemove []string
}

// diffOrganizationClients compares per client ID, not through the set hash: the hash covers
// use_for_member_access, so flipping that flag would read as a removal plus an addition, and
// re-associating a still-associated client is answered with a 409.
func diffOrganizationClients(
	currentClients []*managementv3.OrganizationClient,
	desiredClients []*managementv3.CreateOrganizationClientRequestItem,
) organizationClientsDiff {
	useForMemberAccessByClientID := make(map[string]bool, len(currentClients))
	for _, client := range currentClients {
		useForMemberAccessByClientID[client.GetClientID()] = client.GetUseForMemberAccess()
	}

	var diff organizationClientsDiff

	desiredClientIDs := make(map[string]struct{}, len(desiredClients))
	for _, client := range desiredClients {
		desiredClientIDs[client.ClientID] = struct{}{}

		useForMemberAccess, associated := useForMemberAccessByClientID[client.ClientID]
		switch {
		case !associated:
			diff.toAdd = append(diff.toAdd, client)
		case useForMemberAccess != client.UseForMemberAccess:
			diff.toUpdate = append(diff.toUpdate, client)
		}
	}

	for _, client := range currentClients {
		if _, desired := desiredClientIDs[client.GetClientID()]; !desired {
			diff.toRemove = append(diff.toRemove, client.GetClientID())
		}
	}

	// The list endpoint has no stable ordering, so sort to keep an apply reproducible and to
	// keep clients from moving between batches.
	byClientID := func(first, second *managementv3.CreateOrganizationClientRequestItem) int {
		return strings.Compare(first.ClientID, second.ClientID)
	}
	slices.SortFunc(diff.toAdd, byClientID)
	slices.SortFunc(diff.toUpdate, byClientID)
	sort.Strings(diff.toRemove)

	return diff
}

func applyOrganizationClientsDiff(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
	diff organizationClientsDiff,
) error {
	// Disassociations are applied first so that swapping one client for another cannot trip
	// the limit of 100 associated clients per organization.
	if err := removeOrganizationClients(ctx, apiv3, organizationID, diff.toRemove); err != nil {
		return err
	}

	if err := addOrganizationClients(ctx, apiv3, organizationID, diff.toAdd); err != nil {
		return err
	}

	for _, client := range diff.toUpdate {
		updateReq := &managementv3.UpdateOrganizationClientRequestContent{
			UseForMemberAccess: auth0.Bool(client.UseForMemberAccess),
		}

		if _, err := apiv3.Organizations.Clients.Update(ctx, organizationID, client.ClientID, updateReq); err != nil {
			return err
		}
	}

	return nil
}

// addOrganizationClients associates the given clients with the organization, splitting the
// work into requests of at most organizationClientsBatchSize clients each.
func addOrganizationClients(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
	clients []*managementv3.CreateOrganizationClientRequestItem,
) error {
	for batch := range slices.Chunk(clients, organizationClientsBatchSize) {
		createReq := &managementv3.CreateOrganizationClientsRequestContent{
			Clients: batch,
		}

		if _, err := apiv3.Organizations.Clients.Create(ctx, organizationID, createReq); err != nil {
			return err
		}
	}

	return nil
}

// removeOrganizationClients disassociates the given clients from the organization, splitting
// the work into requests of at most organizationClientsBatchSize client IDs each.
func removeOrganizationClients(
	ctx context.Context,
	apiv3 *managementv3client.Management,
	organizationID string,
	clientIDs []string,
) error {
	for batch := range slices.Chunk(clientIDs, organizationClientsBatchSize) {
		deleteReq := &managementv3.DeleteOrganizationClientsRequestContent{
			Clients: batch,
		}

		if err := apiv3.Organizations.Clients.Delete(ctx, organizationID, deleteReq); err != nil {
			// The endpoint is idempotent for clients that are not associated and for clients
			// that no longer exist, so a 404 means the organization itself is gone.
			if internalError.IsStatusNotFound(err) {
				return nil
			}

			return err
		}
	}

	return nil
}

func guardAgainstErasingUnwantedOrganizationClients(
	organizationID string,
	alreadyAssociatedClients []*managementv3.OrganizationClient,
	clientsToAdd []*managementv3.CreateOrganizationClientRequestItem,
) diag.Diagnostics {
	if len(alreadyAssociatedClients) == 0 {
		return nil
	}

	// The list endpoint does not guarantee a stable ordering, so both sides get sorted
	// before they are compared.
	alreadyAssociatedClientIDs := make([]string, 0, len(alreadyAssociatedClients))
	for _, client := range alreadyAssociatedClients {
		alreadyAssociatedClientIDs = append(alreadyAssociatedClientIDs, client.GetClientID())
	}
	sort.Strings(alreadyAssociatedClientIDs)

	clientIDsToAdd := make([]string, 0, len(clientsToAdd))
	for _, client := range clientsToAdd {
		clientIDsToAdd = append(clientIDsToAdd, client.ClientID)
	}
	sort.Strings(clientIDsToAdd)

	if cmp.Equal(clientIDsToAdd, alreadyAssociatedClientIDs) {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization with non empty clients",
			Detail: cmp.Diff(clientIDsToAdd, alreadyAssociatedClientIDs) +
				fmt.Sprintf("\nThe organization already has clients associated with it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_organization_clients.<given-name> %s'.", organizationID),
		},
	}
}
