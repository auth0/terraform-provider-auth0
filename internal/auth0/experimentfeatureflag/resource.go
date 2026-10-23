package experimentfeatureflag

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"

	management "github.com/auth0/go-auth0/v3/management"
	managementv3 "github.com/auth0/go-auth0/v3/management/client"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
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
		CustomizeDiff: customdiff.All(
			validateUniqueVariationAndOverrideNames,
			validateVariationCountOnActivation,
			customdiff.ComputedIf("variation_ids", variationNamesChange),
			validateFeatureFlagParameters,
		),
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
				Set:         hashFeatureFlagParameter,
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
			"variations": {
				Type:     schema.TypeSet,
				Optional: true,
				Set:      hashVariation,
				Description: "The variations of the feature flag. Each variation overrides parameter values " +
					"of the feature flag.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							Description: "The name of the variation. Must be unique within the feature flag and contain " +
								"3 to 255 characters without surrounding whitespace. Changing the name deletes the " +
								"variation and creates a new one with a new ID.",
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
										Description: "The name of a parameter on the feature flag. Each name can appear once.",
									},
									"type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice(parameterTypes, false),
										Description:  "The type of the parameter value. One of " + strings.Join(parameterTypes, ", ") + ".",
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
										Description: "The override value as a string. For object and array values, provide JSON; " +
											"for boolean and number values, provide their literal text.",
									},
								},
							},
						},
					},
				},
			},
			"variation_ids": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: "A map of variation names to variation IDs. Use it to reference a variation by ID, " +
					"for example in an experiment allocation, because the `variations` blocks do not include IDs. " +
					"An ID stays the same when you change the description or overrides of a variation. " +
					"A variation gets a new ID when you rename it, remove and add it again, or when the feature flag is replaced.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(statusValues, false),
				Description: fmt.Sprintf(
					"The lifecycle status of the feature flag. One of %s. Defaults to `%s` when omitted on creation.",
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

// Hash typed values so equivalent JSON, number and boolean strings keep the same set identity.
func hashFeatureFlagParameter(v interface{}) int {
	parameter := v.(map[string]interface{})
	name, _ := parameter["name"].(string)
	paramType, _ := parameter["type"].(string)
	value, _ := parameter["value"].(string)
	description, _ := parameter["description"].(string)

	if typed, err := stringToTyped(paramType, value); err == nil {
		if canonical, err := typedToString(typed); err == nil {
			value = canonical
		}
	}

	// Description must participate so editing it still produces a change.
	return schema.HashString(fmt.Sprintf("%q:%q:%q:%q", name, paramType, value, description))
}

// validateFeatureFlagParameters reports invalid parameter values at plan time instead of mid-apply.
func validateFeatureFlagParameters(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	config := diff.GetRawConfig()
	if !config.IsKnown() || config.IsNull() {
		return nil
	}

	return validateParameterValues(config.GetAttr("parameters"))
}

func validateParameterValues(parameters cty.Value) error {
	if !parameters.IsKnown() || parameters.IsNull() {
		return nil
	}

	var result *multierror.Error
	for it := parameters.ElementIterator(); it.Next(); {
		_, parameter := it.Element()
		if !parameter.IsKnown() || parameter.IsNull() {
			continue
		}

		paramType := parameter.GetAttr("type")
		rawValue := parameter.GetAttr("value")
		// Values from unknown expressions are checked again at apply time.
		if !paramType.IsKnown() || paramType.IsNull() || !rawValue.IsKnown() || rawValue.IsNull() {
			continue
		}

		if _, err := stringToTyped(paramType.AsString(), rawValue.AsString()); err != nil {
			name := parameter.GetAttr("name")
			if name.IsKnown() && !name.IsNull() {
				err = fmt.Errorf("parameter %q: %w", name.AsString(), err)
			}
			result = multierror.Append(result, err)
		}
	}

	return result.ErrorOrNil()
}

// hashVariation uses the override hashes, so equivalent override text does not cause a diff.
func hashVariation(v interface{}) int {
	variation := v.(map[string]interface{})
	name, _ := variation["name"].(string)
	description, _ := variation["description"].(string)

	var overrides []interface{}
	switch raw := variation["overrides"].(type) {
	case *schema.Set:
		overrides = raw.List()
	case []interface{}:
		overrides = raw
	}

	overrideHashes := make([]int, 0, len(overrides))
	for _, item := range overrides {
		overrideHashes = append(overrideHashes, hashOverride(item))
	}
	sort.Ints(overrideHashes)

	return schema.HashString(fmt.Sprintf("%q:%q:%v", name, description, overrideHashes))
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

// validateUniqueVariationAndOverrideNames rejects two variations with the same name and two overrides
// for the same parameter, because the API keys variations and overrides by name.
func validateUniqueVariationAndOverrideNames(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	seenVariations := make(map[string]bool)
	for _, item := range diff.Get("variations").(*schema.Set).List() {
		variation := item.(map[string]interface{})
		name := variation["name"].(string)
		if name == "" {
			continue
		}
		if seenVariations[name] {
			return fmt.Errorf("variation %q is configured more than once", name)
		}
		seenVariations[name] = true

		seenOverrides := make(map[string]bool)
		for _, override := range overrideNames(variation["overrides"].(*schema.Set)) {
			if seenOverrides[override] {
				return fmt.Errorf("override parameter %q is configured more than once in variation %q", override, name)
			}
			seenOverrides[override] = true
		}
	}

	return nil
}

// minVariationsToActivate is the API minimum for a status change to active.
const minVariationsToActivate = 2

// validateVariationCountOnActivation checks the variation count only when the status
// changes to active, because the API checks the count only at that change.
func validateVariationCountOnActivation(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	if !diff.HasChange("status") || !diff.NewValueKnown("status") || !diff.NewValueKnown("variations") {
		return nil
	}

	if diff.Get("status").(string) != string(management.FeatureFlagStatusEnumActive) {
		return nil
	}

	if count := diff.Get("variations").(*schema.Set).Len(); count < minVariationsToActivate {
		return fmt.Errorf("feature flag must have at least %d variations to set status %q, but %d are configured",
			minVariationsToActivate, management.FeatureFlagStatusEnumActive, count)
	}

	return nil
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

	// The ID is set before the next steps, so a failed step taints the
	// feature flag and the next plan replaces it.
	data.SetId(created.GetID())

	if err := createVariations(ctx, api, data.Id(), data.Get("variations").(*schema.Set).List()); err != nil {
		return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
	}

	// The create request does not accept a status, and activation needs the variations to exist.
	if status := expandFeatureFlagStatusOnCreate(data); status != nil {
		if err := updateFeatureFlagStatus(ctx, api, data.Id(), *status); err != nil {
			return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
		}
	}

	return readFeatureFlag(ctx, data, meta)
}

func readFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	flag, err := api.Experimentation.FeatureFlags.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_experiment_feature_flag", data, err)
	}

	variations, err := api.Experimentation.FeatureFlags.Variations.List(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_experiment_feature_flag", data, err)
	}

	return diag.FromErr(flattenFeatureFlag(data, flag, variations.GetVariations()))
}

// updateFeatureFlag stops at the first failed step. The steps run in this order:
// flag fields first, so the overrides can use new parameters, then the
// variations, then the status, so activation can use the new variations.
func updateFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	if data.HasChanges("name", "description", "parameters") {
		request, err := expandFeatureFlagUpdate(data)
		if err != nil {
			return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
		}

		if _, err := api.Experimentation.FeatureFlags.Update(ctx, data.Id(), request); err != nil {
			if err := internalError.HandleAPIError(data, err); err != nil {
				return append(diag.Errorf("failed to update feature flag %q: %v", data.Id(), err), readFeatureFlag(ctx, data, meta)...)
			}
			return nil
		}
	}

	if data.HasChange("variations") {
		if err := updateVariations(ctx, api, data); err != nil {
			return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
		}
	}

	if data.HasChange("status") {
		status := management.FeatureFlagStatusEnum(data.Get("status").(string))
		if err := updateFeatureFlagStatus(ctx, api, data.Id(), status); err != nil {
			return append(diag.FromErr(err), readFeatureFlag(ctx, data, meta)...)
		}
	}

	return readFeatureFlag(ctx, data, meta)
}

// The API deletes the variations of a feature flag with the feature flag.
func deleteFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	err := api.Experimentation.FeatureFlags.Delete(ctx, data.Id())
	return diag.FromErr(internalError.HandleAPIError(data, err))
}

func updateFeatureFlagStatus(ctx context.Context, api *managementv3.Management, featureFlagID string, status management.FeatureFlagStatusEnum) error {
	request := &management.UpdateFeatureFlagRequestContent{}
	request.SetStatus(&status)

	if _, err := api.Experimentation.FeatureFlags.Update(ctx, featureFlagID, request); err != nil {
		return fmt.Errorf("failed to set status %q on feature flag %q: %w", status, featureFlagID, err)
	}

	return nil
}

func createVariations(ctx context.Context, api *managementv3.Management, featureFlagID string, variations []interface{}) error {
	sortVariationsByName(variations)

	for _, item := range variations {
		variation := item.(map[string]interface{})
		name := variation["name"].(string)

		request, err := expandVariation(variation)
		if err != nil {
			return fmt.Errorf("variation %q: %w", name, err)
		}

		if _, err := api.Experimentation.FeatureFlags.Variations.Create(ctx, featureFlagID, request); err != nil {
			return fmt.Errorf("failed to create variation %q on feature flag %q: %w", name, featureFlagID, err)
		}
	}

	return nil
}

// updateVariations matches variations by name. It deletes the removed names,
// updates the changed ones and then creates the new names, so a renamed
// variation is deleted before its new name is created.
func updateVariations(ctx context.Context, api *managementv3.Management, data *schema.ResourceData) error {
	featureFlagID := data.Id()
	oldValue, newValue := data.GetChange("variations")
	oldVariations := variationsByName(oldValue.(*schema.Set))
	newVariations := variationsByName(newValue.(*schema.Set))

	oldIDs, _ := data.GetChange("variation_ids")
	variationIDs := oldIDs.(map[string]interface{})

	variationID := func(name string) (string, error) {
		id, _ := variationIDs[name].(string)
		if id == "" {
			return "", fmt.Errorf("variation %q on feature flag %q has no ID in state; refresh and try again", name, featureFlagID)
		}
		return id, nil
	}

	// Delete removed variations.
	for _, name := range sortedVariationNames(oldVariations) {
		if _, ok := newVariations[name]; ok {
			continue
		}

		id, err := variationID(name)
		if err != nil {
			return err
		}

		if err := api.Experimentation.FeatureFlags.Variations.Delete(ctx, featureFlagID, id); err != nil && !internalError.IsStatusNotFound(err) {
			return fmt.Errorf("failed to delete variation %q from feature flag %q: %w", name, featureFlagID, err)
		}
	}

	var created []interface{}
	for _, name := range sortedVariationNames(newVariations) {
		oldVariation, ok := oldVariations[name]
		if !ok {
			created = append(created, newVariations[name])
			continue
		}

		// Update param changes of variations.
		request, changed, err := expandVariationUpdate(oldVariation, newVariations[name])
		if err != nil {
			return fmt.Errorf("variation %q: %w", name, err)
		}
		if !changed {
			continue
		}

		id, err := variationID(name)
		if err != nil {
			return err
		}

		if _, err := api.Experimentation.FeatureFlags.Variations.Update(ctx, featureFlagID, id, request); err != nil {
			return fmt.Errorf("failed to update variation %q on feature flag %q: %w", name, featureFlagID, err)
		}
	}

	// Create new variations.
	return createVariations(ctx, api, featureFlagID, created)
}

func variationsByName(variations *schema.Set) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{}, variations.Len())
	for _, item := range variations.List() {
		variation := item.(map[string]interface{})
		result[variation["name"].(string)] = variation
	}
	return result
}

// variationNamesChange reports whether a variation is added, removed or renamed, because
// only these changes add or remove a variation ID; an update keeps the ID.
// It does not use diff.HasChange("variations"): for a set, HasChange compares the element
// text, so "true" and "True" count as a change even when the hashes are equal.
func variationNamesChange(_ context.Context, diff *schema.ResourceDiff, _ interface{}) bool {
	oldValue, newValue := diff.GetChange("variations")

	return !slices.Equal(
		sortedVariationNames(variationsByName(oldValue.(*schema.Set))),
		sortedVariationNames(variationsByName(newValue.(*schema.Set))),
	)
}

func sortedVariationNames(variations map[string]map[string]interface{}) []string {
	keys := make([]string, 0, len(variations))
	for key := range variations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// sortVariationsByName gives a fixed request order, so the HTTP recordings stay stable.
func sortVariationsByName(variations []interface{}) {
	sort.Slice(variations, func(i, j int) bool {
		return variations[i].(map[string]interface{})["name"].(string) < variations[j].(map[string]interface{})["name"].(string)
	})
}
