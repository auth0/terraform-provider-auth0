package eventstream_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testEventStreamCreateEventBridge = `
resource "auth0_event_stream" "my_event_stream" {
  name              = "{{.testName}}-my-eventbridge"
  destination_type  = "eventbridge"
  subscriptions     = ["user.created", "user.updated"]

  eventbridge_configuration {
    aws_account_id = "242849305777"
    aws_region     = "us-east-1"
  }
}
`

const testEventStreamUpdateEventBridge = `
resource "auth0_event_stream" "my_event_stream" {
  name              = "{{.testName}}-my-eventbridge-updated"
  destination_type  = "eventbridge"
  subscriptions     = ["user.updated"]

  eventbridge_configuration {
    aws_account_id = "242849305777"
    aws_region     = "us-east-1"
  }
}
`

const testEventStreamCreateWebhook = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name              = "{{.testName}}-my-webhook"
  destination_type  = "webhook"
  subscriptions     = ["user.created", "user.updated"]

  webhook_configuration {
    webhook_endpoint = "https://eof28wtn4v4506o.m.pipedream.net"
    webhook_authorization {
      method = "bearer"
      token  = "123456789"
    }
  }
}
`

const testEventStreamUpdateWebhook = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name              = "{{.testName}}-my-webhook-updated"
  destination_type  = "webhook"
  subscriptions     = ["user.updated"]

  webhook_configuration {
    webhook_endpoint = "https://webhook.site/updated-endpoint"

    webhook_authorization {
      method   = "basic"
      username = "admin"
      password = "securepass"
    }
  }
}
`

const testEventStreamCreateWebhookWithTokenWO = `
resource "auth0_event_stream" "my_event_stream_webhook" {
  name              = "{{.testName}}-my-webhook"
  destination_type  = "webhook"
  subscriptions     = ["user.created"]

  webhook_configuration {
    webhook_endpoint = "https://test.com"
	webhook_authorization {
	  method		   = "bearer"
	  token_wo         = "secret_token_value"
	  token_wo_version = 1
	}
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
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination_type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "eventbridge_configuration.0.aws_account_id", "242849305777"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "eventbridge_configuration.0.aws_region", "us-east-1"),

					// Subscription assertions.
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream", "subscriptions.*", "user.created"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream", "subscriptions.*", "user.updated"),
				),
			},
			// Update EventBridge name + subscriptions (non-ForceNew).
			{
				Config: acctest.ParseTestName(testEventStreamUpdateEventBridge, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("%s-my-eventbridge-updated", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination_type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "eventbridge_configuration.0.aws_account_id", "242849305777"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "eventbridge_configuration.0.aws_region", "us-east-1"),

					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream", "subscriptions.*", "user.updated"),
				),
			},
			{
				Config: acctest.ParseTestName(testEventStreamCreateWebhook, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "name", fmt.Sprintf("%s-my-webhook", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "destination_type", "webhook"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_endpoint", "https://eof28wtn4v4506o.m.pipedream.net"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.method", "bearer"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.token", "123456789"),

					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.#", "2"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.*", "user.created"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.*", "user.updated"),
				),
			},

			// Step 4: Update Webhook name + subscriptions + webhook_endpoint and authorization details.
			{
				Config: acctest.ParseTestName(testEventStreamUpdateWebhook, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "name", fmt.Sprintf("%s-my-webhook-updated", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "destination_type", "webhook"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.#", "1"),
					resource.TestCheckTypeSetElemAttr("auth0_event_stream.my_event_stream_webhook", "subscriptions.*", "user.updated"),

					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_endpoint", "https://webhook.site/updated-endpoint"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.method", "basic"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.username", "admin"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.password", "securepass"),
				),
			},
			{
				Config: acctest.ParseTestName(testEventStreamCreateWebhookWithTokenWO, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "name", fmt.Sprintf("%s-my-webhook", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "destination_type", "webhook"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_endpoint", "https://test.com"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.method", "bearer"),

					resource.TestCheckNoResourceAttr("auth0_event_stream.my_event_stream_webhook", "webhook_configuration.0.webhook_authorization.0.token_wo"), // token_wo is write-only
				),
			},
		},
	})
}
