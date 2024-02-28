package user

import (
	"context"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewRolesResource will return a new auth0_user_roles (1:many) resource.
func NewRolesResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(_, old, new string, _ *schema.ResourceData) bool {
					return old == "auth0|"+new
				},
				Description: "ID of the user.",
			},
			"roles": {
				Type:        schema.TypeSet,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of IDs of roles assigned to the user.",
			},
		},
		CreateContext: upsertUserRoles,
		ReadContext:   readUserRoles,
		UpdateContext: upsertUserRoles,
		DeleteContext: deleteUserRoles,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage assigned roles for a user.",
	}
}

func upsertUserRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	userID := data.Get("user_id").(string)
	data.SetId(userID)

	if err := persistUserRoles(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return readUserRoles(ctx, data, meta)
}

func readUserRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var roles []*management.Role
	var page int
	for {
		roleList, err := api.User.Roles(ctx, data.Id(), management.Page(page), management.PerPage(100))
		if err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}

		roles = append(roles, roleList.Roles...)

		if !roleList.HasNext() {
			break
		}

		page++
	}

	return diag.FromErr(flattenUserRoles(data, roles))
}

func deleteUserRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Id()

	userRolesToRemove := data.Get("roles").(*schema.Set).List()
	var rmRoles []*management.Role
	for _, rmRole := range userRolesToRemove {
		role := &management.Role{ID: auth0.String(rmRole.(string))}
		rmRoles = append(rmRoles, role)
	}

	if err := api.User.RemoveRoles(ctx, userID, rmRoles); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func persistUserRoles(ctx context.Context, data *schema.ResourceData, meta interface{}) error {
	if !data.HasChange("roles") {
		return nil
	}

	rolesToAdd, rolesToRemove := value.Difference(data, "roles")

	if err := removeUserRoles(ctx, meta, data.Id(), rolesToRemove); err != nil {
		if !internalError.IsStatusNotFound(err) {
			return err
		}
	}

	return assignUserRoles(ctx, meta, data.Id(), rolesToAdd)
}

func removeUserRoles(ctx context.Context, meta interface{}, userID string, userRolesToRemove []interface{}) error {
	if len(userRolesToRemove) == 0 {
		return nil
	}

	var rmRoles []*management.Role
	for _, rmRole := range userRolesToRemove {
		role := &management.Role{ID: auth0.String(rmRole.(string))}
		rmRoles = append(rmRoles, role)
	}

	api := meta.(*config.Config).GetAPI()

	return api.User.RemoveRoles(ctx, userID, rmRoles)
}

func assignUserRoles(ctx context.Context, meta interface{}, userID string, userRolesToAdd []interface{}) error {
	if len(userRolesToAdd) == 0 {
		return nil
	}

	var addRoles []*management.Role
	for _, addRole := range userRolesToAdd {
		roleID := addRole.(string)
		role := &management.Role{ID: &roleID}
		addRoles = append(addRoles, role)
	}

	api := meta.(*config.Config).GetAPI()

	return api.User.AssignRoles(ctx, userID, addRoles)
}
