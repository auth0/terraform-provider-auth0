package user

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewRoleResource will return a new auth0_user_role (1:1) resource.
func NewRoleResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(_, oldVal, newVal string, _ *schema.ResourceData) bool {
					return oldVal == "auth0|"+newVal
				},
				Description: "ID of the user.",
			},
			"role_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the role assigned to the user.",
			},
			"role_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the role.",
			},
			"role_description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the role.",
			},
		},
		CreateContext: createUserRole,
		ReadContext:   readUserRole,
		DeleteContext: deleteUserRole,
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("user_id", "role_id"),
		},
		Description: "With this resource, you can manage assigned roles for a user.",
	}
}

func createUserRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	roleID := data.Get("role_id").(string)
	rolesToAssign := []*management.Role{{ID: &roleID}}

	if err := api.User.AssignRoles(ctx, userID, rolesToAssign); err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, userID, roleID)

	return readUserRole(ctx, data, meta)
}

func readUserRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)

	rolesList, err := api.User.Roles(ctx, userID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	roleID := data.Get("role_id").(string)
	for _, role := range rolesList.Roles {
		if role.GetID() == roleID {
			return diag.FromErr(flattenUserRole(data, role))
		}
	}

	data.SetId("")
	return nil
}

func deleteUserRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	roleID := data.Get("role_id").(string)
	rolesToRemove := []*management.Role{{ID: &roleID}}

	if err := api.User.RemoveRoles(ctx, userID, rolesToRemove); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
