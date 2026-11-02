package experimentfeatureflag_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceFeatureFlagByID = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name        = "tf-flag-{{.testName}}"
	description = "created by acceptance test"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
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
	}
}

data "auth0_experiment_feature_flag" "test" {
	id = auth0_experiment_feature_flag.my_flag.id
}
`

func TestAccDataSourceExperimentFeatureFlag(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceFeatureFlagByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_experiment_feature_flag.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "name", fmt.Sprintf("tf-flag-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "description", "created by acceptance test"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "type", "self"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "status", "draft"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "parameters.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag.test", "variations.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_experiment_feature_flag.test", "variations.*", map[string]string{
						"name":        "treatment",
						"description": "created by acceptance test",
						"overrides.#": "1",
					}),
					resource.TestCheckResourceAttrPair(
						"data.auth0_experiment_feature_flag.test", "variation_ids.control",
						"auth0_experiment_feature_flag.my_flag", "variation_ids.control",
					),
					resource.TestCheckResourceAttrPair(
						"data.auth0_experiment_feature_flag.test", "variation_ids.treatment",
						"auth0_experiment_feature_flag.my_flag", "variation_ids.treatment",
					),
				),
			},
		},
	})
}
