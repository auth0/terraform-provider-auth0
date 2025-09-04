package eventstream_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccDataSourceEventStreamConfig = `
resource "auth0_event_stream" "test" {
	name = "Acceptance-Test-DataSource-EventStream-{{.testName}}"
	status = "enabled"
	
	subscriptions {
		event_type = "user.created"
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

data "auth0_event_stream" "test" {
	id = auth0_event_stream.test.id
}
`

const testAccDataSourceEventStreamsConfig = `
resource "auth0_event_stream" "test1" {
	name = "Acceptance-Test-DataSource-EventStream-1-{{.testName}}"
	status = "enabled"
	
	subscriptions {
		event_type = "user.created"
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

resource "auth0_event_stream" "test2" {
	name = "Acceptance-Test-DataSource-EventStream-2-{{.testName}}"
	status = "disabled"
	
	subscriptions {
		event_type = "user.login"
	}

	destination {
		type = "eventbridge"
		aws_account_id = "123456789012"
		aws_region = "us-east-1"
	}
}

data "auth0_event_streams" "test" {
	depends_on = [auth0_event_stream.test1, auth0_event_stream.test2]
}
`

func TestAccDataSourceEventStream(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceEventStreamConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "id"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "name", fmt.Sprintf("Acceptance-Test-DataSource-EventStream-%s", t.Name())),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "status", "enabled"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "subscriptions.#", "1"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "destination.0.type", "webhook"),
					resource.TestCheckResourceAttr("data.auth0_event_stream.test", "destination.0.webhook_endpoint", "https://example.com/webhook/events"),
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "created_at"),
					resource.TestCheckResourceAttrSet("data.auth0_event_stream.test", "updated_at"),
				),
			},
		},
	})
}

func TestAccDataSourceEventStreams(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: acctest.ParseTestName(testAccDataSourceEventStreamsConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.auth0_event_streams.test", "event_streams.#"),
				),
			},
		},
	})
}
