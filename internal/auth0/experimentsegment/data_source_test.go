package experimentsegment_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceSegmentByID = `
resource "auth0_experiment_segment" "test" {
	name        = "tf-segment-{{.testName}}"
	description = "Segment from acceptance test"

	rules {
		match {
			country {
				one_of = ["US"]
			}

			anonymous_session_metadata {
				key    = "plan"
				one_of = ["gold"]
			}
		}
	}
}

data "auth0_experiment_segment" "test" {
	id = auth0_experiment_segment.test.id
}
`

func TestAccDataSourceExperimentSegment(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceSegmentByID, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.auth0_experiment_segment.test", "id", "auth0_experiment_segment.test", "id"),
					resource.TestCheckResourceAttrPair("data.auth0_experiment_segment.test", "name", "auth0_experiment_segment.test", "name"),
					resource.TestCheckResourceAttrPair("data.auth0_experiment_segment.test", "description", "auth0_experiment_segment.test", "description"),
					resource.TestCheckResourceAttr("data.auth0_experiment_segment.test", "status", "active"),
					resource.TestCheckResourceAttr("data.auth0_experiment_segment.test", "type", "self"),
					resource.TestCheckResourceAttr("data.auth0_experiment_segment.test", "rules.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_experiment_segment.test", "rules.0.match.0.country.0.one_of.0", "US"),
					resource.TestCheckResourceAttr("data.auth0_experiment_segment.test", "rules.0.match.0.anonymous_session_metadata.#", "1"),
					resource.TestCheckResourceAttrSet("data.auth0_experiment_segment.test", "created_at"),
				),
			},
		},
	})
}
