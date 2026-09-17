package riskassessment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccWithFalse = `resource "auth0_risk_assessments" "my_risk_assessments_settings" { enabled = false}`
const testAccWithTrue = `resource "auth0_risk_assessments" "my_risk_assessments_settings" { enabled = true}`

func TestAccRiskAssessment(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccWithFalse,
				Check:  resource.TestCheckResourceAttr("auth0_risk_assessments.my_risk_assessments_settings", "enabled", "false"),
			},
			{
				Config: testAccWithTrue,
				Check:  resource.TestCheckResourceAttr("auth0_risk_assessments.my_risk_assessments_settings", "enabled", "true"),
			},
		},
	})
}

const testAccRiskAssessmentInsufficient = `
resource "auth0_risk_assessments" "my_risk_assessments_settings" {
  enabled = true
}`

// TestAccRiskAssessmentInsufficientEntitlement asserts that when the tenant
// lacks the Adaptive MFA entitlement, the 403 response is non-fatal: apply
// succeeds with a warning and produces a non-empty follow-up plan (because
// the configuration could not be applied or read back).
func TestAccRiskAssessmentInsufficientEntitlement(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:             testAccRiskAssessmentInsufficient,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
