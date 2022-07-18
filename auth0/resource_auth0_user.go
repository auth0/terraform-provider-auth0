package auth0

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

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "auth0|"+new
				},
			},
			"connection_name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"username": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"family_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"given_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"nickname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"password": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"email": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"email_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"verify_email": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"phone_number": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"phone_verified": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"user_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			"app_metadata": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: structure.SuppressJsonDiff,
			},
			"blocked": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"picture": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"roles": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
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
		d.Set("user_id", user.ID),
		d.Set("username", user.Username),
		d.Set("name", user.Name),
		d.Set("family_name", user.FamilyName),
		d.Set("given_name", user.GivenName),
		d.Set("nickname", user.Nickname),
		d.Set("email", user.Email),
		d.Set("email_verified", user.EmailVerified),
		d.Set("verify_email", user.VerifyEmail),
		d.Set("phone_number", user.PhoneNumber),
		d.Set("phone_verified", user.PhoneVerified),
		d.Set("blocked", user.Blocked),
		d.Set("picture", user.Picture),
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
	user, err := expandUser(d)
	if err != nil {
		return diag.FromErr(err)
	}

	api := m.(*management.Management)
	if err := api.User.Create(user); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(user.ID))

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
	user := &management.User{
		ID:            String(d, "user_id", IsNewResource()),
		Connection:    String(d, "connection_name"),
		Email:         String(d, "email", IsNewResource(), HasChange()),
		Name:          String(d, "name"),
		GivenName:     String(d, "given_name"),
		FamilyName:    String(d, "family_name"),
		Username:      String(d, "username", IsNewResource(), HasChange()),
		Nickname:      String(d, "nickname"),
		Password:      String(d, "password", IsNewResource(), HasChange()),
		PhoneNumber:   String(d, "phone_number", IsNewResource(), HasChange()),
		EmailVerified: Bool(d, "email_verified", IsNewResource(), HasChange()),
		VerifyEmail:   Bool(d, "verify_email", IsNewResource(), HasChange()),
		PhoneVerified: Bool(d, "phone_verified", IsNewResource(), HasChange()),
		Picture:       String(d, "picture"),
		Blocked:       Bool(d, "blocked"),
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
		return JSON(d, metadataType+"_metadata")
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
		roles = append(roles, auth0.StringValue(role.ID))
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
	toAdd, toRemove := Diff(d, "roles")

	if err := removeUserRoles(api, d.Id(), toRemove.List()); err != nil {
		return err
	}

	return assignUserRoles(api, d.Id(), toAdd.List())
}

func removeUserRoles(api *management.Management, userID string, userRolesToRemove []interface{}) error {
	var rmRoles []*management.Role
	for _, rmRole := range userRolesToRemove {
		role := &management.Role{ID: auth0.String(rmRole.(string))}
		rmRoles = append(rmRoles, role)
	}

	if len(rmRoles) == 0 {
		return nil
	}

	err := api.User.RemoveRoles(userID, rmRoles)
	if err != nil {
		// Ignore 404 errors as the role may have been deleted
		// prior to un-assigning them from the user.
		if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
			return nil
		}
	}

	return err
}

func assignUserRoles(api *management.Management, userID string, userRolesToAdd []interface{}) error {
	var addRoles []*management.Role
	for _, addRole := range userRolesToAdd {
		role := &management.Role{ID: auth0.String(addRole.(string))}
		addRoles = append(addRoles, role)
	}

	if len(addRoles) == 0 {
		return nil
	}

	return api.User.AssignRoles(userID, addRoles)
}

func userHasChange(u *management.User) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return u.String() != "{}"
}
