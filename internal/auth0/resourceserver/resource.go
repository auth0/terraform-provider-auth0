package resourceserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

const auth0ManagementAPIName = "Auth0 Management API"

// NewResource will return a new auth0_resource_server resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createResourceServer,
		ReadContext:   readResourceServer,
		UpdateContext: updateResourceServer,
		DeleteContext: deleteResourceServer,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can set up APIs that can be consumed from your authorized applications.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Friendly name for the resource server. Cannot include `<` or `>` characters.",
			},
			"identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "Unique identifier for the resource server. Used as the audience parameter " +
					"for authorization calls. Cannot be changed once set.",
			},
			"signing_alg": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Algorithm used to sign JWTs. Options include `HS256` and `RS256`.",
			},
			"signing_secret": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: func(i interface{}, k string) (s []string, es []error) {
					v, ok := i.(string)
					if !ok {
						es = append(es, fmt.Errorf("expected type of %s to be string", k))
						return
					}
					minLength := 16
					if len(v) < minLength {
						es = append(es, fmt.Errorf("expected length of %s to be at least %d, %q is %d", k, minLength, v, len(v)))
					}
					return
				},
				Description: "Secret used to sign tokens when using symmetric algorithms (HS256).",
			},
			"allow_offline_access": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether refresh tokens can be issued for this resource server.",
			},
			"token_lifetime": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Description: "Number of seconds during which access tokens issued for this resource server " +
					"from the token endpoint remain valid.",
			},
			"token_lifetime_for_web": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				Description: "Number of seconds during which access tokens issued for this resource server via " +
					"implicit or hybrid flows remain valid. Cannot be greater than the `token_lifetime` value.",
			},
			"skip_consent_for_verifiable_first_party_clients": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether to skip user consent for applications flagged as first party.",
			},
			"verification_location": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "URL from which to retrieve JWKs for this resource server. " +
					"Used for verifying the JWT sent to Auth0 for token introspection.",
			},
			"enforce_policies": {
				Type:     schema.TypeBool,
				Computed: true,
				Optional: true,
				Description: "If this setting is enabled, RBAC authorization policies will be enforced for this API. " +
					"Role and permission assignments will be evaluated during the login transaction.",
			},
			"token_dialect": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"access_token",
					"access_token_authz",
					"rfc9068_profile",
					"rfc9068_profile_authz",
				}, true),
				Description: "Dialect of access tokens that should be issued for this resource server. " +
					"Options include `access_token`, `rfc9068_profile`, `access_token_authz`, and `rfc9068_profile_authz`. " +
					"`access_token` is a JWT containing standard Auth0 claims. `rfc9068_profile` is a JWT conforming to the IETF JWT Access Token Profile. " +
					"`access_token_authz` is a JWT containing standard Auth0 claims, including RBAC permissions claims. `rfc9068_profile_authz` is a JWT conforming to the IETF JWT Access Token Profile, including RBAC permissions claims. " +
					"RBAC permissions claims are available if RBAC (`enforce_policies`) is enabled for this API. " +
					"For more details, refer to [Access Token Profiles](https://auth0.com/docs/secure/tokens/access-tokens/access-token-profiles).",
			},
			"consent_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"transactional-authorization-with-mfa",
					"null",
				}, true),
				Description: "Consent policy for this resource server. " +
					"Options include `transactional-authorization-with-mfa`, or `null` to disable.",
			},
			"authorization_details": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "Authorization details for this resource server.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Type of authorization details.",
						},
						"disable": {
							Type:          schema.TypeBool,
							Optional:      true,
							Description:   "Disable authorization details.",
						},
					},
				},
			},
		},
	}
}

func createResourceServer(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer := expandResourceServer(data)

	if err := api.ResourceServer.Create(ctx, resourceServer); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(resourceServer.GetID())

	config := data.GetRawConfig()
	if isConsentPolicyNull(config) {
		type nilableResourceServer struct {
			ConsentPolicy *string `json:"consent_policy"`
		}
		if err := api.Request(ctx, http.MethodPatch, api.URI("resource-servers", data.Id()), nilableResourceServer{
			ConsentPolicy: nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if isAuthorizationDetailsNull(config) {
		type nilableResourceServer struct {
			AuthorizationDetails *[]management.ResourceServerAuthorizationDetails `json:"authorization_details"`
		}
		if err := api.Request(ctx, http.MethodPatch, api.URI("resource-servers", data.Id()), nilableResourceServer{
			AuthorizationDetails: nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return readResourceServer(ctx, data, meta)
}

func readResourceServer(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer, err := api.ResourceServer.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	// Ensuring the ID is the resource server ID and not the identifier,
	// as both can be used to find a resource server with the Read() func.
	data.SetId(resourceServer.GetID())

	return diag.FromErr(flattenResourceServer(data, resourceServer))
}

func updateResourceServer(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	resourceServer := expandResourceServer(data)

	if err := api.ResourceServer.Update(ctx, data.Id(), resourceServer); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	config := data.GetRawConfig()
	if isConsentPolicyNull(config) {
		type nilableResourceServer struct {
			ConsentPolicy *string `json:"consent_policy"`
		}
		if err := api.Request(ctx, http.MethodPatch, api.URI("resource-servers", data.Id()), nilableResourceServer{
			ConsentPolicy: nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	if isAuthorizationDetailsNull(config) {
		type nilableResourceServer struct {
			AuthorizationDetails *[]management.ResourceServerAuthorizationDetails `json:"authorization_details"`
		}
		if err := api.Request(ctx, http.MethodPatch, api.URI("resource-servers", data.Id()), nilableResourceServer{
			AuthorizationDetails: nil,
		}); err != nil {
			return diag.FromErr(err)
		}
	}

	return readResourceServer(ctx, data, meta)
}

func deleteResourceServer(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if resourceServerIsAuth0ManagementAPI(data.GetRawState()) {
		return nil
	}

	api := meta.(*config.Config).GetAPI()

	if err := api.ResourceServer.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
