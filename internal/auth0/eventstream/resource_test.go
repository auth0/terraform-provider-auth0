package eventstream_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccEventStreamWebhookConfig = `
resource "auth0_event_stream" "my_event_stream" {
	name = "Acceptance-Test-EventStream-webhook-{{.testName}}"
	status = "enabled"
	
	subscriptions {
		event_type = "user.created"
	}
	
	subscriptions {
		event_type = "user.updated"
	}

	destination {
		type = "webhook"
		webhook_endpoint = "https://example.com/webhook/events"
		webhook_authorization {
			method = "basic"
			username = "test_user"
			password = "test_password"
		}
	}
}
`

const testAccEventStreamWebhookConfigUpdate = `
resource "auth0_event_stream" "my_event_stream" {
	name = "Acceptance-Test-EventStream-webhook-updated-{{.testName}}"
	status = "disabled"
	
	subscriptions {
		event_type = "user.created"
	}

	destination {
		type = "webhook"
		webhook_endpoint = "https://example.com/webhook/events-updated"
		webhook_authorization {
			method = "bearer"
			token = "test_bearer_token"
		}
	}
}
`

const testAccEventStreamEventBridgeConfig = `
resource "auth0_event_stream" "my_event_stream" {
	name = "Acceptance-Test-EventStream-eventbridge-{{.testName}}"
	status = "enabled"
	
	subscriptions {
		event_type = "user.login"
	}

	destination {
		type = "eventbridge"
		aws_account_id = "123456789012"
		aws_region = "us-east-1"
	}
}
`

const testAccEventStreamActionConfig = `
resource "auth0_event_stream" "my_event_stream" {
	name = "Acceptance-Test-EventStream-action-{{.testName}}"
	status = "enabled"
	
	subscriptions {
		event_type = "user.logout"
	}

	destination {
		type = "action"
		action_id = "action_123456789"
	}
}
`

func TestAccEventStreamWebhook(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccEventStreamWebhookConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("Acceptance-Test-EventStream-webhook-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "status", "enabled"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "2"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.type", "webhook"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.webhook_endpoint", "https://example.com/webhook/events"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.webhook_authorization.0.method", "basic"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.webhook_authorization.0.username", "test_user"),
				),
			},
			{
				Config: acctest.ParseTestName(testAccEventStreamWebhookConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("Acceptance-Test-EventStream-webhook-updated-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "status", "disabled"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.type", "webhook"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.webhook_endpoint", "https://example.com/webhook/events-updated"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.webhook_authorization.0.method", "bearer"),
				),
			},
		},
	})
}

func TestAccEventStreamEventBridge(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccEventStreamEventBridgeConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("Acceptance-Test-EventStream-eventbridge-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "status", "enabled"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.aws_account_id", "123456789012"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.aws_region", "us-east-1"),
				),
			},
		},
	})
}

func TestAccEventStreamAction(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccEventStreamActionConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "name", fmt.Sprintf("Acceptance-Test-EventStream-action-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "status", "enabled"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.type", "action"),
					resource.TestCheckResourceAttr("auth0_event_stream.my_event_stream", "destination.0.action_id", "action_123456789"),
				),
			},
		},
	})
}
