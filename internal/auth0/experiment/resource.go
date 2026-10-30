package experiment

import (
	"context"
	"fmt"
	"strings"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var (
	authenticationFlowValues = []string{
		string(management.AuthenticationFlowEnumAuthentication),
		string(management.AuthenticationFlowEnumMfaEnrollment),
		string(management.AuthenticationFlowEnumMfaChallenge),
		string(management.AuthenticationFlowEnumPasswordReset),
		string(management.AuthenticationFlowEnumPasskeyEnrollment),
		string(management.AuthenticationFlowEnumAll),
	}
	allocationStrategyValues = []string{
		string(management.AllocationStrategyEnumPercentage),
		string(management.AllocationStrategyEnumSegment),
	}
	defaultConfigValues = []string{
		string(management.DefaultConfigEnumTenant),
		string(management.DefaultConfigEnumFlag),
	}
	statusValues = []string{
		string(management.ExperimentStatusEnumDraft),
		string(management.ExperimentStatusEnumActive),
		string(management.ExperimentStatusEnumPaused),
		string(management.ExperimentStatusEnumCompleted),
		string(management.ExperimentStatusEnumArchived),
	}
)

// NewResource returns the schema for the auth0_experiment resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createExperiment,
		ReadContext:   readExperiment,
		UpdateContext: updateExperiment,
		DeleteContext: deleteExperiment,
		Description:   "Create and manage an Experiment Center experiment. (EA only)",
		CustomizeDiff: customdiff.All(
			forceNewOnRemovedAllocations,
			forceNewOnRemovedLevels,
			forceNewOnRemovedDefaultConfig,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the experiment.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the experiment.",
			},
			"feature_flag_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the feature flag evaluated by this experiment. Changing this replaces the experiment.",
			},
			"authentication_flow": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(authenticationFlowValues, false),
				Description:  "The authentication flow targeted by the experiment. One of " + strings.Join(authenticationFlowValues, ", ") + ".",
			},
			"default_config": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(defaultConfigValues, false),
				Description: "The configuration source for an Auth0-managed feature flag. One of " + strings.Join(defaultConfigValues, ", ") +
					". Removing it replaces the experiment.",
			},
			"allocation_strategy": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringInSlice(allocationStrategyValues, false),
				Description:  "The traffic allocation strategy. One of " + strings.Join(allocationStrategyValues, ", ") + ". Changing this replaces the experiment.",
			},
			"allocations": {
				Type:        schema.TypeList,
				Optional:    true,
				MinItems:    1,
				Description: "The ordered traffic allocations. At least one is required when configured. Removing all allocations replaces the experiment.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"variation_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The ID of the feature flag variation assigned to this allocation.",
						},
						"is_control": {
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Whether this allocation is the control variation. At most one allocation can be the control.",
						},
						"weight": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The percentage weight of this allocation. Only for the percentage strategy. Omit when using levels.",
						},
						"segment_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The segment targeted by this allocation. Only for the segment strategy. Omit for the fallback allocation.",
						},
						"priority": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The evaluation priority of this allocation. Only for the segment strategy.",
						},
						"is_fallback": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Whether this allocation is the fallback. Only for the segment strategy.",
						},
						"variation_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the assigned variation.",
						},
						"segment_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the targeted segment.",
						},
						"variation_snapshot": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The snapshot of the variation.",
						},
						"segment_snapshot": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The snapshot of the segment.",
						},
					},
				},
			},
			"levels": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The ordered ramp levels. Only for the percentage strategy. Removing them replaces the experiment.",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"current_level": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The current ramp level, if levels are configured.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(statusValues, false),
				Description: "The lifecycle status of the experiment. One of " + strings.Join(statusValues, ", ") +
					". A non-draft status on creation is applied internally after the experiment is created as a draft. If the status update fails," +
					" Terraform marks the entire resource tainted and plans to replace it on the next apply. An active experiment cannot be deleted.",
			},
			"is_valid": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the experiment is ready to activate.",
			},
			"feature_flag_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the feature flag evaluated by this experiment.",
			},
			"feature_flag_snapshot": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The snapshot of the feature flag.",
			},
			"editable_fields": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Fields that the API reports as editable.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"started_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the experiment was first activated.",
			},
			"ended_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the experiment was completed.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the experiment was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the experiment was updated.",
			},
		},
	}
}

// PATCH rejects both an empty and a null allocations value, so removal requires replacement.
func forceNewOnRemovedAllocations(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	config := diff.GetRawConfig()
	if !config.IsKnown() || config.IsNull() || !config.GetAttr("allocations").IsKnown() {
		return nil
	}

	oldValue, newValue := diff.GetChange("allocations")
	oldAllocations, _ := oldValue.([]interface{})
	newAllocations, _ := newValue.([]interface{})
	if len(oldAllocations) > 0 && len(newAllocations) == 0 {
		return diff.ForceNew("allocations")
	}

	return nil
}

// PATCH rejects both an empty and a null levels value, so removal requires replacement.
func forceNewOnRemovedLevels(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	config := diff.GetRawConfig()
	if !config.IsKnown() || config.IsNull() || !config.GetAttr("levels").IsKnown() {
		return nil
	}

	oldValue, newValue := diff.GetChange("levels")
	oldLevels, _ := oldValue.([]interface{})
	newLevels, _ := newValue.([]interface{})
	if len(oldLevels) > 0 && len(newLevels) == 0 {
		return diff.ForceNew("levels")
	}

	return nil
}

// PATCH rejects a null default_config value, so removal requires replacement.
func forceNewOnRemovedDefaultConfig(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	config := diff.GetRawConfig()
	if !config.IsKnown() || config.IsNull() || !config.GetAttr("default_config").IsKnown() {
		return nil
	}

	oldValue, newValue := diff.GetChange("default_config")
	if oldValue.(string) != "" && newValue.(string) == "" {
		return diff.ForceNew("default_config")
	}

	return nil
}

func createExperiment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandExperiment(data)
	if err != nil {
		return diag.FromErr(err)
	}

	created, err := api.Experimentation.Experiments.Create(ctx, request)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(created.GetID())

	var diagnostics diag.Diagnostics
	requestedStatus := data.Get("status").(string)
	if requestedStatus != "" && requestedStatus != string(management.ExperimentStatusEnumDraft) {
		if err := patchExperimentStatus(ctx, meta, data.Id(), requestedStatus); err != nil {
			diagnostics = append(diagnostics, diag.FromErr(fmt.Errorf(
				"experiment %q was created, but setting status to %q failed: %w",
				data.Id(), requestedStatus, err,
			))...)
		}
	}

	return append(diagnostics, readExperiment(ctx, data, meta)...)
}

func patchExperimentStatus(ctx context.Context, meta interface{}, id, status string) error {
	api := meta.(*config.Config).GetAPIV3()
	transition := management.ExperimentTransitionStatusEnum(status)
	request := &management.UpdateExperimentRequestParameters{}
	request.SetStatus(&transition)

	_, err := api.Experimentation.Experiments.Update(ctx, id, request)
	return err
}

func readExperiment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	experiment, err := api.Experimentation.Experiments.Get(ctx, data.Id())
	if err != nil {
		// Keep the POST-created ID if an immediate read returns 404.
		if data.IsNewResource() && internalError.IsStatusNotFound(err) {
			return diag.FromErr(err)
		}
		return internalError.HandleReadAPIError("auth0_experiment", data, err)
	}

	return diag.FromErr(flattenExperiment(data, experiment))
}

func updateExperiment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandExperimentUpdate(data)
	if err != nil {
		return append(diag.FromErr(err), readExperiment(ctx, data, meta)...)
	}
	if request != nil {
		if _, err := api.Experimentation.Experiments.Update(ctx, data.Id(), request); err != nil {
			return append(diag.FromErr(internalError.HandleAPIError(data, err)), readExperiment(ctx, data, meta)...)
		}
	}

	// A separate status PATCH avoids the API's frozen-snapshot error for metadata edits.
	if data.HasChange("status") {
		if err := patchExperimentStatus(ctx, meta, data.Id(), data.Get("status").(string)); err != nil {
			return append(diag.FromErr(internalError.HandleAPIError(data, err)), readExperiment(ctx, data, meta)...)
		}
	}

	return readExperiment(ctx, data, meta)
}

func deleteExperiment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if data.Get("status").(string) == string(management.ExperimentStatusEnumActive) {
		return diag.Errorf("cannot delete active experiment %q: set status to \"paused\" or \"completed\" in a separate apply before destroying or replacing it", data.Id())
	}

	api := meta.(*config.Config).GetAPIV3()

	err := api.Experimentation.Experiments.Delete(ctx, data.Id())
	return diag.FromErr(internalError.HandleAPIError(data, err))
}
