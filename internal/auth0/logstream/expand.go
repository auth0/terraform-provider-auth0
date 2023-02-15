package logstream

import (
	"log"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandLogStream(d *schema.ResourceData) *management.LogStream {
	config := d.GetRawConfig()

	logStream := &management.LogStream{
		Name: value.String(config.GetAttr("name")),
	}

	logStreamType := value.String(config.GetAttr("type"))
	if d.IsNewResource() {
		logStream.Type = logStreamType
	}

	if !d.IsNewResource() {
		logStream.Status = value.String(config.GetAttr("status"))
	}

	filtersConfig := config.GetAttr("filters")
	if !filtersConfig.IsNull() {
		filters := make([]map[string]string, 0)

		filtersConfig.ForEachElement(func(_ cty.Value, filter cty.Value) (stop bool) {
			filters = append(filters, *value.MapOfStrings(filter))
			return stop
		})

		logStream.Filters = &filters
	}

	config.GetAttr("sink").ForEachElement(func(_ cty.Value, sink cty.Value) (stop bool) {
		switch *logStreamType {
		case management.LogStreamTypeAmazonEventBridge:
			// LogStreamTypeAmazonEventBridge cannot be updated.
			if d.IsNewResource() {
				logStream.Sink = expandLogStreamSinkAmazonEventBridge(sink)
			}
		case management.LogStreamTypeAzureEventGrid:
			// LogStreamTypeAzureEventGrid cannot be updated.
			if d.IsNewResource() {
				logStream.Sink = expandLogStreamSinkAzureEventGrid(sink)
			}
		case management.LogStreamTypeHTTP:
			logStream.Sink = expandLogStreamSinkHTTP(sink)
		case management.LogStreamTypeDatadog:
			logStream.Sink = expandLogStreamSinkDatadog(sink)
		case management.LogStreamTypeSplunk:
			logStream.Sink = expandLogStreamSinkSplunk(sink)
		case management.LogStreamTypeSumo:
			logStream.Sink = expandLogStreamSinkSumo(sink)
		case management.LogStreamTypeMixpanel:
			logStream.Sink = expandLogStreamSinkMixpanel(sink)
		case management.LogStreamTypeSegment:
			logStream.Sink = expandLogStreamSinkSegment(sink)
		default:
			log.Printf("[WARN]: Unsupported log stream sink %s", logStream.GetType())
			log.Printf("[WARN]: Raise an issue with the auth0 provider in order to support it:")
			log.Printf("[WARN]: 	https://github.com/auth0/terraform-provider-auth0/issues/new")
		}

		return stop
	})

	return logStream
}

func expandLogStreamSinkAmazonEventBridge(config cty.Value) *management.LogStreamSinkAmazonEventBridge {
	return &management.LogStreamSinkAmazonEventBridge{
		AccountID: value.String(config.GetAttr("aws_account_id")),
		Region:    value.String(config.GetAttr("aws_region")),
	}
}

func expandLogStreamSinkAzureEventGrid(config cty.Value) *management.LogStreamSinkAzureEventGrid {
	return &management.LogStreamSinkAzureEventGrid{
		SubscriptionID: value.String(config.GetAttr("azure_subscription_id")),
		ResourceGroup:  value.String(config.GetAttr("azure_resource_group")),
		Region:         value.String(config.GetAttr("azure_region")),
		PartnerTopic:   value.String(config.GetAttr("azure_partner_topic")),
	}
}

func expandLogStreamSinkHTTP(config cty.Value) *management.LogStreamSinkHTTP {
	httpSink := &management.LogStreamSinkHTTP{
		ContentFormat: value.String(config.GetAttr("http_content_format")),
		ContentType:   value.String(config.GetAttr("http_content_type")),
		Endpoint:      value.String(config.GetAttr("http_endpoint")),
		Authorization: value.String(config.GetAttr("http_authorization")),
	}

	customHeadersConfig := config.GetAttr("http_custom_headers")
	if !customHeadersConfig.IsNull() {
		customHeaders := make([]map[string]string, 0)

		customHeadersConfig.ForEachElement(func(_ cty.Value, httpHeader cty.Value) (stop bool) {
			customHeaders = append(customHeaders, *value.MapOfStrings(httpHeader))
			return stop
		})

		httpSink.CustomHeaders = &customHeaders
	}

	return httpSink
}
func expandLogStreamSinkDatadog(config cty.Value) *management.LogStreamSinkDatadog {
	return &management.LogStreamSinkDatadog{
		Region: value.String(config.GetAttr("datadog_region")),
		APIKey: value.String(config.GetAttr("datadog_api_key")),
	}
}
func expandLogStreamSinkSegment(config cty.Value) *management.LogStreamSinkSegment {
	return &management.LogStreamSinkSegment{
		WriteKey: value.String(config.GetAttr("segment_write_key")),
	}
}
func expandLogStreamSinkSplunk(config cty.Value) *management.LogStreamSinkSplunk {
	return &management.LogStreamSinkSplunk{
		Domain: value.String(config.GetAttr("splunk_domain")),
		Token:  value.String(config.GetAttr("splunk_token")),
		Port:   value.String(config.GetAttr("splunk_port")),
		Secure: value.Bool(config.GetAttr("splunk_secure")),
	}
}
func expandLogStreamSinkSumo(config cty.Value) *management.LogStreamSinkSumo {
	return &management.LogStreamSinkSumo{
		SourceAddress: value.String(config.GetAttr("sumo_source_address")),
	}
}
func expandLogStreamSinkMixpanel(config cty.Value) *management.LogStreamSinkMixpanel {
	return &management.LogStreamSinkMixpanel{
		Region:                 value.String(config.GetAttr("mixpanel_region")),
		ProjectID:              value.String(config.GetAttr("mixpanel_project_id")),
		ServiceAccountUsername: value.String(config.GetAttr("mixpanel_service_account_username")),
		ServiceAccountPassword: value.String(config.GetAttr("mixpanel_service_account_password")),
	}
}
