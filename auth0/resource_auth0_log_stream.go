package auth0

import (
	"context"
	"log"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newLogStream() *schema.Resource {
	return &schema.Resource{
		CreateContext: createLogStream,
		ReadContext:   readLogStream,
		UpdateContext: updateLogStream,
		DeleteContext: deleteLogStream,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"eventbridge",
					"eventgrid",
					"http",
					"datadog",
					"splunk",
					"sumo",
				}, true),
				ForceNew:    true,
				Description: "Type of the log stream, which indicates the sink provider",
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"active",
					"paused",
					"suspended",
				}, false),
				Description: "Status of the LogStream",
			},
			"filters": {
				Type:     schema.TypeList,
				Optional: true,
				Description: "Only logs events matching these filters will be delivered by the stream." +
					" If omitted or empty, all events will be delivered.",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"sink": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_account_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.aws_region"},
						},
						"aws_region": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.aws_account_id"},
						},
						"aws_partner_event_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Name of the Partner Event Source to be used with AWS, if the type is 'eventbridge'",
						},
						"azure_subscription_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.azure_resource_group", "sink.0.azure_region"},
						},
						"azure_resource_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.azure_subscription_id", "sink.0.azure_region"},
						},
						"azure_region": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.azure_subscription_id", "sink.0.azure_resource_group"},
						},
						"azure_partner_topic": {
							Type:        schema.TypeString,
							Computed:    true,
							Optional:    true,
							Description: "Name of the Partner Topic to be used with Azure, if the type is 'eventgrid'",
						},
						"http_content_format": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.http_endpoint", "sink.0.http_authorization", "sink.0.http_content_type"},
							Description:  "HTTP Content Format can be JSONLINES, JSONARRAY or JSONOBJECT",
							ValidateFunc: validation.StringInSlice([]string{
								"JSONLINES",
								"JSONARRAY",
								"JSONOBJECT",
							}, false),
						},
						"http_content_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "HTTP Content Type",
							RequiredWith: []string{"sink.0.http_endpoint", "sink.0.http_authorization", "sink.0.http_content_format"},
						},
						"http_endpoint": {
							Type:         schema.TypeString,
							Optional:     true,
							Description:  "HTTP endpoint",
							RequiredWith: []string{"sink.0.http_content_format", "sink.0.http_authorization", "sink.0.http_content_type"},
						},
						"http_authorization": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.http_content_format", "sink.0.http_endpoint", "sink.0.http_content_type"},
						},
						"http_custom_headers": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Schema{Type: schema.TypeString},
							},
							Optional:    true,
							Default:     nil,
							Description: "Custom HTTP headers",
						},

						"datadog_region": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.datadog_api_key"},
							ValidateFunc: validation.StringInSlice(
								[]string{"us", "eu", "us3", "us5"},
								false,
							),
						},
						"datadog_api_key": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.datadog_region"},
						},
						"splunk_domain": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.splunk_token", "sink.0.splunk_port", "sink.0.splunk_secure"},
						},
						"splunk_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_port", "sink.0.splunk_secure"},
						},
						"splunk_port": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_token", "sink.0.splunk_secure"},
						},
						"splunk_secure": {
							Type:         schema.TypeBool,
							Optional:     true,
							Default:      nil,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_port", "sink.0.splunk_token"},
						},
						"sumo_source_address": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  nil,
						},
					},
				},
			},
		},
	}
}

func createLogStream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logStream := expandLogStream(d)

	api := m.(*management.Management)
	if err := api.LogStream.Create(logStream); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(logStream.GetID())

	// The Management API only allows updating a log stream's status. Therefore,
	// if the status field was present in the configuration, we perform an
	// additional operation to modify it.
	status := String(d, "status")
	if status != nil && status != logStream.Status {
		if err := api.LogStream.Update(logStream.GetID(), &management.LogStream{Status: status}); err != nil {
			return diag.FromErr(err)
		}
	}

	return readLogStream(ctx, d, m)
}

func readLogStream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	logStream, err := api.LogStream.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", logStream.Name),
		d.Set("status", logStream.Status),
		d.Set("type", logStream.Type),
		d.Set("filters", logStream.Filters),
		d.Set("sink", flattenLogStreamSink(logStream.Sink)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateLogStream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	logStream := expandLogStream(d)
	api := m.(*management.Management)
	if err := api.LogStream.Update(d.Id(), logStream); err != nil {
		return diag.FromErr(err)
	}

	return readLogStream(ctx, d, m)
}

func deleteLogStream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.LogStream.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func flattenLogStreamSink(sink interface{}) []interface{} {
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

func expandLogStream(d ResourceData) *management.LogStream {
	logStream := &management.LogStream{
		Name:    String(d, "name"),
		Type:    String(d, "type", IsNewResource()),
		Status:  String(d, "status", Not(IsNewResource())),
		Filters: List(d, "filters").List(),
	}

	streamType := d.Get("type").(string)
	List(d, "sink").Elem(func(d ResourceData) {
		switch streamType {
		case management.LogStreamTypeAmazonEventBridge:
			// LogStreamTypeAmazonEventBridge cannot be updated.
			if d.IsNewResource() {
				logStream.Sink = expandLogStreamSinkAmazonEventBridge(d)
			}
		case management.LogStreamTypeAzureEventGrid:
			// LogStreamTypeAzureEventGrid cannot be updated.
			if d.IsNewResource() {
				logStream.Sink = expandLogStreamSinkAzureEventGrid(d)
			}
		case management.LogStreamTypeHTTP:
			logStream.Sink = expandLogStreamSinkHTTP(d)
		case management.LogStreamTypeDatadog:
			logStream.Sink = expandLogStreamSinkDatadog(d)
		case management.LogStreamTypeSplunk:
			logStream.Sink = expandLogStreamSinkSplunk(d)
		case management.LogStreamTypeSumo:
			logStream.Sink = expandLogStreamSinkSumo(d)
		default:
			log.Printf("[WARN]: Unsupported log stream sink %s", streamType)
			log.Printf("[WARN]: Raise an issue with the auth0 provider in order to support it:")
			log.Printf("[WARN]: 	https://github.com/auth0/terraform-provider-auth0/issues/new")
		}
	})

	return logStream
}

func expandLogStreamSinkAmazonEventBridge(d ResourceData) *management.LogStreamSinkAmazonEventBridge {
	return &management.LogStreamSinkAmazonEventBridge{
		AccountID: String(d, "aws_account_id"),
		Region:    String(d, "aws_region"),
	}
}

func expandLogStreamSinkAzureEventGrid(d ResourceData) *management.LogStreamSinkAzureEventGrid {
	return &management.LogStreamSinkAzureEventGrid{
		SubscriptionID: String(d, "azure_subscription_id"),
		ResourceGroup:  String(d, "azure_resource_group"),
		Region:         String(d, "azure_region"),
		PartnerTopic:   String(d, "azure_partner_topic"),
	}
}

func expandLogStreamSinkHTTP(d ResourceData) *management.LogStreamSinkHTTP {
	return &management.LogStreamSinkHTTP{
		ContentFormat: String(d, "http_content_format"),
		ContentType:   String(d, "http_content_type"),
		Endpoint:      String(d, "http_endpoint"),
		Authorization: String(d, "http_authorization"),
		CustomHeaders: List(d, "http_custom_headers").List(),
	}
}
func expandLogStreamSinkDatadog(d ResourceData) *management.LogStreamSinkDatadog {
	return &management.LogStreamSinkDatadog{
		Region: String(d, "datadog_region"),
		APIKey: String(d, "datadog_api_key"),
	}
}
func expandLogStreamSinkSplunk(d ResourceData) *management.LogStreamSinkSplunk {
	return &management.LogStreamSinkSplunk{
		Domain: String(d, "splunk_domain"),
		Token:  String(d, "splunk_token"),
		Port:   String(d, "splunk_port"),
		Secure: Bool(d, "splunk_secure"),
	}
}
func expandLogStreamSinkSumo(d ResourceData) *management.LogStreamSinkSumo {
	return &management.LogStreamSinkSumo{
		SourceAddress: String(d, "sumo_source_address"),
	}
}
