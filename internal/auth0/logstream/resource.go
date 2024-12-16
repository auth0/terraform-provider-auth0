package logstream

import (
	"context"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var validLogStreamTypes = []string{
	"eventbridge",
	"eventgrid",
	"http",
	"datadog",
	"splunk",
	"sumo",
	"mixpanel",
	"segment",
}

// NewResource will return a new auth0_log_stream resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createLogStream,
		ReadContext:   readLogStream,
		UpdateContext: updateLogStream,
		DeleteContext: deleteLogStream,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage your Auth0 log streams.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the log stream.",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(validLogStreamTypes, true),
				ForceNew:     true,
				Description: "Type of the log stream, which indicates the sink provider. " +
					"Options include: `" + strings.Join(validLogStreamTypes, "`, `") + "`.",
			},
			"is_priority": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set True for priority log streams, False for non-priority",
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
				Description: "The current status of the log stream. Options are \"active\", \"paused\", \"suspended\".",
			},
			"filters": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Description: "Only logs events matching these filters will be delivered by the stream." +
					" If omitted or empty, all events will be delivered. " +
					"Filters available: `auth.ancillary.fail`, `auth.ancillary.success`, `auth.login.fail`, " +
					"`auth.login.notification`, `auth.login.success`, `auth.logout.fail`, `auth.logout.success`, " +
					"`auth.signup.fail`, `auth.signup.success`, `auth.silent_auth.fail`, `auth.silent_auth.success`, " +
					"`auth.token_exchange.fail`, `auth.token_exchange.success`, `management.fail`, `management.success`, " +
					"`system.notification`, `user.fail`, `user.notification`, `user.success`, `other`.",
				Elem: &schema.Schema{
					Type: schema.TypeMap,
					Elem: &schema.Schema{
						Type: schema.TypeString,
					},
				},
			},
			"sink": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The sink configuration for the log stream.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_account_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.aws_region"},
							Description:  "The AWS Account ID.",
						},
						"aws_region": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ap-east-1",
								"ap-northeast-1",
								"ap-northeast-2",
								"ap-northeast-3",
								"ap-south-1",
								"ap-southeast-1",
								"ap-southeast-2",
								"ca-central-1",
								"cn-north-1",
								"cn-northwest-1",
								"eu-central-1",
								"eu-north-1",
								"eu-west-1",
								"eu-west-2",
								"eu-west-3",
								"me-south-1",
								"sa-east-1",
								"us-gov-east-1",
								"us-gov-west-1",
								"us-east-1",
								"us-east-2",
								"us-west-1",
								"us-west-2",
							}, false),
							RequiredWith: []string{"sink.0.aws_account_id"},
							Description: "The region in which the EventBridge event source will be created. " +
								"Possible values: `ap-east-1`, `ap-northeast-1`, `ap-northeast-2`, `ap-northeast-3`, " +
								"`ap-south-1`, `ap-southeast-1`, `ap-southeast-2`, `ca-central-1`, `cn-north-1`, " +
								"`cn-northwest-1`, `eu-central-1`, `eu-north-1`, `eu-west-1`, `eu-west-2`, `eu-west-3`, " +
								"`me-south-1`, `sa-east-1`, `us-gov-east-1`, `us-gov-west-1`, `us-east-1`, `us-east-2`, " +
								"`us-west-1`, `us-west-2`.",
						},
						"aws_partner_event_source": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							Description: "Name of the Partner Event Source to be used with AWS. " +
								"Generally generated by Auth0 and passed to AWS, so this should " +
								"be an output attribute.",
						},
						"azure_subscription_id": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.azure_resource_group", "sink.0.azure_region"},
							Description:  "The unique alphanumeric string that identifies your Azure subscription.",
						},
						"azure_resource_group": {
							Type:         schema.TypeString,
							Optional:     true,
							ForceNew:     true,
							RequiredWith: []string{"sink.0.azure_subscription_id", "sink.0.azure_region"},
							Description: "The Azure EventGrid resource group which allows you to manage all " +
								"Azure assets within one subscription.",
						},
						"azure_region": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"australiacentral",
								"australiaeast",
								"australiasoutheast",
								"brazilsouth",
								"canadacentral",
								"canadaeast",
								"centralindia",
								"centralus",
								"eastasia",
								"eastus",
								"eastus2",
								"francecentral",
								"germanywestcentral",
								"japaneast",
								"japanwest",
								"koreacentral",
								"koreasouth",
								"northcentralus",
								"northeurope",
								"norwayeast",
								"southafricanorth",
								"southcentralus",
								"southeastasia",
								"southindia",
								"switzerlandnorth",
								"uaenorth",
								"uksouth",
								"ukwest",
								"westcentralus",
								"westeurope",
								"westindia",
								"westus",
								"westus2",
							}, false),
							RequiredWith: []string{"sink.0.azure_subscription_id", "sink.0.azure_resource_group"},
							Description: "The Azure region code. Possible values: `australiacentral`, `australiaeast`, " +
								"`australiasoutheast`, `brazilsouth`, `canadacentral`, `canadaeast`, `centralindia`, " +
								"`centralus`, `eastasia`, `eastus`, `eastus2`, `francecentral`, " +
								"`germanywestcentral`, `japaneast`, `japanwest`, `koreacentral`, `koreasouth`, " +
								"`northcentralus`, `northeurope`, `norwayeast`, `southafricanorth`, `southcentralus`, " +
								"`southeastasia`, `southindia`, `switzerlandnorth`, `uaenorth`, `uksouth`, `ukwest`, " +
								"`westcentralus`, `westeurope`, `westindia`, `westus`, `westus2`.",
						},
						"azure_partner_topic": {
							Type:     schema.TypeString,
							Computed: true,
							Optional: true,
							Description: "Name of the Partner Topic to be used with Azure. " +
								"Generally should not be specified.",
						},
						"http_content_format": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							Description: "The format of data sent over HTTP. Options are " +
								"\"JSONLINES\", \"JSONARRAY\" or \"JSONOBJECT\"",
							ValidateFunc: validation.StringInSlice([]string{
								"JSONLINES",
								"JSONARRAY",
								"JSONOBJECT",
							}, false),
						},
						"http_content_type": {
							Type:     schema.TypeString,
							Optional: true,
							Description: "The \"Content-Type\" header to send over HTTP. " +
								"Common value is \"application/json\".",
						},
						"http_endpoint": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
							Description:  "The HTTP endpoint to send streaming logs.",
						},
						"http_authorization": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "Sent in the HTTP \"Authorization\" header with each request.",
						},
						"http_custom_headers": {
							Type: schema.TypeList,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Schema{
									Type: schema.TypeString,
								},
							},
							Optional:    true,
							Computed:    true,
							Default:     nil,
							Description: "Additional HTTP headers to be included as part of the HTTP request.",
						},
						"datadog_region": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.datadog_api_key"},
							ValidateFunc: validation.StringInSlice(
								[]string{"us", "eu", "us3", "us5"},
								false,
							),
							Description: "The Datadog region. Possible values: `us`, `eu`, `us3`, `us5`.",
						},
						"datadog_api_key": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.datadog_region"},
							Description:  "The Datadog API key.",
						},
						"splunk_domain": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.splunk_token", "sink.0.splunk_port", "sink.0.splunk_secure"},
							Description:  "The Splunk domain name.",
						},
						"splunk_token": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_port", "sink.0.splunk_secure"},
							Description:  "The Splunk access token.",
						},
						"splunk_port": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_token", "sink.0.splunk_secure"},
							Description:  "The Splunk port.",
						},
						"splunk_secure": {
							Type:         schema.TypeBool,
							Optional:     true,
							Default:      nil,
							RequiredWith: []string{"sink.0.splunk_domain", "sink.0.splunk_port", "sink.0.splunk_token"},
							Description:  "This toggle should be turned off when using self-signed certificates.",
						},
						"sumo_source_address": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  nil,
							Description: "Generated URL for your defined HTTP source in " +
								"Sumo Logic for collecting streaming data from Auth0.",
						},
						"mixpanel_region": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.mixpanel_service_account_password", "sink.0.mixpanel_project_id", "sink.0.mixpanel_service_account_username"},
							Description: "The Mixpanel region. Options are [\"us\", \"eu\"]. " +
								"EU is required for customers with EU data residency requirements.",
						},
						"mixpanel_project_id": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.mixpanel_region", "sink.0.mixpanel_service_account_username", "sink.0.mixpanel_service_account_password"},
							Description:  "The Mixpanel project ID, found on the Project Settings page.",
						},
						"mixpanel_service_account_username": {
							Type:         schema.TypeString,
							Optional:     true,
							RequiredWith: []string{"sink.0.mixpanel_region", "sink.0.mixpanel_project_id", "sink.0.mixpanel_service_account_password"},
							Description:  "The Mixpanel Service Account username. Services Accounts can be created in the Project Settings page.",
						},
						"mixpanel_service_account_password": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							RequiredWith: []string{"sink.0.mixpanel_region", "sink.0.mixpanel_project_id", "sink.0.mixpanel_service_account_username"},
							Description:  "The Mixpanel Service Account password.",
						},
						"segment_write_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "The [Segment Write Key](https://segment.com/docs/connections/find-writekey/).",
						},
					},
				},
			},
		},
	}
}

func createLogStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	logStream := expandLogStream(data)

	if err := api.LogStream.Create(ctx, logStream); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(logStream.GetID())

	// The Management API only allows updating a log stream's status.
	// Therefore, if the status field was present in the configuration,
	// we perform an additional operation to modify it.
	if status := data.Get("status").(string); status != "" && status != logStream.GetStatus() {
		logStreamWithStatus := &management.LogStream{Status: &status}
		return diag.FromErr(api.LogStream.Update(ctx, logStream.GetID(), logStreamWithStatus))
	}

	return readLogStream(ctx, data, meta)
}

func readLogStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	logStream, err := api.LogStream.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenLogStream(data, logStream))
}

func updateLogStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	logStream := expandLogStream(data)

	if err := api.LogStream.Update(ctx, data.Id(), logStream); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readLogStream(ctx, data, meta)
}

func deleteLogStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.LogStream.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
