package organization

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
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
	mutex := meta.(*config.Config).GetMutex()

	organizationID := data.Get("organization_id").(string)

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	alreadyMembers, err := api.Organization.Members(organizationID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId(organizationID)

	membersToAdd := *value.Strings(data.GetRawConfig().GetAttr("members"))

	if diagnostics := guardAgainstErasingUnwantedMembers(
		organizationID,
		alreadyMembers.Members,
		membersToAdd,
	); diagnostics.HasError() {
		data.SetId("")
		return diagnostics
	}

	if err := api.Organization.AddMembers(organizationID, membersToAdd); err != nil {
		return diag.FromErr(err)
	}

	return readOrganizationMembers(ctx, data, meta)
}

func readOrganizationMembers(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	members, err := api.Organization.Members(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("organization_id", data.Id()),
		data.Set("members", flattenOrganizationMembers(members.Members)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateOrganizationMembers(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	organizationID := data.Id()

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	toAdd, toRemove := value.Difference(data, "members")

	removeMembers := make([]string, 0)
	for _, member := range toRemove {
		removeMembers = append(removeMembers, member.(string))
	}

	if len(removeMembers) > 0 {
		if err := api.Organization.DeleteMember(organizationID, removeMembers); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}
	}

	addMembers := make([]string, 0)
	for _, member := range toAdd {
		addMembers = append(addMembers, member.(string))
	}

	if len(addMembers) > 0 {
		if err := api.Organization.AddMembers(organizationID, addMembers); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}
	}

	return readOrganizationMembers(ctx, data, meta)
}

func deleteOrganizationMembers(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	organizationID := data.Id()
	membersToRemove := *value.Strings(data.GetRawState().GetAttr("members"))

	if len(membersToRemove) == 0 {
		return nil
	}

	mutex.Lock(organizationID)
	defer mutex.Unlock(organizationID)

	if err := api.Organization.DeleteMember(organizationID, membersToRemove); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}

func guardAgainstErasingUnwantedMembers(
	organizationID string,
	alreadyMembers []management.OrganizationMember,
	membersToAdd []string,
) diag.Diagnostics {
	if len(alreadyMembers) == 0 {
		return nil
	}

	return diag.Diagnostics{
		diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Organization with non empty members",
			Detail: cmp.Diff(membersToAdd, alreadyMembers) +
				fmt.Sprintf("\nThe organization already has members attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_organization_members.<given-name> %s'.", organizationID),
		},
	}
}

func flattenOrganizationMembers(members []management.OrganizationMember) []string {
	if len(members) == 0 {
		return nil
	}

	flattenedMembers := make([]string, 0)
	for _, member := range members {
		flattenedMembers = append(flattenedMembers, member.GetUserID())
	}

	return flattenedMembers
}
