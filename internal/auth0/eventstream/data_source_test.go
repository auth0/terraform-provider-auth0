package eventstream_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccGivenAnEventStreamEventBridge = `
resource "auth0_event_stream" "my_event_stream" {
  name = "{{.testName}}-my-event-bridge"
  subscriptions = [
    "user.created"
  ]

  destination {
    type = "eventbridge"
    configuration = jsonencode({
      aws_account_id = "242849305777"
      aws_region     = "us-east-1"
    })
  }
}
`

const testAccGivenAnEventStreamWebhook = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name = "{{.testName}}-my-webhook"
  subscriptions = [
    "user.created"
  ]
  destination {
	type = "webhook"
	configuration = jsonencode(
		{
			webhook_endpoint = "https://eof28wtn4v4506o.m.pipedream.net"
			webhook_authorization = {
				method = "bearer"
				token = "123456789"
			}
		}
	)
  }
}
`

const testAccDataSourceEventStreamByIDEventBridge = testAccGivenAnEventStreamEventBridge + `
data "auth0_event_stream" "test" {
  id = auth0_event_stream.my_event_stream.id
}
`

const testAccDataSourceEventStreamByIDWebhook = testAccGivenAnEventStreamWebhook + `
data "auth0_event_stream" "test" {
  id = auth0_event_stream.my_event_stream_webhook.id
}
`

const testAccDataSourceEventStreamNonExistentID = `
data "auth0_event_stream" "test" {
  id = "est_invalid_id_1234567890"
}
`

func TestAccDataSourceEventStream(t *testing.T) {
	testName := strings.ToLower(t.Name())

	acctest.Test(t, resource.TestCase{
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			{
				Config:      `data "auth0_event_stream" "test" { }`,
				ExpectError: regexp.MustCompile(`The argument "id" is required, but no definition was found.`),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceEventStreamNonExistentID, testName),
				ExpectError: regexp.MustCompile(
					"Object didn't pass validation for format event-stream-id",
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceEventStreamByIDEventBridge, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "name", fmt.Sprintf("%s-my-event-bridge", testName)),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "subscriptions.0", "user.created"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "destination.0.type", "eventbridge"),
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "destination.0.configuration"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccDataSourceEventStreamByIDWebhook, testName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "name", fmt.Sprintf("%s-my-webhook", testName)),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "subscriptions.0", "user.created"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "destination.0.type", "webhook"),
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "destination.0.configuration"),
				),
			},
		},
	})
}
