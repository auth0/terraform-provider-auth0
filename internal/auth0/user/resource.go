package user

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
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
				DiffSuppressFunc: func(_, oldVal, newVal string, _ *schema.ResourceData) bool {
					return oldVal == "auth0|"+newVal
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
			"custom_domain_header": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Sets the `Auth0-Custom-Domain` header on all requests for this resource",
			},
		},
	}
}

func createUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	user, err := expandUser(data)
	if err != nil {
		return diag.FromErr(err)
	}

	var reqOptions []management.RequestOption
	customDomainHeader, ok := data.GetOk("custom_domain_header")
	if ok {
		reqOptions = append(reqOptions, management.CustomDomainHeader(customDomainHeader.(string)))
	}

	if err := api.User.Create(ctx, user, reqOptions...); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(user.GetID())

	return readUser(ctx, data, meta)
}

func readUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var reqOptions []management.RequestOption
	customDomainHeader, ok := data.GetOk("custom_domain_header")
	if ok {
		reqOptions = append(reqOptions, management.CustomDomainHeader(customDomainHeader.(string)))
	}
	user, err := api.User.Read(ctx, data.Id(), reqOptions...)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenUser(data, user))
}

func updateUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	user, err := expandUser(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = validateUser(user); err != nil {
		return diag.FromErr(err)
	}

	api := meta.(*config.Config).GetAPI()
	if userHasChange(user) {
		var reqOptions []management.RequestOption
		customDomainHeader, ok := data.GetOk("custom_domain_header")
		if ok {
			reqOptions = append(reqOptions, management.CustomDomainHeader(customDomainHeader.(string)))
		}
		if err := api.User.Update(ctx, data.Id(), user, reqOptions...); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	return readUser(ctx, data, meta)
}

func deleteUser(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	var reqOptions []management.RequestOption
	customDomainHeader, ok := data.GetOk("custom_domain_header")
	if ok {
		reqOptions = append(reqOptions, management.CustomDomainHeader(customDomainHeader.(string)))
	}

	if err := api.User.Delete(ctx, data.Id(), reqOptions...); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
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

func userHasChange(u *management.User) bool {
	// Hacky but we need to tell if an
	// empty json is sent to the api.
	return u.String() != "{}"
}
