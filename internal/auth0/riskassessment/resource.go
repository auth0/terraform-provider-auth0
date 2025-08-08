package riskassessment

import (
	"context"

	"github.com/auth0/terraform-provider-auth0/internal/value"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewResource will return a new auth0_risk_assessments resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		ReadContext:   readRiskAssessmentSettings,
		CreateContext: createRiskAssessmentSettings,
		UpdateContext: updateRiskAssessmentSettings,
		DeleteContext: schema.NoopContext, // Singleton, no delete.
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether risk assessment is enabled or not.",
			},
		},
		Description: "Resource for managing general Risk Assessment settings.",
	}
}

func createRiskAssessmentSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	data.SetId(id.UniqueId())
	return updateRiskAssessmentSettings(ctx, data, meta)
}

func readRiskAssessmentSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	settings, err := api.RiskAssessment.ReadSettings(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := data.Set("enabled", settings.Enabled); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func updateRiskAssessmentSettings(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	rawConfig := data.GetRawConfig()

	setting := &management.RiskAssessmentSettings{
		Enabled: value.Bool(rawConfig.GetAttr("enabled")),
	}

	if err := api.RiskAssessment.UpdateSettings(ctx, setting); err != nil {
		return diag.FromErr(err)
	}

	return readRiskAssessmentSettings(ctx, data, meta)
}
