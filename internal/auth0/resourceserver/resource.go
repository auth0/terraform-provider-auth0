package resourceserver

import (
	"context"
	"fmt"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

const auth0ManagementAPI = "Auth0 Management API"

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
			"scopes": {
				Type:     schema.TypeSet,
				Optional: true,
				Deprecated: "Managing scopes through the `scopes` attribute is deprecated and it will be removed in a future major version. " +
					"Migrate to the `auth0_resource_server_scope` or `auth0_resource_server_scopes` resources to manage role scopes instead. " +
					"Check the [MIGRATION GUIDE](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md#resource-server-scopes) for more info.",
				Description: "List of permissions (scopes) used by this resource server." +
					"Managing scopes through the `scopes` attribute is deprecated and it will be removed in a future major version. " +
					"Migrate to the `auth0_resource_server_scope` or `auth0_resource_server_scopes` resources to manage role scopes instead. " +
					"Check the [MIGRATION GUIDE](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md#resource-server-scopes) for more info.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Name of the permission (scope). Examples include " +
								"`read:appointments` or `delete:appointments`.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the permission (scope).",
						},
					},
				},
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
					min := 16
					if len(v) < min {
						es = append(es, fmt.Errorf("expected length of %s to be at least %d, %q is %d", k, min, v, len(v)))
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
				ValidateFunc: validation.StringInSlice([]string{
					"access_token",
					"access_token_authz",
				}, true),
				Description: "Dialect of access tokens that should be issued for this resource server. " +
					"Options include `access_token` or `access_token_authz`. " +
					"If this setting is set to `access_token_authz`, the Permissions claim will be added to the access token. " +
					"Only available if RBAC (`enforce_policies`) is enabled for this API.",
			},
		},
	}
}

func createResourceServer(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	resourceServer := expandResourceServer(d)
	if err := api.ResourceServer.Create(resourceServer); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resourceServer.GetID())

	return readResourceServer(ctx, d, m)
}

func readResourceServer(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	resourceServer, err := api.ResourceServer.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Ensuring the ID is the resource server ID and not the identifier,
	// as both can be used to find a resource server with the Read() func.
	d.SetId(resourceServer.GetID())

	result := multierror.Append(
		d.Set("name", resourceServer.GetName()),
		d.Set("identifier", resourceServer.GetIdentifier()),
		d.Set("token_lifetime", resourceServer.GetTokenLifetime()),
		d.Set("allow_offline_access", resourceServer.GetAllowOfflineAccess()),
		d.Set("token_lifetime_for_web", resourceServer.GetTokenLifetimeForWeb()),
		d.Set("signing_alg", resourceServer.GetSigningAlgorithm()),
		d.Set("signing_secret", resourceServer.GetSigningSecret()),
		d.Set(
			"skip_consent_for_verifiable_first_party_clients",
			resourceServer.GetSkipConsentForVerifiableFirstPartyClients(),
		),
	)

	if resourceServer.GetName() != auth0ManagementAPI {
		result = multierror.Append(
			result,
			d.Set("verification_location", resourceServer.GetVerificationLocation()),
			d.Set("enforce_policies", resourceServer.GetEnforcePolicies()),
			d.Set("token_dialect", resourceServer.GetTokenDialect()),
			d.Set("scopes", flattenResourceServerScopes(resourceServer.GetScopes())),
		)
	}

	return diag.FromErr(result.ErrorOrNil())
}

func updateResourceServer(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	resourceServer := expandResourceServer(d)

	if err := api.ResourceServer.Update(d.Id(), resourceServer); err != nil {
		return diag.FromErr(err)
	}

	return readResourceServer(ctx, d, m)
}

func deleteResourceServer(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	if isManagementAPI(d.GetRawState()) {
		d.SetId("")
		return nil
	}

	api := m.(*config.Config).GetAPI()

	if err := api.ResourceServer.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func expandResourceServer(d *schema.ResourceData) *management.ResourceServer {
	config := d.GetRawConfig()

	resourceServer := &management.ResourceServer{
		TokenLifetime: value.Int(config.GetAttr("token_lifetime")),
		SkipConsentForVerifiableFirstPartyClients: value.Bool(
			config.GetAttr("skip_consent_for_verifiable_first_party_clients"),
		),
	}

	if d.IsNewResource() {
		resourceServer.Identifier = value.String(config.GetAttr("identifier"))
	}

	if !isManagementAPI(d.GetRawState()) {
		resourceServer.Name = value.String(config.GetAttr("name"))
		resourceServer.Scopes = expandResourceServerScopes(config.GetAttr("scopes"))
		resourceServer.SigningAlgorithm = value.String(config.GetAttr("signing_alg"))
		resourceServer.SigningSecret = value.String(config.GetAttr("signing_secret"))
		resourceServer.AllowOfflineAccess = value.Bool(config.GetAttr("allow_offline_access"))
		resourceServer.TokenLifetimeForWeb = value.Int(config.GetAttr("token_lifetime_for_web"))
		resourceServer.EnforcePolicies = value.Bool(config.GetAttr("enforce_policies"))
		resourceServer.TokenDialect = value.String(config.GetAttr("token_dialect"))
		resourceServer.VerificationLocation = value.String(config.GetAttr("verification_location"))
	}

	return resourceServer
}

func expandResourceServerScopes(scopes cty.Value) *[]management.ResourceServerScope {
	resourceServerScopes := make([]management.ResourceServerScope, 0)

	scopes.ForEachElement(func(_ cty.Value, scope cty.Value) (stop bool) {
		resourceServerScopes = append(resourceServerScopes, management.ResourceServerScope{
			Value:       value.String(scope.GetAttr("value")),
			Description: value.String(scope.GetAttr("description")),
		})

		return stop
	})

	return &resourceServerScopes
}

func flattenResourceServerScopes(resourceServerScopes []management.ResourceServerScope) []map[string]interface{} {
	scopes := make([]map[string]interface{}, len(resourceServerScopes))

	for index, scope := range resourceServerScopes {
		scopes[index] = map[string]interface{}{
			"value":       scope.Value,
			"description": scope.Description,
		}
	}

	return scopes
}

func isManagementAPI(state cty.Value) bool {
	if state.IsNull() {
		return false
	}

	return state.GetAttr("name").AsString() == auth0ManagementAPI
}
