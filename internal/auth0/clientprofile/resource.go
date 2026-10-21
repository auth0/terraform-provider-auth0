package clientprofile

import (
	"context"
	"errors"
	"fmt"

	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

const (
	resourceType = "auth0_client_profile"

	// MaxLifetime is the largest refresh token lifetime accepted by the API, in seconds (one year).
	maxLifetime = 31557600

	tooManyEntitiesCode = "too_many_entities"
	preLoginPrompt      = string(management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnumPreLoginPrompt)
)

var (
	applyPolicyTypes = internalSchema.ToStrSlice(
		management.ClientProfileApplyPolicyExternalMetadataTypeEnumCimd,
	)
	applyPolicyCreatedBy = internalSchema.ToStrSlice(
		management.ClientProfileApplyPolicyExternalMetadataCreatedByEnumClient,
	)
	organizationUsages = internalSchema.ToStrSlice(
		management.ClientProfileClientConfigurationOrganizationUsageEnumDeny,
		management.ClientProfileClientConfigurationOrganizationUsageEnumAllow,
		management.ClientProfileClientConfigurationOrganizationUsageEnumRequire,
	)
	organizationRequireBehaviors = internalSchema.ToStrSlice(
		management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnumNoPrompt,
		management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnumPreLoginPrompt,
		management.ClientProfileClientConfigurationOrganizationRequireBehaviorEnumPostLoginPrompt,
	)
	organizationDiscoveryMethods = internalSchema.ToStrSlice(
		management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnumEmail,
		management.ClientProfileClientConfigurationOrganizationDiscoveryMethodEnumOrganizationName,
	)
)

// NewResource will return a new auth0_client_profile resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createClientProfile,
		ReadContext:   readClientProfile,
		UpdateContext: updateClientProfile,
		DeleteContext: deleteClientProfile,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateClientProfile,
		Description: "With this resource, you can manage the client profile (application profile) of a tenant. " +
			"A client profile applies runtime policy to dynamically registered clients, such as clients " +
			"registered through a Client ID Metadata Document (CIMD). A tenant can have only one client " +
			"profile; if one already exists, import it instead of creating a new one. " +
			"Destroying this resource deletes the tenant's client profile, which immediately removes the " +
			"governance it applies to all matched clients. The docs recommend `prevent_destroy`, and the " +
			"resource requires the `create:client_profiles`, `read:client_profiles`, `update:client_profiles` " +
			"and `delete:client_profiles` scopes. " +
			"This feature is available in Early Access only. (EA only)",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 128),
				Description:  "The name of the client profile.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 256),
				Description: "The description of the client profile. Removing the attribute clears " +
					"the description in Auth0.",
			},
			"apply_policy": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Runtime matching rules that determine which clients this profile governs. " +
					"Defaults to dynamically registered CIMD clients.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"external_metadata_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(applyPolicyTypes, false),
							Description: fmt.Sprintf(
								"The type of external metadata of the clients to match. Defaults to `%s`. Possible values: %v.",
								applyPolicyTypes[0], applyPolicyTypes,
							),
						},
						"external_metadata_created_by": {
							Type:         schema.TypeString,
							Optional:     true,
							Computed:     true,
							ValidateFunc: validation.StringInSlice(applyPolicyCreatedBy, false),
							Description: fmt.Sprintf(
								"The creator of the clients to match. Defaults to `%s`, which means the client was registered dynamically. Possible values: %v.",
								applyPolicyCreatedBy[0], applyPolicyCreatedBy,
							),
						},
					},
				},
			},
			"client_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Description: "Runtime policy applied to matched clients. Omitted settings fall back to each client's " +
					"own setting. Auth0 replaces this block as a whole, so removing it clears all of its settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"refresh_token": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "Refresh token settings applied to matched clients. Auth0 replaces this block " +
								"as a whole, so removing it clears all of its settings.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"token_lifetime": envelope(
										"The absolute lifetime of a refresh token. If only one of `token_lifetime` and "+
											"`idle_token_lifetime` is set, Auth0 derives the other.",
										true,
										&schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntBetween(1, maxLifetime),
											Description:  "The lifetime in seconds.",
										},
									),
									"idle_token_lifetime": envelope(
										"The inactivity lifetime of a refresh token. It must not exceed `token_lifetime`. "+
											"If only one of `token_lifetime` and `idle_token_lifetime` is set, Auth0 derives the other.",
										true,
										&schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntBetween(1, maxLifetime),
											Description:  "The lifetime in seconds.",
										},
									),
									"leeway": envelope(
										"The period in which a refresh token may be reused without triggering reuse detection.",
										false,
										&schema.Schema{
											Type:         schema.TypeInt,
											ValidateFunc: validation.IntAtLeast(0),
											Description:  "The leeway in seconds. `0` disables the leeway.",
										},
									),
								},
							},
						},
						"organization_usage": envelope(
							"Whether matched clients can use organizations.",
							false,
							&schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(organizationUsages, false),
								Description:  fmt.Sprintf("Possible values: %v.", organizationUsages),
							},
						),
						"organization_require_behavior": envelope(
							"Defines how to proceed during an authentication transaction when the organization usage is `require`.",
							false,
							&schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validation.StringInSlice(organizationRequireBehaviors, false),
								Description:  fmt.Sprintf("Possible values: %v.", organizationRequireBehaviors),
							},
						),
						"organization_discovery_methods": envelope(
							fmt.Sprintf(
								"Methods used to discover the organization during login. Requires the "+
									"`organization_require_behavior` default value to be `%s`.", preLoginPrompt,
							),
							false,
							&schema.Schema{
								Type:     schema.TypeSet,
								MinItems: 1,
								Elem: &schema.Schema{
									Type:         schema.TypeString,
									ValidateFunc: validation.StringInSlice(organizationDiscoveryMethods, false),
								},
								Description: fmt.Sprintf("Possible values: %v.", organizationDiscoveryMethods),
							},
						),
						"require_proof_of_possession": envelope(
							"Whether matched clients must use proof of possession (DPoP) for their tokens.",
							false,
							&schema.Schema{
								Type:        schema.TypeBool,
								Description: "Whether proof of possession is required.",
							},
						),
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the client profile was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the client profile was updated.",
			},
		},
	}
}

// envelope returns a single-item block that mirrors the API's `{ "default_value": X }` wrapper, which
// keeps the schema aligned with the API should the wrapper gain more keys. When computed is set,
// Auth0 may fill in the block itself, as it does for the derived refresh token lifetimes.
func envelope(description string, computed bool, value *schema.Schema) *schema.Schema {
	if computed {
		value.Optional = true
		value.Computed = true
	} else {
		value.Required = true
	}

	return &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Computed:    computed,
		MaxItems:    1,
		Description: description,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{"default_value": value},
		},
	}
}

func createClientProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	profile, err := apiv3.ClientProfiles.Create(ctx, expandClientProfileCreate(data.GetRawConfig()))
	if err != nil {
		return diag.FromErr(mapCreateError(err))
	}

	data.SetId(profile.GetID())

	return readClientProfile(ctx, data, meta)
}

func readClientProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	profile, err := apiv3.ClientProfiles.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError(resourceType, data, err)
	}

	return flattenClientProfile(data, profile)
}

func updateClientProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	if _, err := apiv3.ClientProfiles.Update(ctx, data.Id(), expandClientProfileUpdate(data.GetRawConfig())); err != nil {
		return diag.FromErr(err)
	}

	return readClientProfile(ctx, data, meta)
}

func deleteClientProfile(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv3 := meta.(*config.Config).GetAPIV3()

	// Deleting is idempotent in the API, but a 404 is treated as success too.
	if err := apiv3.ClientProfiles.Delete(ctx, data.Id()); err != nil && !internalError.IsStatusNotFound(err) {
		return diag.FromErr(err)
	}

	data.SetId("")

	return nil
}

// mapCreateError turns the errors returned when a profile already exists into an actionable message.
// Other 403s are returned untouched, as a missing feature flag or scope cannot be told apart by status.
func mapCreateError(err error) error {
	var conflict *management.ConflictError
	if errors.As(err, &conflict) || internalError.ForbiddenErrorCode(err) == tooManyEntitiesCode {
		return fmt.Errorf(
			"a client profile already exists for this tenant; import it with `terraform import %s.<name> <profile_id>` "+
				"or an `import {}` block instead of creating a new one: %w", resourceType, err,
		)
	}

	return err
}

// validateClientProfile performs plan-time validation for UX only, as the API enforces the same rules.
func validateClientProfile(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	return checkClientProfileConfig(diff.GetRawConfig())
}

// checkClientProfileConfig enforces the cross-field rules the API also enforces. Values that are
// unknown at plan time are skipped, as they cannot be judged yet.
func checkClientProfileConfig(raw cty.Value) error {
	configuration := value.FirstBlock(value.Attr(raw, "client_configuration"))
	block := value.FirstBlock(value.Attr(configuration, "refresh_token"))
	tokenLifetime := value.IntIfSet(envelopeValue(block, "token_lifetime"))
	idleTokenLifetime := value.IntIfSet(envelopeValue(block, "idle_token_lifetime"))
	if tokenLifetime != nil && idleTokenLifetime != nil && *idleTokenLifetime > *tokenLifetime {
		return fmt.Errorf(
			"`client_configuration.refresh_token.idle_token_lifetime` (%d) must not exceed "+
				"`client_configuration.refresh_token.token_lifetime` (%d)",
			*idleTokenLifetime, *tokenLifetime,
		)
	}

	methods := envelopeValue(configuration, "organization_discovery_methods")
	behavior := envelopeValue(configuration, "organization_require_behavior")
	if value.IsSet(methods) && methods.LengthInt() > 0 && behavior.IsKnown() {
		if behavior.IsNull() || behavior.AsString() != preLoginPrompt {
			return fmt.Errorf("`client_configuration.organization_discovery_methods` requires "+
				"`client_configuration.organization_require_behavior` to be %q", preLoginPrompt)
		}
	}

	return nil
}
