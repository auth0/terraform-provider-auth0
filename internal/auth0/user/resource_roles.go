package user

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewRolesResource will return a new auth0_user_roles (1:many) resource.
func NewRolesResource() *schema.Resource {
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

	if err := persistUserRoles(data, meta); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	return readUserRoles(ctx, data, meta)
}

func readUserRoles(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	rolesList, err := api.User.Roles(data.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	var userRoles []string
	for _, role := range rolesList.Roles {
		userRoles = append(userRoles, role.GetID())
	}

	result := multierror.Append(
		data.Set("user_id", data.Id()),
		data.Set("roles", userRoles),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func deleteUserRoles(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	mutex := meta.(*config.Config).GetMutex()

	userID := data.Id()
	mutex.Lock(userID)
	defer mutex.Unlock(userID)

	userRolesToRemove := data.Get("roles").(*schema.Set).List()
	var rmRoles []*management.Role
	for _, rmRole := range userRolesToRemove {
		role := &management.Role{ID: auth0.String(rmRole.(string))}
		rmRoles = append(rmRoles, role)
	}

	if err := api.User.RemoveRoles(userID, rmRoles); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}

		return diag.FromErr(err)
	}

	return nil
}
