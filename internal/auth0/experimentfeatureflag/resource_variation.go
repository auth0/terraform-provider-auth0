package experimentfeatureflag

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
)

// NewVariationResource returns the auth0_experiment_feature_flag_variation resource.
func NewVariationResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createVariation,
		ReadContext:   readVariation,
		UpdateContext: updateVariation,
		DeleteContext: deleteVariation,
		Description:   "Create and manage an Experiment Center feature flag variation (EA only).",
		Importer: &schema.ResourceImporter{
			StateContext: internalSchema.ImportResourceGroupID("feature_flag_id", "variation_id"),
		},
		CustomizeDiff: validateUniqueOverrideNames,
		Schema: map[string]*schema.Schema{
			"feature_flag_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the feature flag that owns the variation.",
			},
			"variation_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the variation within its feature flag.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the variation. Must be unique within the feature flag and contain 3 to 255 characters without surrounding whitespace.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the variation. Must contain 3 to 1024 characters when set.",
			},
			"overrides": {
				Type:        schema.TypeSet,
				Optional:    true,
				Set:         hashOverride,
				Description: "Parameter overrides for the variation. Omit for a control variation.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of a parameter on the parent feature flag. Each name can appear once.",
						},
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.StringInSlice(parameterTypes, false),
							Description:  "The type of the parameter value. One of " + strings.Join(parameterTypes, ", ") + ".",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The override value as a string. For object and array values, provide JSON; for boolean and number values, provide their literal text.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the variation was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the variation was updated.",
			},
		},
	}
}

func createVariation(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandVariation(data)
	if err != nil {
		return diag.FromErr(err)
	}

	featureFlagID := data.Get("feature_flag_id").(string)
	created, err := api.Experimentation.FeatureFlags.Variations.Create(ctx, featureFlagID, request)
	if err != nil {
		return diag.FromErr(err)
	}

	internalSchema.SetResourceGroupID(data, featureFlagID, created.GetID())
	if err := data.Set("variation_id", created.GetID()); err != nil {
		return diag.FromErr(err)
	}

	return readVariation(ctx, data, meta)
}

func readVariation(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	variation, err := api.Experimentation.FeatureFlags.Variations.Get(
		ctx,
		data.Get("feature_flag_id").(string),
		data.Get("variation_id").(string),
	)
	if err != nil {
		return internalError.HandleReadAPIError("auth0_experiment_feature_flag_variation", data, err)
	}

	return diag.FromErr(flattenVariation(data, variation))
}

func updateVariation(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandVariationUpdate(data)
	if err != nil {
		return append(diag.FromErr(err), readVariation(ctx, data, meta)...)
	}

	_, err = api.Experimentation.FeatureFlags.Variations.Update(
		ctx,
		data.Get("feature_flag_id").(string),
		data.Get("variation_id").(string),
		request,
	)
	return append(diag.FromErr(internalError.HandleAPIError(data, err)), readVariation(ctx, data, meta)...)
}

func deleteVariation(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	err := api.Experimentation.FeatureFlags.Variations.Delete(
		ctx,
		data.Get("feature_flag_id").(string),
		data.Get("variation_id").(string),
	)
	return diag.FromErr(internalError.HandleAPIError(data, err))
}

// hashOverride hashes the typed value instead of its text, so "1.0" and "1", "true" and "True"
// or JSON with different spacing, do not cause a diff.
func hashOverride(v interface{}) int {
	override := v.(map[string]interface{})
	name, _ := override["name"].(string)
	paramType, _ := override["type"].(string)
	value, _ := override["value"].(string)

	if typed, err := stringToTyped(paramType, value); err == nil {
		if canonical, err := typedToString(typed); err == nil {
			value = canonical
		}
	}

	return schema.HashString(name + "\x00" + paramType + "\x00" + value)
}

func overrideNames(overrides *schema.Set) []string {
	names := make([]string, 0, overrides.Len())
	for _, item := range overrides.List() {
		names = append(names, item.(map[string]interface{})["name"].(string))
	}
	return names
}

// validateUniqueOverrideNames rejects two override blocks for the same
// parameter, because the API receives overrides as a map keyed by name.
func validateUniqueOverrideNames(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	seen := make(map[string]bool)
	for _, name := range overrideNames(diff.Get("overrides").(*schema.Set)) {
		if seen[name] {
			return fmt.Errorf("override parameter %q is configured more than once", name)
		}
		seen[name] = true
	}

	return nil
}
