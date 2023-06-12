package user

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewRoleResource will return a new auth0_user_role (1:1) resource.
func NewRoleResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "auth0|"+new
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
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage assigned roles for a user.",
	}
}

func createUserRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)
	data.SetId(userID)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(userID + "-roles")
	defer mutex.Unlock(userID + "-roles")

	roleID := data.Get("role_id").(string)
	if err := api.User.AssignRoles(userID, []*management.Role{{ID: &roleID}}); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readUserRole(ctx, data, meta)
}

func readUserRole(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)

	rolesList, err := api.User.Roles(userID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	roleID := data.Get("role_id").(string)
	for _, role := range rolesList.Roles {
		if role.GetID() == roleID {
			result := multierror.Append(
				data.Set("role_name", role.GetName()),
				data.Set("role_description", role.GetDescription()),
			)
			return diag.FromErr(result.ErrorOrNil())
		}
	}

	data.SetId("")
	return nil
}

func deleteUserRole(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	userID := data.Get("user_id").(string)

	mutex := meta.(*config.Config).GetMutex()
	mutex.Lock(userID + "-roles")
	defer mutex.Unlock(userID + "-roles")

	roleID := data.Get("role_id").(string)
	if err := api.User.RemoveRoles(userID, []*management.Role{{ID: &roleID}}); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	data.SetId("")

	return nil
}
