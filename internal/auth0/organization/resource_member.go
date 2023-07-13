package organization

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewMemberResource will return a new auth0_organization_member resource.
func NewMemberResource() *schema.Resource {
	return &schema.Resource{
		Description:   "This resource is used to manage the assignment of members and their roles within an organization.",
		CreateContext: createOrganizationMember,
		ReadContext:   readOrganizationMember,
		DeleteContext: deleteOrganizationMember,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID(internalSchema.SeparatorColon, "organization_id", "user_id"),
		},
		Schema: map[string]*schema.Schema{
			"organization_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the organization to assign the member to.",
			},
			"user_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to add as an organization member.",
			},
		},
	}
}

func createOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	userID := d.Get("user_id").(string)
	organizationID := d.Get("organization_id").(string)

	if err := api.Organization.AddMembers(ctx, organizationID, []string{userID}); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(organizationID + ":" + userID)

	return readOrganizationMember(ctx, d, m)
}

func readOrganizationMember(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	organizationID := data.Get("organization_id").(string)

	members, err := api.Organization.Members(ctx, organizationID)
	if err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	userID := data.Get("user_id").(string)
	for _, member := range members.Members {
		if member.GetUserID() == userID {
			return nil
		}
	}

	data.SetId("")
	return nil
}

func deleteOrganizationMember(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	organizationID := d.Get("organization_id").(string)
	userID := d.Get("user_id").(string)

	if err := api.Organization.DeleteMembers(ctx, organizationID, []string{userID}); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
