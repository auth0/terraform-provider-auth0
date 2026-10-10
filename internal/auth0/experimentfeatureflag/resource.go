package experimentfeatureflag

import (
	"context"
	"fmt"
	"strings"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var (
	// ParameterTypes are the allowed value types for a feature flag parameter.
	parameterTypes = []string{
		string(management.FeatureFlagConfigParamTypeEnumBoolean),
		string(management.FeatureFlagConfigParamTypeEnumString),
		string(management.FeatureFlagConfigParamTypeEnumNumber),
		string(management.FeatureFlagConfigParamTypeEnumArray),
		string(management.FeatureFlagConfigParamTypeEnumObject),
	}
	// StatusValues are the allowed lifecycle states of a feature flag.
	statusValues = []string{
		string(management.FeatureFlagStatusEnumDraft),
		string(management.FeatureFlagStatusEnumActive),
		string(management.FeatureFlagStatusEnumArchived),
	}
)

// NewResource will return a new auth0_experiment_feature_flag resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createFeatureFlag,
		ReadContext:   readFeatureFlag,
		UpdateContext: updateFeatureFlag,
		DeleteContext: deleteFeatureFlag,
		CustomizeDiff: validateFeatureFlagStatusOnCreate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Create and manage an Experiment Center feature flag (EA only).",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the feature flag.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the feature flag. Min length 3 when present.",
			},
			"parameters": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				Description: "The parameters carried by the feature flag.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The key of the parameter.",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(parameterTypes, false),
							Description:  "The value type of the parameter. One of " + strings.Join(parameterTypes, ", "),
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The default value of the parameter, as a string. For `type = \"object\"` or " +
								"`type = \"array\"` supply a JSON-encoded string; for `boolean`/`number` supply the literal value as a string.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the parameter.",
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(statusValues, false),
				Description: fmt.Sprintf(
					"The lifecycle status of the feature flag. One of %s. On creation, "+
						"omit this attribute or set it to `%s`. Can be updated to other value after creation.",
					strings.Join(statusValues, ", "), management.FeatureFlagStatusEnumDraft,
				),
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the feature flag: `self` for user-created flags or `auth0` for built-in flags.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the feature flag was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the feature flag was updated.",
			},
		},
	}
}

func validateFeatureFlagStatusOnCreate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	// ResourceDiff has no IsNewResource; an empty ID marks a create or replacement plan.
	if diff.Id() != "" {
		return nil
	}

	config := diff.GetRawConfig()
	if !config.IsKnown() || config.IsNull() {
		return nil
	}

	status := config.GetAttr("status")
	if !status.IsKnown() || status.IsNull() || status.AsString() == string(management.FeatureFlagStatusEnumDraft) {
		return nil
	}

	return fmt.Errorf("cannot set status %q when creating a feature flag; "+
		"omit it or use %q. You can update the status after creation",
		status.AsString(), management.FeatureFlagStatusEnumDraft)
}

func createFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandFeatureFlag(data)
	if err != nil {
		return diag.FromErr(err)
	}

	created, err := api.Experimentation.FeatureFlags.Create(ctx, request)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(created.GetID())

	return readFeatureFlag(ctx, data, meta)
}

func readFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	flag, err := api.Experimentation.FeatureFlags.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_experiment_feature_flag", data, err)
	}

	return diag.FromErr(flattenFeatureFlag(data, flag))
}

func updateFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandFeatureFlagUpdate(data)
	if err != nil {
		return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
	}

	_, err = api.Experimentation.FeatureFlags.Update(ctx, data.Id(), request)
	return append(diag.FromErr(internalError.HandleAPIError(data, err)), readFeatureFlag(ctx, data, meta)...)
}

func deleteFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	err := api.Experimentation.FeatureFlags.Delete(ctx, data.Id())
	return diag.FromErr(internalError.HandleAPIError(data, err))
}
