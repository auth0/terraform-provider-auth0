package experimentfeatureflag_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceVariationByID = `
resource "auth0_experiment_feature_flag" "my_flag" {
	name = "tf-flag-{{.testName}}"

	parameters {
		name  = "show_feature"
		type  = "boolean"
		value = "false"
	}
}

resource "auth0_experiment_feature_flag_variation" "my_variation" {
	feature_flag_id = auth0_experiment_feature_flag.my_flag.id
	name            = "tf-variation-{{.testName}}"
	description     = "created by acceptance test"

	overrides {
		name  = "show_feature"
		type  = "boolean"
		value = "true"
	}
}

data "auth0_experiment_feature_flag_variation" "test" {
	feature_flag_id = auth0_experiment_feature_flag_variation.my_variation.feature_flag_id
	variation_id    = auth0_experiment_feature_flag_variation.my_variation.variation_id
}
`

func TestAccDataSourceExperimentFeatureFlagVariation(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceVariationByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.auth0_experiment_feature_flag_variation.test", "id", "auth0_experiment_feature_flag_variation.my_variation", "id"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag_variation.test", "name", fmt.Sprintf("tf-variation-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag_variation.test", "description", "created by acceptance test"),
					resource.TestCheckResourceAttr("data.auth0_experiment_feature_flag_variation.test", "overrides.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("data.auth0_experiment_feature_flag_variation.test", "overrides.*", map[string]string{
						"name":  "show_feature",
						"type":  "boolean",
						"value": "true",
					}),
					resource.TestCheckResourceAttrSet("data.auth0_experiment_feature_flag_variation.test", "created_at"),
				),
			},
		},
	})
}
