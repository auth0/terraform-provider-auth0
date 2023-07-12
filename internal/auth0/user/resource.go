package user

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

type validateUserFunc func(*management.User) error

// NewResource will return a new auth0_user resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createUser,
		ReadContext:   readUser,
		UpdateContext: updateUser,
		DeleteContext: deleteUser,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage user identities, including resetting passwords, " +
			"and creating, provisioning, blocking, and deleting users.",
		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "auth0|"+new
				},
				Description: "ID of the user.",
			},
			"connection_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the connection from which the user information was sourced.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Username of the user. Only valid if the connection requires a username.",
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Name of the user. This value can only be updated if the connection is a " +
					"database connection (using the Auth0 store), a passwordless connection (email or sms) or " +
					"has disabled 'Sync user profile attributes at each login'. For more information, see: " +
					"[Configure Identity Provider Connection for User Profile Updates](https://auth0.com/docs/manage-users/user-accounts/user-profiles/configure-connection-sync-with-auth0).",
			},
			"family_name": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Family name of the user. This value can only be updated if the connection is a " +
					"database connection (using the Auth0 store), a passwordless connection (email or sms) or " +
					"has disabled 'Sync user profile attributes at each login'. For more information, see: " +
					"[Configure Identity Provider Connection for User Profile Updates](https://auth0.com/docs/manage-users/user-accounts/user-profiles/configure-connection-sync-with-auth0).",
			},
			"given_name": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "Given name of the user. This value can only be updated if the connection is a " +
					"database connection (using the Auth0 store), a passwordless connection (email or sms) or " +
					"has disabled 'Sync user profile attributes at each login'. For more information, see: " +
					"[Configure Identity Provider Connection for User Profile Updates](https://auth0.com/docs/manage-users/user-accounts/user-profiles/configure-connection-sync-with-auth0).",
			},
			"nickname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Preferred nickname or alias of the user. This value can only be updated if the connection is a " +
					"database connection (using the Auth0 store), a passwordless connection (email or sms) or " +
					"has disabled 'Sync user profile attributes at each login'. For more information, see: " +
					"[Configure Identity Provider Connection for User Profile Updates](https://auth0.com/docs/manage-users/user-accounts/user-profiles/configure-connection-sync-with-auth0).",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Initial password for this user. Required for non-passwordless connections (SMS and email).",
			},
			"email": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Email address of the user.",
			},
			"email_verified": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the email address has been verified.",
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Indicates whether the user will receive a verification email after creation. " +
					"Overrides behavior of `email_verified` parameter.",
			},
			"phone_number": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Phone number for the user; follows the E.164 recommendation. Used for SMS connections. ",
			},
			"phone_verified": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the phone number has been verified.",
			},
			"user_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description: "Custom fields that store info about the user that does " +
					"not impact a user's core functionality. Examples include work address, home address, and user preferences.",
			},
			"app_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
				Description: "Custom fields that store info about the user that impact the user's core " +
					"functionality, such as how an application functions or what the user can access. " +
					"Examples include support plans and IDs for external accounts.",
			},
			"blocked": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the user is blocked or not.",
			},
			"picture": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				Description: "Picture of the user. This value can only be updated if the connection is a " +
					"database connection (using the Auth0 store), a passwordless connection (email or sms) or " +
					"has disabled 'Sync user profile attributes at each login'. For more information, see: " +
					"[Configure Identity Provider Connection for User Profile Updates](https://auth0.com/docs/manage-users/user-accounts/user-profiles/configure-connection-sync-with-auth0).",
			},
		},
	}
}

func readUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	user, err := api.User.Read(ctx, d.Id())
	if err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	err = flattenUser(d, user)

	return diag.FromErr(err)
}

func createUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	user, err := expandUser(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.User.Create(ctx, user); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.GetID())

	if err = persistUserRoles(ctx, d, m); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return readUser(ctx, d, m)
		}

		return diag.FromErr(err)
	}

	return readUser(ctx, d, m)
}

func updateUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	user, err := expandUser(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = validateUser(user); err != nil {
		return diag.FromErr(err)
	}

	api := m.(*config.Config).GetAPI()
	if userHasChange(user) {
		if err := api.User.Update(ctx, d.Id(), user); err != nil {
			return diag.FromErr(err)
		}
	}

	if err = persistUserRoles(ctx, d, m); err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return readUser(ctx, d, m)
		}

		return diag.FromErr(err)
	}

	return readUser(ctx, d, m)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	if err := api.User.Delete(ctx, d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func validateUser(user *management.User) error {
	validations := []validateUserFunc{
		validateNoUsernameAndPasswordSimultaneously(),
		validateNoUsernameAndEmailVerifiedSimultaneously(),
		validateNoPasswordAndEmailVerifiedSimultaneously(),
	}

	var result *multierror.Error
	for _, validationFunc := range validations {
		if err := validationFunc(user); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func validateNoUsernameAndPasswordSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		if user.Username != nil && user.Password != nil {
			return fmt.Errorf("cannot update username and password simultaneously")
		}
		return nil
	}
}

func validateNoUsernameAndEmailVerifiedSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		if user.Username != nil && user.EmailVerified != nil {
			return fmt.Errorf("cannot update username and email_verified simultaneously")
		}
		return nil
	}
}

func validateNoPasswordAndEmailVerifiedSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		if user.Password != nil && user.EmailVerified != nil {
			return fmt.Errorf("cannot update password and email_verified simultaneously")
		}
		return nil
	}
}

func persistUserRoles(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	if !d.HasChange("roles") {
		return nil
	}

	rolesToAdd, rolesToRemove := value.Difference(d, "roles")

	if err := removeUserRoles(ctx, meta, d.Id(), rolesToRemove); err != nil {
		return err
	}

	return assignUserRoles(ctx, meta, d.Id(), rolesToAdd)
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

func userHasChange(u *management.User) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return u.String() != "{}"
}
