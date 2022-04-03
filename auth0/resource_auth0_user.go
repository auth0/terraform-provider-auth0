package auth0

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

type validateUserFunc func(*management.User) error

func newUser() *schema.Resource {
	return &schema.Resource{
		Create: createUser,
		Read:   readUser,
		Update: updateUser,
		Delete: deleteUser,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return old == "auth0|"+new
				},
				StateFunc: func(s interface{}) string {
					return strings.ToLower(s.(string))
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

func readUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	user, err := api.User.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
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

	userMeta, err := structure.FlattenJsonToString(user.UserMetadata)
	if err != nil {
		return err
	}
	result = multierror.Append(result, d.Set("user_metadata", userMeta))

	appMeta, err := structure.FlattenJsonToString(user.AppMetadata)
	if err != nil {
		return err
	}
	result = multierror.Append(result, d.Set("app_metadata", appMeta))

	roleList, err := api.User.Roles(d.Id())
	if err != nil {
		return err
	}
	result = multierror.Append(
		result,
		d.Set("roles", func() []interface{} {
			var roles []interface{}
			for _, role := range roleList.Roles {
				roles = append(roles, auth0.StringValue(role.ID))
			}
			return roles
		}()),
	)

	return result.ErrorOrNil()
}

func createUser(d *schema.ResourceData, m interface{}) error {
	user, err := buildUser(d)
	if err != nil {
		return err
	}

	api := m.(*management.Management)
	if err := api.User.Create(user); err != nil {
		return err
	}
	d.SetId(auth0.StringValue(user.ID))

	d.Partial(true)
	if err = assignUserRoles(d, m); err != nil {
		return err
	}
	d.Partial(false)

	return readUser(d, m)
}

func updateUser(d *schema.ResourceData, m interface{}) error {
	user, err := buildUser(d)
	if err != nil {
		return err
	}

	if err = validateUser(user); err != nil {
		return err
	}

	api := m.(*management.Management)
	if userHasChange(user) {
		if err := api.User.Update(d.Id(), user); err != nil {
			return err
		}
	}

	d.Partial(true)
	if err = assignUserRoles(d, m); err != nil {
		return fmt.Errorf("failed assigning user roles. %s", err)
	}
	d.Partial(false)

	return readUser(d, m)
}

func deleteUser(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.User.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func buildUser(d *schema.ResourceData) (*management.User, error) {
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

	var err error
	user.UserMetadata, err = JSON(d, "user_metadata")
	if err != nil {
		return nil, err
	}

	user.AppMetadata, err = JSON(d, "app_metadata")
	if err != nil {
		return nil, err
	}

	return user, nil
}

func validateUser(user *management.User) error {
	var result *multierror.Error
	validations := []validateUserFunc{
		validateNoUsernameAndPasswordSimultaneously(),
		validateNoUsernameAndEmailVerifiedSimultaneously(),
		validateNoPasswordAndEmailVerifiedSimultaneously(),
	}
	for _, validationFunc := range validations {
		if err := validationFunc(user); err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

func validateNoUsernameAndPasswordSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		var err error
		if user.Username != nil && user.Password != nil {
			err = fmt.Errorf("cannot update username and password simultaneously")
		}
		return err
	}
}

func validateNoUsernameAndEmailVerifiedSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		var err error
		if user.Username != nil && user.EmailVerified != nil {
			err = fmt.Errorf("cannot update username and email_verified simultaneously")
		}
		return err
	}
}

func validateNoPasswordAndEmailVerifiedSimultaneously() validateUserFunc {
	return func(user *management.User) error {
		var err error
		if user.Password != nil && user.EmailVerified != nil {
			err = fmt.Errorf("cannot update password and email_verified simultaneously")
		}
		return err
	}
}

func assignUserRoles(d *schema.ResourceData, m interface{}) error {
	add, rm := Diff(d, "roles")

	var addRoles []*management.Role
	for _, addRole := range add.List() {
		addRoles = append(
			addRoles,
			&management.Role{
				ID: auth0.String(addRole.(string)),
			},
		)
	}

	var rmRoles []*management.Role
	for _, rmRole := range rm.List() {
		rmRoles = append(
			rmRoles,
			&management.Role{
				ID: auth0.String(rmRole.(string)),
			},
		)
	}

	api := m.(*management.Management)

	if len(rmRoles) > 0 {
		if err := api.User.RemoveRoles(d.Id(), rmRoles); err != nil {
			// Ignore 404 errors as the role may have been deleted
			// prior to un-assigning them from the user.
			if mErr, ok := err.(management.Error); ok {
				if mErr.Status() != http.StatusNotFound {
					return err
				}
			} else {
				return err
			}
		}
	}

	if len(addRoles) > 0 {
		if err := api.User.AssignRoles(d.Id(), addRoles); err != nil {
			return err
		}
	}

	d.SetPartial("roles")

	return nil
}

func userHasChange(u *management.User) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return u.String() != "{}"
}
