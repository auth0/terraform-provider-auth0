package experiment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceExperimentByID = testExperimentDraftFlag + testExperimentPercentage + `
data "auth0_experiment" "test" {
	id = auth0_experiment.test.id
}
`

func TestAccDataSourceExperiment(t *testing.T) {
	const dataSource = "data.auth0_experiment.test"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceExperimentByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource, "id", "auth0_experiment.test", "id"),
					resource.TestCheckResourceAttrPair(dataSource, "name", "auth0_experiment.test", "name"),
					resource.TestCheckResourceAttrPair(dataSource, "description", "auth0_experiment.test", "description"),
					resource.TestCheckResourceAttrPair(dataSource, "feature_flag_id", "auth0_experiment.test", "feature_flag_id"),
					resource.TestCheckResourceAttr(dataSource, "allocation_strategy", "percentage"),
					resource.TestCheckResourceAttr(dataSource, "status", "draft"),
					resource.TestCheckResourceAttr(dataSource, "allocations.#", "2"),
					resource.TestCheckResourceAttr(dataSource, "allocations.0.weight", "50"),
					resource.TestCheckResourceAttrSet(dataSource, "created_at"),
				),
			},
		},
	})
}
