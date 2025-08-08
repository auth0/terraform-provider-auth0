package riskassessment_test

import (
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccWithZero = `resource "auth0_risk_assessments_new_device_settings" "my_risk_assessments_settings" { remember_for = 0}`
const testAccWithGreaterThanZero = `resource "auth0_risk_assessments_new_device_settings" "my_risk_assessments_settings" { remember_for = 20}`

func TestAccRiskAssessmentNewDevice(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      testAccWithZero,
				ExpectError: regexp.MustCompile("expected remember_for to be at least"),
			},
			{
				Config: testAccWithGreaterThanZero,
				Check:  resource.TestCheckResourceAttr("auth0_risk_assessments_new_device_settings.my_risk_assessments_settings", "remember_for", strconv.Itoa(20)),
			},
		},
	})
}
