package experimentsegment_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testSegmentCreate = `
resource "auth0_experiment_segment" "test" {
	name        = "tf-segment-{{.testName}}"
	description = "Segment from acceptance test"

	rules {
		match {
			country {
				one_of = ["US", "CA", "US"]
			}
			domain {
				starts_with = ["www"]
			}
			device_type {
				contains = ["mobile"]
			}
			client_id {
				exists = false
			}
			anonymous_session_metadata {
				key    = "plan"
				one_of = ["gold"]
			}
		}
		not_match {
			region {
				ends_with = ["-west"]
			}
		}
	}

	rules {
		match {
			organization_id {
				exists = true
			}
		}
	}
}
`

const testSegmentUpdate = `
resource "auth0_experiment_segment" "test" {
	name = "tf-segment-updated-{{.testName}}"

	rules {
		match {
			country {
				one_of = ["US", "CA"]
			}
			anonymous_session_metadata {
				key      = "plan"
				contains = ["gold"]
			}
			anonymous_session_metadata {
				key    = "internal"
				exists = false
			}
		}
		not_match {
			browser {
				ends_with = ["bot"]
			}
		}
	}
}
`

const testSegmentArchive = `
resource "auth0_experiment_segment" "test" {
	name   = "tf-segment-updated-{{.testName}}"
	status = "archived"

	rules {
		match {
			country {
				one_of = ["US", "CA"]
			}
			anonymous_session_metadata {
				key      = "plan"
				contains = ["gold"]
			}
			anonymous_session_metadata {
				key    = "internal"
				exists = false
			}
		}
		not_match {
			browser {
				ends_with = ["bot"]
			}
		}
	}
}
`

func TestAccExperimentSegment(t *testing.T) {
	const address = "auth0_experiment_segment.test"

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testSegmentCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(address, "id"),
					resource.TestCheckResourceAttr(address, "name", fmt.Sprintf("tf-segment-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "description", "Segment from acceptance test"),
					resource.TestCheckResourceAttr(address, "type", "self"),
					resource.TestCheckResourceAttr(address, "status", "active"),
					resource.TestCheckResourceAttr(address, "rules.#", "2"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.country.0.one_of.#", "3"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.country.0.one_of.2", "US"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.client_id.0.exists", "false"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.anonymous_session_metadata.#", "1"),
					resource.TestCheckResourceAttr(address, "rules.0.not_match.0.region.0.ends_with.0", "-west"),
					resource.TestCheckResourceAttrSet(address, "created_at"),
					resource.TestCheckResourceAttrSet(address, "updated_at"),
				),
			},
			{
				Config: acctest.ParseTestName(testSegmentCreate, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: acctest.ParseTestName(testSegmentUpdate, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "name", fmt.Sprintf("tf-segment-updated-%s", t.Name())),
					resource.TestCheckResourceAttr(address, "description", ""),
					resource.TestCheckResourceAttr(address, "status", "active"),
					resource.TestCheckResourceAttr(address, "rules.#", "1"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.country.0.one_of.#", "2"),
					resource.TestCheckResourceAttr(address, "rules.0.match.0.anonymous_session_metadata.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(address, "rules.0.match.0.anonymous_session_metadata.*", map[string]string{
						"key":        "plan",
						"contains.0": "gold",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(address, "rules.0.match.0.anonymous_session_metadata.*", map[string]string{
						"key":    "internal",
						"exists": "false",
					}),
				),
			},
			{
				ResourceName:      address,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ParseTestName(testSegmentArchive, t.Name()),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(address, "status", "archived"),
					resource.TestCheckResourceAttr(address, "rules.#", "1"),
				),
			},
		},
	})
}
