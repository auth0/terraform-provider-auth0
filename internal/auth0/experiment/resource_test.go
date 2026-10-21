package experiment_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testExperimentVariations = `
resource "auth0_experiment_feature_flag_variation" "control" {
	feature_flag_id = auth0_experiment_feature_flag.test.id
	name            = "tf-control-{{.testName}}"
}

resource "auth0_experiment_feature_flag_variation" "treatment" {
	feature_flag_id = auth0_experiment_feature_flag.test.id
	name            = "tf-treatment-{{.testName}}"

	# Both variations POST to the same URL; record and replay them in one order.
	depends_on = [auth0_experiment_feature_flag_variation.control]

	overrides {
		name  = "show_new_login"
		type  = "boolean"
		value = "true"
	}
}
`

const testExperimentDraftFlag = `
resource "auth0_experiment_feature_flag" "test" {
	name = "tf-flag-{{.testName}}"

	parameters {
		name  = "show_new_login"
		type  = "boolean"
		value = "false"
	}
}
` + testExperimentVariations

const testExperimentActiveFlag = `
resource "auth0_experiment_feature_flag" "test" {
	name   = "tf-flag-{{.testName}}"
	status = "active"

	parameters {
		name  = "show_new_login"
		type  = "boolean"
		value = "false"
	}
}
` + testExperimentVariations

const testExperimentPercentage = `
resource "auth0_experiment" "test" {
	name                = "tf-experiment-{{.testName}}"
	description         = "Experiment from acceptance test"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "authentication"
	allocation_strategy = "percentage"

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.control.variation_id
		is_control   = true
		weight       = 50
	}

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.treatment.variation_id
		is_control   = false
		weight       = 50
	}
}
`

const testExperimentPercentageUpdate = `
resource "auth0_experiment" "test" {
	name                = "tf-experiment-updated-{{.testName}}"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "authentication"
	allocation_strategy = "percentage"
	status              = "%s"

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.control.variation_id
		is_control   = true
		weight       = 70
	}

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.treatment.variation_id
		is_control   = false
		weight       = 30
	}
}
`

func TestAccExperiment(t *testing.T) {
	const address = "auth0_experiment.test"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+testExperimentPercentage, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(address, "id"),
					resource.TestCheckResourceAttr(address, "name", fmt.Sprintf("tf-experiment-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "description", "Experiment from acceptance test"),
					resource.TestCheckResourceAttr(address, "authentication_flow", "authentication"),
					resource.TestCheckResourceAttr(address, "allocation_strategy", "percentage"),
					resource.TestCheckResourceAttr(address, "status", "draft"),
					resource.TestCheckResourceAttr(address, "allocations.#", "2"),
					resource.TestCheckResourceAttr(address, "allocations.0.is_control", "true"),
					resource.TestCheckResourceAttr(address, "allocations.0.weight", "50"),
					resource.TestCheckResourceAttr(address, "allocations.0.variation_name", fmt.Sprintf("tf-control-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "allocations.1.is_fallback", "false"),
					resource.TestCheckResourceAttr(address, "feature_flag_name", fmt.Sprintf("tf-flag-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "feature_flag_snapshot", ""),
					resource.TestCheckResourceAttr(address, "started_at", ""),
					resource.TestCheckResourceAttrSet(address, "created_at"),
					resource.TestCheckResourceAttrSet(address, "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+testExperimentPercentage, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+fmt.Sprintf(testExperimentPercentageUpdate, "draft"), t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "name", fmt.Sprintf("tf-experiment-updated-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "description", ""),
					resource.TestCheckResourceAttr(address, "allocations.0.weight", "70"),
					resource.TestCheckResourceAttr(address, "allocations.1.weight", "30"),
				),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentPercentageUpdate, "draft"), t.Name()),
				Check:  resource.TestCheckResourceAttr("auth0_experiment_feature_flag.test", "status", "active"),
			},
			{
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentPercentageUpdate, "active"), t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "status", "active"),
					resource.TestCheckResourceAttrSet(address, "started_at"),
					resource.TestCheckResourceAttrSet(address, "feature_flag_snapshot"),
					resource.TestCheckResourceAttrSet(address, "allocations.0.variation_snapshot"),
				),
			},
			{
				// An active experiment cannot be deleted, so pause it before the test destroys it.
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentPercentageUpdate, "paused"), t.Name()),
				Check:  resource.TestCheckResourceAttr(address, "status", "paused"),
			},
		},
	})
}

const testExperimentRamp = `
resource "auth0_experiment" "test" {
	name                = "tf-experiment-{{.testName}}"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "password_reset"
	allocation_strategy = "percentage"
	levels              = [%s]

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.control.variation_id
		is_control   = true
	}

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.treatment.variation_id
		is_control   = false
	}
}
`

const testExperimentRampRemoved = `
resource "auth0_experiment" "test" {
	name                = "tf-experiment-{{.testName}}"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "password_reset"
	allocation_strategy = "percentage"

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.control.variation_id
		is_control   = true
		weight       = 50
	}

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.treatment.variation_id
		is_control   = false
		weight       = 50
	}
}
`

func TestAccExperimentRampLevels(t *testing.T) {
	const address = "auth0_experiment.test"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+fmt.Sprintf(testExperimentRamp, "10, 50, 100"), t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "levels.#", "3"),
					resource.TestCheckResourceAttr(address, "levels.0", "10"),
					resource.TestCheckResourceAttr(address, "current_level", "10"),
				),
			},
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+fmt.Sprintf(testExperimentRamp, "25, 75, 100"), t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "levels.0", "25"),
					resource.TestCheckResourceAttr(address, "current_level", "25"),
				),
			},
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+testExperimentRampRemoved, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "levels.#", "0"),
					resource.TestCheckResourceAttr(address, "allocations.0.weight", "50"),
				),
			},
		},
	})
}

const testExperimentSegment = `
resource "auth0_experiment_segment" "test" {
	name = "tf-segment-{{.testName}}"

	rules {
		match {
			country {
				contains = ["US"]
			}
		}
	}
}

resource "auth0_experiment" "test" {
	name                = "tf-experiment-{{.testName}}"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "mfa_challenge"
	allocation_strategy = "segment"
	status              = "%s"

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.control.variation_id
		is_control   = true
		segment_id   = auth0_experiment_segment.test.id
		priority     = 1
	}

	allocations {
		variation_id = auth0_experiment_feature_flag_variation.treatment.variation_id
		is_control   = false
		is_fallback  = true
	}
}
`

func TestAccExperimentSegmentAllocations(t *testing.T) {
	const address = "auth0_experiment.test"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+fmt.Sprintf(testExperimentSegment, "draft"), t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "allocation_strategy", "segment"),
					resource.TestCheckResourceAttrPair(address, "allocations.0.segment_id", "auth0_experiment_segment.test", "id"),
					resource.TestCheckResourceAttr(address, "allocations.0.segment_name", fmt.Sprintf("tf-segment-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "allocations.0.priority", "1"),
					resource.TestCheckResourceAttr(address, "allocations.1.is_fallback", "true"),
					resource.TestCheckResourceAttr(address, "allocations.1.segment_id", ""),
					resource.TestCheckResourceAttr(address, "allocations.0.segment_snapshot", ""),
				),
			},
			{
				Config: acctest.ParseTestName(testExperimentDraftFlag+fmt.Sprintf(testExperimentSegment, "draft"), t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentSegment, "draft"), t.Name()),
				Check:  resource.TestCheckResourceAttr("auth0_experiment_feature_flag.test", "status", "active"),
			},
			{
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentSegment, "active"), t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "status", "active"),
					resource.TestCheckResourceAttrSet(address, "allocations.0.segment_snapshot"),
					resource.TestCheckResourceAttr(address, "allocations.1.segment_snapshot", ""),
				),
			},
			{
				// An active experiment cannot be deleted, so pause it before the test destroys it.
				Config: acctest.ParseTestName(testExperimentActiveFlag+fmt.Sprintf(testExperimentSegment, "paused"), t.Name()),
				Check:  resource.TestCheckResourceAttr(address, "status", "paused"),
			},
		},
	})
}

const testExperimentCreateActiveStatus = testExperimentDraftFlag + `
resource "auth0_experiment" "test" {
	name                = "tf-experiment-{{.testName}}"
	feature_flag_id     = auth0_experiment_feature_flag.test.id
	authentication_flow = "authentication"
	allocation_strategy = "percentage"
	status              = "active"
}
`

func TestAccExperimentCreateRejectsActiveStatus(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testExperimentCreateActiveStatus, t.Name()),
				ExpectError: regexp.MustCompile(`(?s)cannot set status "active".*omit it or use "draft"`),
			},
		},
	})
}
