package provider

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

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

type validateUserFunc func(*management.User) error

func newUser() *schema.Resource {
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the user.",
			},
			"family_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Family name of the user.",
			},
			"given_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Given name of the user.",
			},
			"nickname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Preferred nickname or alias of the user.",
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Picture of the user.",
			},
			"roles": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Set of IDs of roles assigned to the user.",
			},
		},
	}
}

func readUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	user, err := api.User.Read(d.Id())
	if err != nil {
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("user_id", user.GetID()),
		d.Set("username", user.GetUsername()),
		d.Set("name", user.GetName()),
		d.Set("family_name", user.GetFamilyName()),
		d.Set("given_name", user.GetGivenName()),
		d.Set("nickname", user.GetNickname()),
		d.Set("email", user.GetEmail()),
		d.Set("email_verified", user.GetEmailVerified()),
		d.Set("verify_email", user.GetVerifyEmail()),
		d.Set("phone_number", user.GetPhoneNumber()),
		d.Set("phone_verified", user.GetPhoneVerified()),
		d.Set("blocked", user.GetBlocked()),
		d.Set("picture", user.GetPicture()),
	)

	var userMeta string
	if user.UserMetadata != nil {
		userMeta, err = structure.FlattenJsonToString(*user.UserMetadata)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	result = multierror.Append(result, d.Set("user_metadata", userMeta))

	var appMeta string
	if user.AppMetadata != nil {
		appMeta, err = structure.FlattenJsonToString(*user.AppMetadata)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	result = multierror.Append(result, d.Set("app_metadata", appMeta))

	roleList, err := api.User.Roles(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	result = multierror.Append(result, d.Set("roles", flattenUserRoles(roleList)))

	return diag.FromErr(result.ErrorOrNil())
}

func createUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	user, err := expandUser(d)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.User.Create(user); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.GetID())

	if err = updateUserRoles(d, api); err != nil {
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

	api := m.(*management.Management)
	if userHasChange(user) {
		if err := api.User.Update(d.Id(), user); err != nil {
			return diag.FromErr(err)
		}
	}

	if err = updateUserRoles(d, api); err != nil {
		return diag.Errorf("failed assigning user roles. %s", err)
	}

	return readUser(ctx, d, m)
}

func deleteUser(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.User.Delete(d.Id()); err != nil {
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

func expandUser(d *schema.ResourceData) (*management.User, error) {
	config := d.GetRawConfig()

	user := &management.User{
		Connection: value.String(config.GetAttr("connection_name")),
		Name:       value.String(config.GetAttr("name")),
		GivenName:  value.String(config.GetAttr("given_name")),
		FamilyName: value.String(config.GetAttr("family_name")),
		Nickname:   value.String(config.GetAttr("nickname")),
		Picture:    value.String(config.GetAttr("picture")),
		Blocked:    value.Bool(config.GetAttr("blocked")),
	}

	if d.IsNewResource() {
		user.ID = value.String(config.GetAttr("user_id"))
	}
	if d.IsNewResource() || d.HasChange("email") {
		user.Email = value.String(config.GetAttr("email"))
	}
	if d.IsNewResource() || d.HasChange("username") {
		user.Username = value.String(config.GetAttr("username"))
	}
	if d.IsNewResource() || d.HasChange("password") {
		user.Password = value.String(config.GetAttr("password"))
	}
	if d.IsNewResource() || d.HasChange("phone_number") {
		user.PhoneNumber = value.String(config.GetAttr("phone_number"))
	}
	if d.IsNewResource() || d.HasChange("email_verified") {
		user.EmailVerified = value.Bool(config.GetAttr("email_verified"))
	}
	if d.IsNewResource() || d.HasChange("verify_email") {
		user.VerifyEmail = value.Bool(config.GetAttr("verify_email"))
	}
	if d.IsNewResource() || d.HasChange("phone_verified") {
		user.PhoneVerified = value.Bool(config.GetAttr("phone_verified"))
	}

	if d.HasChange("user_metadata") {
		userMetadata, err := expandMetadata(d, "user")
		if err != nil {
			return nil, err
		}
		user.UserMetadata = &userMetadata
	}

	if d.HasChange("app_metadata") {
		appMetadata, err := expandMetadata(d, "app")
		if err != nil {
			return nil, err
		}
		user.AppMetadata = &appMetadata
	}

	return user, nil
}

func expandMetadata(d *schema.ResourceData, metadataType string) (map[string]interface{}, error) {
	oldMetadata, newMetadata := d.GetChange(metadataType + "_metadata")
	if oldMetadata == "" {
		return value.MapFromJSON(d.GetRawConfig().GetAttr(metadataType + "_metadata"))
	}

	if newMetadata == "" {
		return map[string]interface{}{}, nil
	}

	oldMap, err := structure.ExpandJsonFromString(oldMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	newMap, err := structure.ExpandJsonFromString(newMetadata.(string))
	if err != nil {
		return map[string]interface{}{}, err
	}

	for key := range oldMap {
		if _, ok := newMap[key]; !ok {
			newMap[key] = nil
		}
	}

	return newMap, nil
}

func flattenUserRoles(roleList *management.RoleList) []interface{} {
	var roles []interface{}
	for _, role := range roleList.Roles {
		roles = append(roles, role.GetID())
	}
	return roles
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

func updateUserRoles(d *schema.ResourceData, api *management.Management) error {
	if !d.HasChange("roles") {
		return nil
	}

	oldValue, newValue := d.GetChange("roles")

	rolesToAdd := newValue.(*schema.Set).Difference(oldValue.(*schema.Set))
	rolesToRemove := oldValue.(*schema.Set).Difference(newValue.(*schema.Set))

	if err := removeUserRoles(api, d.Id(), rolesToRemove.List()); err != nil {
		return err
	}

	return assignUserRoles(api, d.Id(), rolesToAdd.List())
}

func removeUserRoles(api *management.Management, userID string, userRolesToRemove []interface{}) error {
	if len(userRolesToRemove) == 0 {
		return nil
	}

	var rmRoles []*management.Role
	for _, rmRole := range userRolesToRemove {
		role := &management.Role{ID: auth0.String(rmRole.(string))}
		rmRoles = append(rmRoles, role)
	}

	err := api.User.RemoveRoles(userID, rmRoles)
	if err != nil {
		// Ignore 404 errors as the role may have been deleted prior to un-assigning them from the user.
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}
	}

	return err
}

func assignUserRoles(api *management.Management, userID string, userRolesToAdd []interface{}) error {
	if len(userRolesToAdd) == 0 {
		return nil
	}

	var addRoles []*management.Role
	for _, addRole := range userRolesToAdd {
		roleID := addRole.(string)
		role := &management.Role{ID: &roleID}
		addRoles = append(addRoles, role)
	}

	return api.User.AssignRoles(userID, addRoles)
}

func userHasChange(u *management.User) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return u.String() != "{}"
}
