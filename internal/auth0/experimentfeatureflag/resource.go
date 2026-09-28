package experimentfeatureflag

import (
	"context"
	"reflect"
	"strings"

	management "github.com/auth0/go-auth0/v3/management"
	managementv3 "github.com/auth0/go-auth0/v3/management/client"
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
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Create and manage an Experiment Center feature flag together with its variations. (EA only)",
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
			"variation": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The variations of the feature flag.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
							Description: "Unique name of the variation. Name is used to correlate variations. " +
								"Hence changing it recreates the variation with a new ID.",
						},
						"description": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Description of the variation. Min length 3 when present.",
						},
						"overrides": {
							Type:     schema.TypeMap,
							Required: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Description: "Parameter overrides for this variation, keyed by parameter name. " +
								"Each value is a string (JSON-encoded for `object`/`array` parameters) and must differ from " +
								"the flag's default. May override a subset of the parameters.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the variation.",
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
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(statusValues, false),
				Description:  "The lifecycle status of the feature flag. One of " + strings.Join(statusValues, ", "),
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

	// Variations must be created through their own sub-endpoint; a nested array on the flag is rejected.
	variations, err := expandVariations(data)
	if err != nil {
		return diag.FromErr(err)
	}
	for _, v := range variations {
		if _, err := api.Experimentation.FeatureFlags.Variations.Create(ctx, data.Id(), v.createRequest()); err != nil {
			return diag.FromErr(err)
		}
	}

	// Apply the desired status last, once variations exist (activation requires at least two).
	if desired := data.Get("status").(string); desired != "" && desired != string(created.GetStatus()) {
		if _, err := api.Experimentation.FeatureFlags.UpdateStatus(ctx, data.Id(), &management.UpdateFeatureFlagStatusRequestContent{
			Status: management.FeatureFlagStatusEnum(desired),
		}); err != nil {
			return diag.FromErr(err)
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

	variations, err := readTrackedVariations(ctx, api, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(flattenFeatureFlag(data, flag, variations))
}

// readTrackedVariations returns the flag's variations that Terraform tracks in state, in state order.
func readTrackedVariations(ctx context.Context, api *managementv3.Management, data *schema.ResourceData) ([]*management.Variation, error) {
	stateVariations := data.Get("variation").([]interface{})
	// Nothing tracked (e.g. right after a plain passthrough import): skip the List call.
	if len(stateVariations) == 0 {
		return nil, nil
	}

	// One List (its response carries every field a per-id Get would) instead of N Gets.
	response, err := api.Experimentation.FeatureFlags.Variations.List(ctx, data.Id())
	if err != nil {
		return nil, err
	}

	apiVariations := make(map[string]*management.Variation, len(response.Variations))
	for _, v := range response.Variations {
		apiVariations[v.GetName()] = v
	}

	// Retain state-known variations in state order. Ignore out-of-band additions and API ordering;
	// drop out-of-band deletions so the next plan recreates them.
	variations := make([]*management.Variation, 0, len(stateVariations))
	for _, raw := range stateVariations {
		stateVariationName, _ := raw.(map[string]interface{})["name"].(string)
		if v, ok := apiVariations[stateVariationName]; ok {
			variations = append(variations, v)
		}
	}

	return variations, nil
}

func updateFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandFeatureFlagUpdate(data)
	if err != nil {
		return diag.FromErr(err)
	}

	if request != nil {
		if _, err := api.Experimentation.FeatureFlags.Update(ctx, data.Id(), request); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	if data.HasChange("variation") {
		if err := reconcileVariations(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}
	}

	if data.HasChange("status") {
		if desired := data.Get("status").(string); desired != "" {
			if _, err := api.Experimentation.FeatureFlags.UpdateStatus(ctx, data.Id(), &management.UpdateFeatureFlagStatusRequestContent{
				Status: management.FeatureFlagStatusEnum(desired),
			}); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return readFeatureFlag(ctx, data, meta)
}

func deleteFeatureFlag(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	if err := api.Experimentation.FeatureFlags.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

// variationUpdate pairs a changed variation's server-assigned id with its desired config.
type variationUpdate struct {
	id  string
	cfg variationConfig
}

// variationDiff is the set of variation changes to apply.
type variationDiff struct {
	toAdd    []variationConfig
	toUpdate []variationUpdate
	toRemove []string
}

// reconcileVariations applies the variation diff to the flag so it matches configuration.
func reconcileVariations(ctx context.Context, api *managementv3.Management, data *schema.ResourceData) error {
	diff, err := diffVariations(data)
	if err != nil {
		return err
	}

	// Create and update before delete: a flag must always keep at least one variation.
	for _, cfg := range diff.toAdd {
		if _, err := api.Experimentation.FeatureFlags.Variations.Create(ctx, data.Id(), cfg.createRequest()); err != nil {
			return err
		}
	}

	for _, u := range diff.toUpdate {
		if _, err := api.Experimentation.FeatureFlags.Variations.Update(ctx, data.Id(), u.id, u.cfg.updateRequest()); err != nil {
			return err
		}
	}

	for _, id := range diff.toRemove {
		if err := api.Experimentation.FeatureFlags.Variations.Delete(ctx, data.Id(), id); err != nil {
			return err
		}
	}

	return nil
}

// diffVariations classifies the configured variations into adds, updates, and removes.
func diffVariations(data *schema.ResourceData) (variationDiff, error) {
	paramTypesByName := parameterTypesByName(data.Get("parameters").(*schema.Set))

	oldRaw, newRaw := data.GetChange("variation")

	// Match by name: the API enforces per-flag uniqueness (409 variation_exists), so names map 1:1.
	// This preserves server IDs across removals and reordering (experiments reference them); renames
	// become remove + add and receive a new ID.
	oldByName := make(map[string]map[string]interface{})
	for _, raw := range oldRaw.([]interface{}) {
		entry := raw.(map[string]interface{})
		oldByName[entry["name"].(string)] = entry
	}

	var diff variationDiff
	newNames := make(map[string]struct{})

	for _, raw := range newRaw.([]interface{}) {
		newEntry := raw.(map[string]interface{})
		newName := newEntry["name"].(string)
		newNames[newName] = struct{}{}

		cfg, err := expandVariation(newEntry, paramTypesByName)
		if err != nil {
			return variationDiff{}, err
		}

		// Unseen name: a new variation. Existing name: patch only if its content changed; the
		// PATCH target is the matched prior-state entry's computed id.
		old, exists := oldByName[newName]
		if !exists {
			diff.toAdd = append(diff.toAdd, cfg)
			continue
		}

		if variationEntryChanged(old, newEntry) {
			diff.toUpdate = append(diff.toUpdate, variationUpdate{id: old["id"].(string), cfg: cfg})
		}
	}

	// Names in prior state but no longer in config are removed.
	for name, old := range oldByName {
		if _, exist := newNames[name]; !exist {
			diff.toRemove = append(diff.toRemove, old["id"].(string))
		}
	}

	return diff, nil
}

// variationEntryChanged reports whether a matched variation's patchable fields changed.
func variationEntryChanged(old, updated map[string]interface{}) bool {
	// Name is the correlation key; a changed name denotes a different variation. Ignore computed
	// fields (id, created_at, updated_at); only description and overrides are patchable.
	if old["description"].(string) != updated["description"].(string) {
		return true
	}
	return !reflect.DeepEqual(old["overrides"], updated["overrides"])
}
