package auth0

import (
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/random"
)

func init() {
	resource.AddTestSweepers("auth0_log_stream", &resource.Sweeper{
		Name: "auth0_log_stream",
		F: func(_ string) error {
			api, err := Auth0()
			if err != nil {
				return err
			}

			logStreams, err := api.LogStream.List()
			if err != nil {
				return err
			}

			var result *multierror.Error
			for _, logStream := range logStreams {
				log.Printf("[DEBUG] ➝ %s", logStream.GetName())

				if strings.Contains(logStream.GetName(), "Test") {
					result = multierror.Append(
						result,
						api.LogStream.Delete(logStream.GetID()),
					)

					log.Printf("[DEBUG] ✗ %v\n", logStream.GetName())
				}
			}

			return result.ErrorOrNil()
		},
	})
}

func TestAccLogStreamHTTP(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(testAccLogStreamHTTPConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-http-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "status", "paused"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: random.Template(testAccLogStreamHTTPConfigUpdateFormatToJSONARRAY, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-http-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json; charset=utf-8"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONARRAY"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: random.Template(testAccLogStreamHTTPConfigUpdateFormatToJSONOBJECT, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-http-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json; charset=utf-8"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONOBJECT"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: random.Template(testAccLogStreamHTTPConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-http-new-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: random.Template(testAccLogStreamHTTPConfigUpdateCustomHTTPHeaders, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-http-new-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.#", "2"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.0.header", "foo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.0.value", "bar"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.1.header", "bar"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.1.value", "foo"),
				),
			},
		},
	})
}

const testAccLogStreamHTTPConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-{{.random}}"
	type = "http"
	status = "paused"
	sink {
	  http_endpoint = "https://example.com/webhook/logs"
	  http_content_type = "application/json"
	  http_content_format = "JSONLINES"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	}
}
`

const testAccLogStreamHTTPConfigUpdateFormatToJSONARRAY = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-{{.random}}"
	type = "http"
	sink {
	  http_endpoint = "https://example.com/webhook/logs"
	  http_content_type = "application/json; charset=utf-8"
	  http_content_format = "JSONARRAY"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	}
}
`

const testAccLogStreamHTTPConfigUpdateFormatToJSONOBJECT = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-{{.random}}"
	type = "http"
	sink {
	  http_endpoint = "https://example.com/webhook/logs"
	  http_content_type = "application/json; charset=utf-8"
	  http_content_format = "JSONOBJECT"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	}
}
`

const testAccLogStreamHTTPConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-new-{{.random}}"
	type = "http"
	sink {
	  http_endpoint = "https://example.com/logs"
	  http_content_type = "application/json"
	  http_content_format = "JSONLINES"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	}
}
`

const testAccLogStreamHTTPConfigUpdateCustomHTTPHeaders = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-new-{{.random}}"
	type = "http"
	sink {
	  http_endpoint = "https://example.com/logs"
	  http_content_type = "application/json"
	  http_content_format = "JSONLINES"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	  http_custom_headers = [
        {
          header = "foo"
          value  = "bar"
        },
		{
          header = "bar"
          value  = "foo"
        }
      ]
	}
}
`

func TestAccLogStreamEventBridge(t *testing.T) {
	rand := random.String(6)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(logStreamAwsEventBridgeConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-aws-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_account_id", "999999999999"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_region", "us-west-2"),
				),
			},
			{
				Config: random.Template(logStreamAwsEventBridgeConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-aws-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_account_id", "899999999998"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_region", "us-west-1"),
				),
			},
			{
				Config: random.Template(logStreamAwsEventBridgeConfigUpdateName, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-aws-new-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_account_id", "899999999998"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_region", "us-west-1"),
				),
			},
		},
	})
}

const logStreamAwsEventBridgeConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-aws-{{.random}}"
	type = "eventbridge"
	sink {
	  aws_account_id = "999999999999"
	  aws_region = "us-west-2"
	}
}
`
const logStreamAwsEventBridgeConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-aws-{{.random}}"
	type = "eventbridge"
	sink {
	  aws_account_id = "899999999998"
	  aws_region = "us-west-1"
	}
}
`

const logStreamAwsEventBridgeConfigUpdateName = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-aws-new-{{.random}}"
	type = "eventbridge"
	sink {
	  aws_account_id = "899999999998"
	  aws_region = "us-west-1"
	}
}
`

// This test fails it subscription key is not valid, or EventGrid
// Resource Provider is not registered in the subscription.
func TestAccLogStreamEventGrid(t *testing.T) {
	t.Skip("this test requires an active subscription")

	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(logStreamAzureEventGridConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-azure-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventgrid"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_subscription_id", "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_region", "northeurope"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_resource_group", "azure-logs-rg"),
				),
			},
			{
				Config: random.Template(logStreamAzureEventGridConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-azure-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventgrid"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_subscription_id", "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_region", "northeurope"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_resource_group", "azure-logs-rg"),
				),
			},
		},
	})
}

const logStreamAzureEventGridConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-azure-{{.random}}"
	type = "eventgrid"
	sink {
  	  azure_subscription_id = "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"
	  azure_region = "northeurope"
	  azure_resource_group = "azure-logs-rg"
	}
}
`
const logStreamAzureEventGridConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-azure-{{.random}}"
	type = "eventgrid"
	sink {
  	  azure_subscription_id = "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"
	  azure_region = "westeurope"
	  azure_resource_group = "azure-logs-rg"
	}
}
`

func TestAccLogStreamDatadog(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(logStreamDatadogConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-datadog-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "datadog"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_api_key", "121233123455"),
				),
			},
			{
				Config: random.Template(logStreamDatadogConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-datadog-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "datadog"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_region", "eu"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_api_key", "121233123455"),
				),
			},
			{
				Config: random.Template(logStreamDatadogConfigRemoveAndCreate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-datadog-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "datadog"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_region", "eu"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_api_key", "1212331234556667"),
				),
			},
		},
	})
}

const logStreamDatadogConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.random}}"
	type = "datadog"
	sink {
	  datadog_region = "us"
	  datadog_api_key = "121233123455"
	}
}
`
const logStreamDatadogConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.random}}"
	type = "datadog"
	sink {
	  datadog_region = "eu"
	  datadog_api_key = "121233123455"
	}
}
`
const logStreamDatadogConfigRemoveAndCreate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.random}}"
	type = "datadog"
	sink {
	  datadog_region = "eu"
	  datadog_api_key = "1212331234556667"
	}
}
`

func TestAccLogStreamSplunk(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(logStreamSplunkConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-splunk-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "splunk"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_domain", "demo.splunk.com"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_token", "12a34ab5-c6d7-8901-23ef-456b7c89d0c1"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_port", "8088"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_secure", "true"),
				),
			},
			{
				Config: random.Template(logStreamSplunkConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-splunk-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "splunk"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_domain", "prod.splunk.com"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_token", "12a34ab5-c6d7-8901-23ef-456b7c89d0d1"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_port", "8088"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_secure", "true"),
				),
			},
		},
	})
}

const logStreamSplunkConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-splunk-{{.random}}"
	type = "splunk"
	sink {
	  splunk_domain = "demo.splunk.com"
	  splunk_token = "12a34ab5-c6d7-8901-23ef-456b7c89d0c1"
	  splunk_port = "8088"
	  splunk_secure = "true"
	}
}
`
const logStreamSplunkConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-splunk-{{.random}}"
	type = "splunk"
	sink {
	  splunk_domain = "prod.splunk.com"
	  splunk_token = "12a34ab5-c6d7-8901-23ef-456b7c89d0d1"
	  splunk_port = "8088"
	  splunk_secure = "true"
	}
}
`

func TestAccLogStreamSumo(t *testing.T) {
	rand := random.String(6)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: random.Template(logStreamSumoConfig, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-sumo-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "demo.sumo.com"),
				),
			},
			{
				Config: random.Template(logStreamSumoConfigUpdate, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-sumo-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "prod.sumo.com"),
				),
			},
			{
				Config: random.Template(logStreamSumoConfigUpdateWithFilters, rand),
				Check: resource.ComposeTestCheckFunc(
					random.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", "Acceptance-Test-LogStream-sumo-{{.random}}", rand),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "2"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.name", "auth.login.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.name", "auth.signup.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "prod.sumo.com"),
				),
			},
		},
	})
}

const logStreamSumoConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.random}}"
	type = "sumo"
	sink {
	  sumo_source_address = "demo.sumo.com"
	}
}
`
const logStreamSumoConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.random}}"
	type = "sumo"
	sink {
	  sumo_source_address = "prod.sumo.com"
	}
}
`
const logStreamSumoConfigUpdateWithFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.random}}"
	type = "sumo"
	filters = [
		{
			type = "category"
			name = "auth.login.fail"
		},
		{
			type = "category"
			name = "auth.signup.fail"
		}
	]
	sink {
	  sumo_source_address = "prod.sumo.com"
	}
}
`
