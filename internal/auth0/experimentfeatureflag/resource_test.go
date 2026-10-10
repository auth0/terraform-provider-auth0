package experimentfeatureflag_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testFeatureFlagCreate = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "created by acceptance test"
	status      = "draft"

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
}
`

const testFeatureFlagUpdateAndArchive = `
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
}
`

func TestAccExperimentFeatureFlag(t *testing.T) {
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
					resource.TestCheckResourceAttrSet("auth0_experiment_feature_flag.my_flag", "created_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testFeatureFlagUpdateAndArchive, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "description", "updated by acceptance test"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "status", "archived"),
					resource.TestCheckResourceAttr("auth0_experiment_feature_flag.my_flag", "parameters.#", "2"),
				),
			},
		},
	})
}

const testFeatureFlagCreateActiveStatus = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name   = "tf-flag-{{.testName}}"
	status = "active"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
	}
}
`

func TestAccExperimentFeatureFlagCreateRejectsActiveStatus(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config:      acctest.ParseTestName(testFeatureFlagCreateActiveStatus, t.Name()),
				ExpectError: regexp.MustCompile(`(?s)cannot set status "active".*omit it or use "draft"`),
			},
		},
	})
}
