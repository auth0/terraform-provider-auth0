package riskassessment_test

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccWithZero = `resource "auth0_risk_assessments_new_device" "my_risk_assessments_new_device" { remember_for = 0}`
const testAccWithGreaterThanZero = `resource "auth0_risk_assessments_new_device" "my_risk_assessments_new_device" { remember_for = 20}`

func TestAccRiskAssessmentNewDevice(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccWithZero,
				ExpectError: regexp.MustCompile("expected remember_for to be at least"),
			},
			{
				Config: testAccWithGreaterThanZero,
				Check:  resource.TestCheckResourceAttr("auth0_risk_assessments_new_device.my_risk_assessments_new_device", "remember_for", strconv.Itoa(20)),
			},
		},
	})
}

const testAccRiskAssessmentNewDeviceInsufficient = `
resource "auth0_risk_assessments_new_device" "my_new_device" {
  remember_for = 30
}`

// TestAccRiskAssessmentNewDeviceInsufficientEntitlement asserts that when the
// tenant lacks the Adaptive MFA entitlement, the 403 response is non-fatal:
// apply succeeds with a warning rather than returning an error.
func TestAccRiskAssessmentNewDeviceInsufficientEntitlement(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccRiskAssessmentNewDeviceInsufficient,
			},
		},
	})
}
