package logstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenLogStream(data *schema.ResourceData, logStream *management.LogStream) error {
	result := multierror.Append(
		data.Set("name", logStream.GetName()),
		data.Set("status", logStream.GetStatus()),
		data.Set("type", logStream.GetType()),
		data.Set("filters", logStream.Filters),
		data.Set("sink", flattenLogStreamSink(data, logStream.Sink)),
	)
	return result.ErrorOrNil()
}

func flattenLogStreamSink(d *schema.ResourceData, sink interface{}) []interface{} {
	var m interface{}

	switch sinkType := sink.(type) {
	case *management.LogStreamSinkAmazonEventBridge:
		m = flattenLogStreamSinkAmazonEventBridge(sinkType)
	case *management.LogStreamSinkAzureEventGrid:
		m = flattenLogStreamSinkAzureEventGrid(sinkType)
	case *management.LogStreamSinkHTTP:
		m = flattenLogStreamSinkHTTP(sinkType)
	case *management.LogStreamSinkDatadog:
		m = flattenLogStreamSinkDatadog(sinkType)
	case *management.LogStreamSinkSplunk:
		m = flattenLogStreamSinkSplunk(sinkType)
	case *management.LogStreamSinkSumo:
		m = flattenLogStreamSinkSumo(sinkType)
	case *management.LogStreamSinkMixpanel:
		m = flattenLogStreamSinkMixpanel(d, sinkType)
	case *management.LogStreamSinkSegment:
		m = flattenLogStreamSinkSegment(sinkType)
	}

	return []interface{}{m}
}

func flattenLogStreamSinkAmazonEventBridge(o *management.LogStreamSinkAmazonEventBridge) interface{} {
	return map[string]interface{}{
		"aws_account_id":           o.GetAccountID(),
		"aws_region":               o.GetRegion(),
		"aws_partner_event_source": o.GetPartnerEventSource(),
	}
}

func flattenLogStreamSinkAzureEventGrid(o *management.LogStreamSinkAzureEventGrid) interface{} {
	return map[string]interface{}{
		"azure_subscription_id": o.GetSubscriptionID(),
		"azure_resource_group":  o.GetResourceGroup(),
		"azure_region":          o.GetRegion(),
		"azure_partner_topic":   o.GetPartnerTopic(),
	}
}

func flattenLogStreamSinkHTTP(o *management.LogStreamSinkHTTP) interface{} {
	return map[string]interface{}{
		"http_endpoint":       o.GetEndpoint(),
		"http_content_format": o.GetContentFormat(),
		"http_content_type":   o.GetContentType(),
		"http_authorization":  o.GetAuthorization(),
		"http_custom_headers": o.CustomHeaders,
	}
}

func flattenLogStreamSinkDatadog(o *management.LogStreamSinkDatadog) interface{} {
	return map[string]interface{}{
		"datadog_region":  o.GetRegion(),
		"datadog_api_key": o.GetAPIKey(),
	}
}

func flattenLogStreamSinkSegment(o *management.LogStreamSinkSegment) interface{} {
	return map[string]interface{}{
		"segment_write_key": o.GetWriteKey(),
	}
}

func flattenLogStreamSinkSplunk(o *management.LogStreamSinkSplunk) interface{} {
	return map[string]interface{}{
		"splunk_domain": o.GetDomain(),
		"splunk_token":  o.GetToken(),
		"splunk_port":   o.GetPort(),
		"splunk_secure": o.GetSecure(),
	}
}

func flattenLogStreamSinkSumo(o *management.LogStreamSinkSumo) interface{} {
	return map[string]interface{}{
		"sumo_source_address": o.GetSourceAddress(),
	}
}

func flattenLogStreamSinkMixpanel(d *schema.ResourceData, o *management.LogStreamSinkMixpanel) interface{} {
	return map[string]interface{}{
		"mixpanel_region":                   o.GetRegion(),
		"mixpanel_project_id":               o.GetProjectID(),
		"mixpanel_service_account_username": o.GetServiceAccountUsername(),
		"mixpanel_service_account_password": d.Get("sink.0.mixpanel_service_account_password").(string), // Value does not get read back.
	}
}
