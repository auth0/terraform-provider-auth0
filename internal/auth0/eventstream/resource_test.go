package eventstream_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testEventStreamCreateEventBridge = `
resource "auth0_event_stream" "my_event_stream" {
  name = "{{.testName}}-my-eventbridge"
  subscriptions = [
    "user.created",
    "user.updated"
  ]
  destination {
    type = "eventbridge"
    configuration = jsonencode(
		{
		  aws_account_id = "242849305777"
		  aws_region     = "us-east-1"
		}
	)
  }
}
`

const testEventStreamUpdateEventBridge = `
resource "auth0_event_stream" "my_event_stream" {
  name = "{{.testName}}-my-eventbridge-updated"
  subscriptions = [
    "user.updated"
  ]
  destination {
    type = "eventbridge"
    configuration = jsonencode(
		{
		  aws_account_id = "242849305777"
		  aws_region     = "us-east-1"
		}
	)
  }
}
`

const testEventStreamCreateWebhook = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name = "{{.testName}}-my-webhook"
  subscriptions = [
    "user.created",
    "user.updated"
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

const testEventStreamUpdateWebhook = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name = "{{.testName}}-my-webhook-updated"
  subscriptions = [
    "user.updated"
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

func TestAccEventStream(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testEventStreamCreateEventBridge, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("%s-my-eventbridge", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "2"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.0", "user.created"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.1", "user.updated"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.type", "eventbridge"),
				),
			},
			{
				Config: acctest.ParseTestName(testEventStreamUpdateEventBridge, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("%s-my-eventbridge-updated", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.0", "user.updated"),
				),
			},
			{
				Config: acctest.ParseTestName(testEventStreamCreateWebhook, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "name", fmt.Sprintf("%s-my-webhook", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.#", "2"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.0", "user.created"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.1", "user.updated"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "destination.0.type", "webhook"),
				),
			},
			{
				Config: acctest.ParseTestName(testEventStreamUpdateWebhook, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "name", fmt.Sprintf("%s-my-webhook-updated", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.0", "user.updated"),
				),
			},
		},
	})
}
