package role

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_role resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createRole,
		UpdateContext: updateRole,
		ReadContext:   readRole,
		DeleteContext: deleteRole,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateRole,
		Description: "With this resource, you can create and manage collections of permissions that can be " +
			"assigned to users, which are otherwise known as roles. Permissions (scopes) are created on " +
			"`auth0_resource_server`, then associated with roles and optionally, users using this resource.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the role.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the role.",
			},
			"type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice([]string{"tenant", "organization"}, false),
				Description: "The type of the role. Defaults to `tenant`, for a role available across the " +
					"whole tenant. Set to `organization` to scope the role to a single organization, in " +
					"which case `owner_id` must also be set. The Management API only accepts this field " +
					"on creation, so changing it forces a new role to be created. (EA only)",
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Description: "The ID of the organization owning the role. Only applicable when `type` is " +
					"`organization`, and required in that case. The Management API only accepts this " +
					"field on creation, so changing it forces a new role to be created. (EA only)",
			},
		},
	}
}

func validateRole(_ context.Context, data *schema.ResourceDiff, _ interface{}) error {
	cfg := data.GetRawConfig()
	if cfg.IsNull() {
		return nil
	}

	// The owner_id is usually interpolated from an organization that does not
	// exist yet, so we can only check whether it is set, not what it holds.
	roleType := cfg.GetAttr("type")
	ownerIDIsSet := !cfg.GetAttr("owner_id").IsNull()

	if roleType.IsNull() || !roleType.IsKnown() {
		if ownerIDIsSet {
			return fmt.Errorf("type must be set to organization when owner_id is set")
		}

		return nil
	}

	switch roleType.AsString() {
	case "organization":
		if !ownerIDIsSet {
			return fmt.Errorf("owner_id is required when type is set to organization")
		}
	default:
		if ownerIDIsSet {
			return fmt.Errorf("owner_id can only be set when type is set to organization")
		}
	}

	return nil
}

func createRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role := expandRole(data)

	if err := api.Role.Create(ctx, role); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(role.GetID())

	return readRole(ctx, data, meta)
}

func readRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role, err := api.Role.Read(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_role", data, err)
	}

	return diag.FromErr(flattenRole(data, role))
}

func updateRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	role := expandRole(data)

	if err := api.Role.Update(ctx, data.Id(), role); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readRole(ctx, data, meta)
}

func deleteRole(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Role.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
