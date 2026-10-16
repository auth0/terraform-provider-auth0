package experimentfeatureflag_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testVariationFeatureFlag = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name = "tf-flag-{{.testName}}"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
	}

	parameters {
		name  = "variant_name"
		type  = "string"
		value = "control"
	}

	parameters {
		name  = "max_items"
		type  = "number"
		value = "3"
	}
}
`

const testVariationCreate = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-{{.testName}}"
	description     = "created by acceptance test"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "true"
	}

	overrides {
		name  = "variant_name"
		type  = "string"
		value = "treatment"
	}

	overrides {
		name  = "max_items"
		type  = "number"
		value = "5"
	}
}
`

const testVariationSameValueOtherText = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-{{.testName}}"
	description     = "created by acceptance test"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "True"
	}

	overrides {
		name  = "variant_name"
		type  = "string"
		value = "treatment"
	}

	overrides {
		name  = "max_items"
		type  = "number"
		value = "5.0"
	}
}
`

const testVariationUpdate = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-updated-{{.testName}}"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "true"
	}

	overrides {
		name  = "variant_name"
		type  = "string"
		value = "treatment-b"
	}

	overrides {
		name  = "max_items"
		type  = "number"
		value = "5"
	}
}
`

const testVariationRemoveOverride = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-updated-{{.testName}}"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "true"
	}

	overrides {
		name  = "variant_name"
		type  = "string"
		value = "treatment-b"
	}
}
`

const testVariationRenameOverride = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-updated-{{.testName}}"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "true"
	}

	overrides {
		name  = "max_items"
		type  = "number"
		value = "7"
	}
}
`

const testVariationControl = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-updated-{{.testName}}"
}
`

func TestAccExperimentFeatureFlagVariation(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testVariationCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("auth0_experiment_feature_flag_variation.my_variation", "feature_flag_id", "auth0_experiment_feature_flag.my_flag", "id"),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag_variation.my_variation", "variation_id"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "name", fmt.Sprintf("tf-variation-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "description", "created by acceptance test"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "overrides.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "show_feature",
						"type":  "boolean",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "variant_name",
						"type":  "string",
						"value": "treatment",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "max_items",
						"type":  "number",
						"value": "5",
					}),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag_variation.my_variation", "created_at"),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag_variation.my_variation", "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testVariationSameValueOtherText, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: acctest.ParseTestName(testVariationUpdate, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag_variation.my_variation", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "name", fmt.Sprintf("tf-variation-updated-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "description", ""),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "overrides.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "variant_name",
						"type":  "string",
						"value": "treatment-b",
					}),
				),
			},
			{
				Config: acctest.ParseTestName(testVariationRemoveOverride, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag_variation.my_variation", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "overrides.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "show_feature",
						"type":  "boolean",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "variant_name",
						"type":  "string",
						"value": "treatment-b",
					}),
				),
			},
			{
				Config: acctest.ParseTestName(testVariationRenameOverride, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag_variation.my_variation", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "overrides.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "show_feature",
						"type":  "boolean",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag_variation.my_variation", "overrides.*", map[string]string{
						"name":  "max_items",
						"type":  "number",
						"value": "7",
					}),
				),
			},
			{
				ResourceName:      "auth0_experiment_feature_flag_variation.my_variation",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testVariationControl, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag_variation.my_variation", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag_variation.my_variation", "overrides.#", "0"),
				),
			},
		},
	})
}

const testVariationDuplicateOverrideNames = testVariationFeatureFlag + `
resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-{{.testName}}"

	overrides {
		name  = "max_items"
		type  = "number"
		value = "5"
	}

	overrides {
		name  = "max_items"
		type  = "number"
		value = "6"
	}
}
`

func TestAccExperimentFeatureFlagVariationRejectsDuplicateOverrideNames(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testVariationDuplicateOverrideNames, t.Name()),
				ExpectError: regexp.MustCompile(`override parameter "max_items" is configured more than once`),
			},
		},
	})
}
