package logstream_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
	"github.com/auth0/terraform-provider-auth0/internal/template"
)

func TestAccLogStreamHTTP(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "status", "paused"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfigUpdateFormatToJSONARRAY, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json; charset=utf-8"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONARRAY"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfigUpdateFormatToJSONOBJECT, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/webhook/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json; charset=utf-8"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONOBJECT"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-new-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
				),
			},
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfigUpdateCustomHTTPHeaders, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-new-%s", t.Name())),
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
			{
				Config: template.ParseTestName(testAccLogStreamHTTPConfigEmptyCustomHTTPHeaders, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-http-new-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "http"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_endpoint", "https://example.com/logs"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_type", "application/json"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_content_format", "JSONLINES"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_authorization", "AKIAXXXXXXXXXXXXXXXX"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.http_custom_headers.#", "0"),
				),
			},
		},
	})
}

const testAccLogStreamHTTPConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-http-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-http-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-http-new-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-http-new-{{.testName}}"
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

const testAccLogStreamHTTPConfigEmptyCustomHTTPHeaders = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-http-new-{{.testName}}"
	type = "http"
	sink {
	  http_endpoint = "https://example.com/logs"
	  http_content_type = "application/json"
	  http_content_format = "JSONLINES"
	  http_authorization = "AKIAXXXXXXXXXXXXXXXX"
	  http_custom_headers = []
	}
}
`

func TestAccLogStreamEventBridge(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamAwsEventBridgeConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-aws-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_account_id", "999999999999"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_region", "us-west-2"),
				),
			},
			{
				Config: template.ParseTestName(logStreamAwsEventBridgeConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-aws-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventbridge"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_account_id", "899999999998"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.aws_region", "us-west-1"),
				),
			},
			{
				Config: template.ParseTestName(logStreamAwsEventBridgeConfigUpdateName, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-aws-new-%s", t.Name())),
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
	name = "Acceptance-Test-LogStream-aws-{{.testName}}"
	type = "eventbridge"
	sink {
	  aws_account_id = "999999999999"
	  aws_region = "us-west-2"
	}
}
`
const logStreamAwsEventBridgeConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-aws-{{.testName}}"
	type = "eventbridge"
	sink {
	  aws_account_id = "899999999998"
	  aws_region = "us-west-1"
	}
}
`

const logStreamAwsEventBridgeConfigUpdateName = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-aws-new-{{.testName}}"
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
	if os.Getenv("AUTH0_DOMAIN") != acctest.RecordingsDomain {
		t.Skip()
	}

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamAzureEventGridConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-azure-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventgrid"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_subscription_id", "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_region", "northeurope"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_resource_group", "azure-logs-rg"),
				),
			},
			{
				Config: template.ParseTestName(logStreamAzureEventGridConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-azure-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "eventgrid"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_subscription_id", "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_region", "westeurope"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.azure_resource_group", "azure-logs-rg"),
				),
			},
		},
	})
}

const logStreamAzureEventGridConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-azure-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-azure-{{.testName}}"
	type = "eventgrid"
	sink {
  	  azure_subscription_id = "b69a6835-57c7-4d53-b0d5-1c6ae580b6d5"
	  azure_region = "westeurope"
	  azure_resource_group = "azure-logs-rg"
	}
}
`

func TestAccLogStreamDataDogRegionValidation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProviderFactories: acctest.TestFactories(),
		Steps: []resource.TestStep{
			{
				Config:      fmt.Sprintf(logStreamDatadogInvalidConfig, "uS"),
				ExpectError: regexp.MustCompile(`expected sink.0.datadog_region to be one of \[us eu us3 us5\], got uS`),
			},
			{
				Config:      fmt.Sprintf(logStreamDatadogInvalidConfig, "us9"),
				ExpectError: regexp.MustCompile(`expected sink.0.datadog_region to be one of \[us eu us3 us5\], got us9`),
			},
		},
	})
}

const logStreamDatadogInvalidConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.testName}}"
	type = "datadog"
	sink {
	  datadog_region = "%s"
	  datadog_api_key = "121233123455"
	}
}
`

func TestAccLogStreamDatadog(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamDatadogConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-datadog-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "datadog"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_api_key", "121233123455"),
				),
			},
			{
				Config: template.ParseTestName(logStreamDatadogConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-datadog-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "datadog"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_region", "eu"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.datadog_api_key", "121233123455"),
				),
			},
			{
				Config: template.ParseTestName(logStreamDatadogConfigRemoveAndCreate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-datadog-%s", t.Name())),
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
	name = "Acceptance-Test-LogStream-datadog-{{.testName}}"
	type = "datadog"
	sink {
	  datadog_region = "us"
	  datadog_api_key = "121233123455"
	}
}
`
const logStreamDatadogConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.testName}}"
	type = "datadog"
	sink {
	  datadog_region = "eu"
	  datadog_api_key = "121233123455"
	}
}
`
const logStreamDatadogConfigRemoveAndCreate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-datadog-{{.testName}}"
	type = "datadog"
	sink {
	  datadog_region = "eu"
	  datadog_api_key = "1212331234556667"
	}
}
`

func TestAccLogStreamSplunk(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamSplunkConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-splunk-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "splunk"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_domain", "demo.splunk.com"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_token", "12a34ab5-c6d7-8901-23ef-456b7c89d0c1"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_port", "8088"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.splunk_secure", "true"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSplunkConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-splunk-%s", t.Name())),
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
	name = "Acceptance-Test-LogStream-splunk-{{.testName}}"
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
	name = "Acceptance-Test-LogStream-splunk-{{.testName}}"
	type = "splunk"
	sink {
	  splunk_domain = "prod.splunk.com"
	  splunk_token = "12a34ab5-c6d7-8901-23ef-456b7c89d0d1"
	  splunk_port = "8088"
	  splunk_secure = "true"
	}
}
`

func TestAccLogStreamSegment(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamSegmentConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-segment-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "segment"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.segment_write_key", "121233123455"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSegmentConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-segment-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "segment"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.segment_write_key", "12120908909089"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSegmentConfigUpdateWithFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-segment-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "segment"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "2"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.name", "auth.login.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.name", "auth.signup.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.segment_write_key", "12120908909089"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSegmentConfigUpdateWithEmptyFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-segment-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "segment"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.segment_write_key", "12120908909089"),
				),
			},
		},
	})
}

const logStreamSegmentConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-segment-{{.testName}}"
	type = "segment"
	sink {
		segment_write_key = "121233123455"
	}
}
`
const logStreamSegmentConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-segment-{{.testName}}"
	type = "segment"
	sink {
		segment_write_key = "12120908909089"
	}
}
`

const logStreamSegmentConfigUpdateWithFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-segment-{{.testName}}"
	type = "segment"

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
		segment_write_key = "12120908909089"
	}
}
`

const logStreamSegmentConfigUpdateWithEmptyFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-segment-{{.testName}}"
	type = "segment"

	filters = [ ]

	sink {
		segment_write_key = "12120908909089"
	}
}
`

func TestAccLogStreamSumo(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamSumoConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-sumo-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "demo.sumo.com"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSumoConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-sumo-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "prod.sumo.com"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSumoConfigUpdateWithFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-sumo-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "2"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.name", "auth.login.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.name", "auth.signup.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "prod.sumo.com"),
				),
			},
			{
				Config: template.ParseTestName(logStreamSumoConfigUpdateWithEmptyFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-sumo-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "sumo"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.sumo_source_address", "prod.sumo.com"),
				),
			},
		},
	})
}

const logStreamSumoConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.testName}}"
	type = "sumo"
	sink {
	  sumo_source_address = "demo.sumo.com"
	}
}
`
const logStreamSumoConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.testName}}"
	type = "sumo"
	sink {
	  sumo_source_address = "prod.sumo.com"
	}
}
`

const logStreamSumoConfigUpdateWithFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.testName}}"
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

const logStreamSumoConfigUpdateWithEmptyFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-sumo-{{.testName}}"
	type = "sumo"

	filters = [ ]

	sink {
		sumo_source_address = "prod.sumo.com"
	}
}
`

func TestAccLogStreamMixpanel(t *testing.T) {
	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: template.ParseTestName(logStreamMixpanelConfig, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-mixpanel-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "mixpanel"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_project_id", "123456789"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_username", "fake-account.123abc.mp-service-account"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_password", "8iwyKSzwV2brfakepassGGKhsZ3INozo"),
				),
			},
			{
				Config: template.ParseTestName(logStreamMixpanelConfigUpdate, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-mixpanel-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "mixpanel"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_project_id", "987654321"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_username", "fake-account.123abc.mp-service-account"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_password", "8iwyKSzwV2brfakepassGGKhsZ3INozo"),
				),
			},
			{
				Config: template.ParseTestName(logStreamMixpanelConfigUpdateWithFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-mixpanel-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "mixpanel"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "2"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.0.name", "auth.login.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.type", "category"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.1.name", "auth.signup.fail"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_project_id", "987654321"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_username", "fake-account.123abc.mp-service-account"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_password", "8iwyKSzwV2brfakepassGGKhsZ3INozo"),
				),
			},
			{
				Config: template.ParseTestName(logStreamMixpanelConfigUpdateWithEmptyFilters, t.Name()),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "name", fmt.Sprintf("Acceptance-Test-LogStream-mixpanel-%s", t.Name())),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "type", "mixpanel"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "filters.#", "0"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_region", "us"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_project_id", "987654321"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_username", "fake-account.123abc.mp-service-account"),
					resource.TestCheckResourceAttr("auth0_log_stream.my_log_stream", "sink.0.mixpanel_service_account_password", "8iwyKSzwV2brfakepassGGKhsZ3INozo"),
				),
			},
		},
	})
}

const logStreamMixpanelConfig = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-mixpanel-{{.testName}}"
	type = "mixpanel"
	sink {
		mixpanel_region = "us"
		mixpanel_project_id = "123456789"
		mixpanel_service_account_username = "fake-account.123abc.mp-service-account"
		mixpanel_service_account_password = "8iwyKSzwV2brfakepassGGKhsZ3INozo"
	}
}
`
const logStreamMixpanelConfigUpdate = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-mixpanel-{{.testName}}"
	type = "mixpanel"
	sink {
		mixpanel_region = "us"
		mixpanel_project_id = "987654321"
		mixpanel_service_account_username = "fake-account.123abc.mp-service-account"
		mixpanel_service_account_password = "8iwyKSzwV2brfakepassGGKhsZ3INozo"
	}
}
`

const logStreamMixpanelConfigUpdateWithFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-mixpanel-{{.testName}}"
	type = "mixpanel"

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
		mixpanel_region = "us"
		mixpanel_project_id = "987654321"
		mixpanel_service_account_username = "fake-account.123abc.mp-service-account"
		mixpanel_service_account_password = "8iwyKSzwV2brfakepassGGKhsZ3INozo"
	}
}
`

const logStreamMixpanelConfigUpdateWithEmptyFilters = `
resource "auth0_log_stream" "my_log_stream" {
	name = "Acceptance-Test-LogStream-mixpanel-{{.testName}}"
	type = "mixpanel"

	filters = [ ]

	sink {
		mixpanel_region = "us"
		mixpanel_project_id = "987654321"
		mixpanel_service_account_username = "fake-account.123abc.mp-service-account"
		mixpanel_service_account_password = "8iwyKSzwV2brfakepassGGKhsZ3INozo"
	}
}
`
