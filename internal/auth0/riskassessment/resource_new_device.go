package riskassessment

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	apierr "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewDeviceSettingResource will return a new auth0_risk_assessments_new_device resource.
func NewDeviceSettingResource() *schema.Resource {
	return &schema.Resource{
		ReadContext:   readRiskAssessmentNewDeviceSettings,
		CreateContext: createRiskAssessmentNewDeviceSettings,
		UpdateContext: updateRiskAssessmentNewDeviceSettings,
		DeleteContext: schema.NoopContext,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"remember_for": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntAtLeast(1),
				Description:  "Length of time to remember devices for, in days",
			},
		},
		Description: "Resource for managing Risk Assessment settings for new devices.",
	}
}

func createRiskAssessmentNewDeviceSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateRiskAssessmentNewDeviceSettings(ctx, data, meta)
}

func readRiskAssessmentNewDeviceSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	settings, err := api.RiskAssessment.ReadNewDeviceSettings(ctx)
	if err != nil {
		if apierr.IsInsufficientEntitlement(err) {
			return diag.Diagnostics{apierr.EntitlementWarning("Risk Assessment (New Device)", "its current configuration could not be read")}
		}
		return diag.FromErr(err)
	}

	if err := data.Set("remember_for", settings.RememberFor); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateRiskAssessmentNewDeviceSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	rawConfig := data.GetRawConfig()

	setting := &management.RiskAssessmentSettingsNewDevice{
		RememberFor: value.Int(rawConfig.GetAttr("remember_for")),
	}

	if err := api.RiskAssessment.UpdateNewDeviceSettings(ctx, setting); err != nil {
		if apierr.IsInsufficientEntitlement(err) {
			return diag.Diagnostics{apierr.EntitlementWarning("Risk Assessment (New Device)", "the configuration was not applied")}
		}
		return diag.FromErr(fmt.Errorf("failed to update new device settings: %w", err))
	}

	return readRiskAssessmentNewDeviceSettings(ctx, data, meta)
}
