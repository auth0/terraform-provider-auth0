package organization

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewMembersResource will return a new auth0_organization_members (1:many) resource.
func NewMembersResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage members of an organization.",
		CreateContext: createOrganizationMembers,
		ReadContext:   readOrganizationMembers,
		UpdateContext: updateOrganizationMembers,
		DeleteContext: deleteOrganizationMembers,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to assign the members to.",
			},
			"members": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Add user ID(s) directly from the tenant to become members of the organization.",
			},
		},
	}
}

func createOrganizationMembers(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)

	alreadyMembers, err := fetchAllOrganizationMembers(ctx, api, organizationID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(organizationID)

	membersToAdd := *value.Strings(data.GetRawConfig().GetAttr("members"))

	if diagnostics := guardAgainstErasingUnwantedMembers(
		organizationID,
		alreadyMembers,
		membersToAdd,
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if len(membersToAdd) > len(alreadyMembers) {
		if err := api.Organization.AddMembers(ctx, organizationID, membersToAdd); err != nil {
			return diag.FromErr(err)
		}
	}

	return readOrganizationMembers(ctx, data, meta)
}

func readOrganizationMembers(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	members, err := fetchAllOrganizationMembers(ctx, api, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenOrganizationMembers(data, members))
}

func updateOrganizationMembers(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()

	toAdd, toRemove := value.Difference(data, "members")

	removeMembers := make([]string, 0)
	for _, member := range toRemove {
		removeMembers = append(removeMembers, member.(string))
	}

	if len(removeMembers) > 0 {
		err := api.Organization.DeleteMembers(ctx, organizationID, removeMembers)
		if !internalError.IsStatusNotFound(err) {
			return diag.FromErr(err)
		}
	}

	addMembers := make([]string, 0)
	for _, member := range toAdd {
		addMembers = append(addMembers, member.(string))
	}

	if len(addMembers) > 0 {
		if err := api.Organization.AddMembers(ctx, organizationID, addMembers); err != nil {
			return diag.FromErr(err)
		}
	}

	return readOrganizationMembers(ctx, data, meta)
}

func deleteOrganizationMembers(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	organizationID := data.Id()
	membersToRemove := *value.Strings(data.GetRawState().GetAttr("members"))

	if len(membersToRemove) == 0 {
		return nil
	}

	if err := api.Organization.DeleteMembers(ctx, organizationID, membersToRemove); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func guardAgainstErasingUnwantedMembers(
	organizationID string,
	alreadyMembers []management.OrganizationMember,
	memberIDsToAdd []string,
) diag.Diagnostics {
	if len(alreadyMembers) == 0 {
		return nil
	}

	alreadyMemberIDs := make([]string, 0)
	for _, member := range alreadyMembers {
		alreadyMemberIDs = append(alreadyMemberIDs, member.GetUserID())
	}

	if cmp.Equal(memberIDsToAdd, alreadyMemberIDs) {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization with non empty members",
			Detail: cmp.Diff(memberIDsToAdd, alreadyMemberIDs) +
				fmt.Sprintf("\nThe organization already has members attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_organization_members.<given-name> %s'.", organizationID),
		},
	}
}
