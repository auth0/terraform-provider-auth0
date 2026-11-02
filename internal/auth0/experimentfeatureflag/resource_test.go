package experimentfeatureflag_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testFeatureFlagCreate = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "created by acceptance test"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
	}

	parameters {
		name  = "variant_name"
		type  = "string"
		value = "control"
	}

	variations {
		name = "control"
	}
}
`

const testFeatureFlagActivateWithNewVariation = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "created by acceptance test"
	status      = "active"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
	}

	parameters {
		name  = "variant_name"
		type  = "string"
		value = "control"
	}

	variations {
		name = "control"
	}

	variations {
		name        = "treatment"
		description = "created by acceptance test"

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
	}
}
`

const testFeatureFlagSameValueOtherText = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "created by acceptance test"
	status      = "active"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "False"
		description = "Whether to show the feature."
	}

	parameters {
		name  = "variant_name"
		type  = "string"
		value = "control"
	}

	variations {
		name = "control"
	}

	variations {
		name        = "treatment"
		description = "created by acceptance test"

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
	}
}
`

const testFeatureFlagUpdateVariationInPlace = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "updated by acceptance test"
	status      = "active"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
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

	variations {
		name = "control"
	}

	variations {
		name        = "treatment"
		description = "updated by acceptance test"

		overrides {
			name  = "show_feature"
			type  = "boolean"
			value = "true"
		}

		overrides {
			name  = "max_items"
			type  = "number"
			value = "5"
		}
	}
}
`

const testFeatureFlagRenameVariation = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "updated by acceptance test"
	status      = "active"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
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

	variations {
		name = "baseline"
	}

	variations {
		name        = "treatment"
		description = "updated by acceptance test"

		overrides {
			name  = "show_feature"
			type  = "boolean"
			value = "true"
		}

		overrides {
			name  = "max_items"
			type  = "number"
			value = "5"
		}
	}
}
`

const testFeatureFlagRemoveVariation = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "updated by acceptance test"
	status      = "active"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
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

	variations {
		name        = "treatment"
		description = "updated by acceptance test"

		overrides {
			name  = "show_feature"
			type  = "boolean"
			value = "true"
		}

		overrides {
			name  = "max_items"
			type  = "number"
			value = "5"
		}
	}
}
`

const testFeatureFlagArchive = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "updated by acceptance test"
	status      = "archived"

	parameters {
		name        = "show_feature"
		type        = "boolean"
		value       = "false"
		description = "Whether to show the feature."
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

	variations {
		name        = "treatment"
		description = "updated by acceptance test"

		overrides {
			name  = "show_feature"
			type  = "boolean"
			value = "true"
		}

		overrides {
			name  = "max_items"
			type  = "number"
			value = "5"
		}
	}
}
`

// captureAttr stores an attribute value of the feature flag, so a later step can compare it.
func captureAttr(key string, target *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith("auth0_experiment_feature_flag.my_flag", key, func(value string) error {
		*target = value
		return nil
	})
}

// attrEquals checks that an attribute still has the value captured in an earlier step.
func attrEquals(key string, expected *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith("auth0_experiment_feature_flag.my_flag", key, func(value string) error {
		if value != *expected {
			return fmt.Errorf("expected %s to stay %q, got %q", key, *expected, value)
		}
		return nil
	})
}

// attrDiffers checks that an attribute has a new value since an earlier step.
func attrDiffers(key string, previous *string) resource.TestCheckFunc {
	return resource.TestCheckResourceAttrWith("auth0_experiment_feature_flag.my_flag", key, func(value string) error {
		if value == "" || value == *previous {
			return fmt.Errorf("expected %s to get a new ID, got %q", key, value)
		}
		return nil
	})
}

func TestAccExperimentFeatureFlag(t *testing.T) {
	var controlID, treatmentID string

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testFeatureFlagCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "name", fmt.Sprintf("tf-flag-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "description", "created by acceptance test"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "draft"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "type", "self"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "parameters.#", "2"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*", map[string]string{
						"name":        "control",
						"overrides.#": "0",
					}),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variation_ids.%", "1"),
					captureAttr("variation_ids.control", &controlID),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag.my_flag", "created_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagActivateWithNewVariation, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag.my_flag", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "active"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*", map[string]string{
						"name":        "treatment",
						"description": "created by acceptance test",
						"overrides.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*.overrides.*", map[string]string{
						"name":  "show_feature",
						"type":  "boolean",
						"value": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*.overrides.*", map[string]string{
						"name":  "variant_name",
						"type":  "string",
						"value": "treatment",
					}),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variation_ids.%", "2"),
					attrEquals("variation_ids.control", &controlID),
					captureAttr("variation_ids.treatment", &treatmentID),
				),
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagSameValueOtherText, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagUpdateVariationInPlace, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag.my_flag", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "description", "updated by acceptance test"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "parameters.#", "3"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*", map[string]string{
						"name":        "treatment",
						"description": "updated by acceptance test",
						"overrides.#": "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*.overrides.*", map[string]string{
						"name":  "max_items",
						"type":  "number",
						"value": "5",
					}),
					attrEquals("variation_ids.control", &controlID),
					attrEquals("variation_ids.treatment", &treatmentID),
				),
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagRenameVariation, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag.my_flag", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("auth0_experiment_feature_flag.my_flag", "variations.*", map[string]string{
						"name": "baseline",
					}),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variation_ids.%", "2"),
					resource.TestCheckNoResourceAttr("auth0_experiment_feature_flag.my_flag", "variation_ids.control"),
					attrDiffers("variation_ids.baseline", &controlID),
					attrEquals("variation_ids.treatment", &treatmentID),
				),
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagRemoveVariation, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("auth0_experiment_feature_flag.my_flag", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "active"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "1"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variation_ids.%", "1"),
					attrEquals("variation_ids.treatment", &treatmentID),
				),
			},
			{
				ResourceName:      "auth0_experiment_feature_flag.my_flag",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagArchive, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "archived"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "1"),
				),
			},
		},
	})
}

const testFeatureFlagCreateActive = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name   = "tf-flag-{{.testName}}"
	status = "active"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
	}

	variations {
		name = "control"
	}

	variations {
		name = "treatment"

		overrides {
			name  = "show_feature"
			type  = "boolean"
			value = "true"
		}
	}
}
`

func TestAccExperimentFeatureFlagCreateActive(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testFeatureFlagCreateActive, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "active"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "variations.#", "2"),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag.my_flag", "variation_ids.control"),
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag.my_flag", "variation_ids.treatment"),
				),
			},
		},
	})
}

const testFeatureFlagActivateWithOneVariation = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name   = "tf-flag-{{.testName}}"
	status = "active"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
	}

	variations {
		name = "control"
	}
}
`

const testFeatureFlagDuplicateVariationNames = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name = "tf-flag-{{.testName}}"

	parameters {
		name  = "max_items"
		type  = "number"
		value = "3"
	}

	variations {
		name = "treatment"
	}

	variations {
		name = "treatment"

		overrides {
			name  = "max_items"
			type  = "number"
			value = "5"
		}
	}
}
`

const testFeatureFlagDuplicateOverrideNames = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name = "tf-flag-{{.testName}}"

	parameters {
		name  = "max_items"
		type  = "number"
		value = "3"
	}

	variations {
		name = "treatment"

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
}
`

func TestAccExperimentFeatureFlagRejectsInvalidVariations(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testFeatureFlagActivateWithOneVariation, t.Name()),
				ExpectError: regexp.MustCompile(`(?s)feature flag must have at least 2 variations to set status.*"active", but 1 are configured`),
			},
			{
				Config:      acctest.ParseTestName(testFeatureFlagDuplicateVariationNames, t.Name()),
				ExpectError: regexp.MustCompile(`variation "treatment" is configured more than once`),
			},
			{
				Config:      acctest.ParseTestName(testFeatureFlagDuplicateOverrideNames, t.Name()),
				ExpectError: regexp.MustCompile(`(?s)override parameter "max_items" is configured more than once.*in variation "treatment"`),
			},
		},
	})
}
