package eventstream

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var validEventStreamDestinationTypes = []string{
	"webhook",
	"eventbridge",
	"action",
}

var validEventStreamStatuses = []string{
	"enabled",
	"disabled",
}

// NewResource will return a new auth0_event_stream resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEventStream,
		ReadContext:   readEventStream,
		UpdateContext: updateEventStream,
		DeleteContext: deleteEventStream,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can manage your Auth0 event streams.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the event stream.",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(validEventStreamStatuses, false),
				Description: "Indicates whether the event stream is actively forwarding events. " +
					"Options are `" + strings.Join(validEventStreamStatuses, "`, `") + "`.",
			},
			"subscriptions": {
				Type:        schema.TypeSet,
				Required:    true,
				MinItems:    1,
				MaxItems:    100,
				Description: "List of event types subscribed to in this stream.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Event type to subscribe to.",
						},
					},
				},
			},
			"destination": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "The destination configuration for the event stream.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							ForceNew:     true,
							ValidateFunc: validation.StringInSlice(validEventStreamDestinationTypes, false),
							Description: "Type of the destination. " +
								"Options include: `" + strings.Join(validEventStreamDestinationTypes, "`, `") + "`.",
						},
						// Webhook configuration.
						"webhook_endpoint": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.IsURLWithHTTPS,
							Description:  "Target HTTP endpoint URL for webhook destination.",
						},
						"webhook_authorization": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Authorization configuration for webhook destination.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"basic", "bearer"}, false),
										Description:  "Type of authorization. Options are `basic` or `bearer`.",
									},
									"username": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Username for basic authorization.",
									},
									"password": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "Password for basic authorization.",
									},
									"token": {
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
										Description: "Bearer token for bearer authorization.",
									},
								},
							},
						},
						// EventBridge configuration.
						"aws_account_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "AWS Account ID for EventBridge destination.",
						},
						"aws_region": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								"ap-east-1", "ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
								"ap-south-1", "ap-southeast-1", "ap-southeast-2", "ca-central-1",
								"cn-north-1", "cn-northwest-1", "eu-central-1", "eu-north-1",
								"eu-west-1", "eu-west-2", "eu-west-3", "me-south-1", "sa-east-1",
								"us-gov-east-1", "us-gov-west-1", "us-east-1", "us-east-2",
								"us-west-1", "us-west-2",
							}, false),
							Description: "AWS Region for EventBridge destination.",
						},
						"aws_partner_event_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "AWS Partner Event Source for EventBridge destination.",
						},
						// Action configuration.
						"action_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Action ID for action destination.",
						},
					},
				},
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the event stream was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the event stream was last updated.",
			},
		},
	}
}

func createEventStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	eventStream := expandEventStream(data)

	if err := api.EventStream.Create(ctx, eventStream); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(eventStream.GetID())

	return readEventStream(ctx, data, meta)
}

func readEventStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	eventStream, err := api.EventStream.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenEventStream(data, eventStream))
}

func updateEventStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	eventStream := expandEventStream(data)

	if err := api.EventStream.Update(ctx, data.Id(), eventStream); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readEventStream(ctx, data, meta)
}

func deleteEventStream(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.EventStream.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
