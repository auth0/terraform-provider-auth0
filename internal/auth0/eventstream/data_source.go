package eventstream

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewDataSource will return a new auth0_event_stream data source.
func NewDataSource() *schema.Resource {
	return &schema.Resource{
		ReadContext: readEventStreamForDataSource,
		Description: "Data source to retrieve a specific Auth0 event stream by ID.",
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique identifier for the event stream.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the event stream.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Indicates whether the event stream is actively forwarding events.",
			},
			"subscriptions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "List of event types subscribed to in this stream.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"event_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Event type subscribed to.",
						},
					},
				},
			},
			"destination": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The destination configuration for the event stream.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the destination.",
						},
						"webhook_endpoint": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Target HTTP endpoint URL for webhook destination.",
						},
						"webhook_authorization": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Authorization configuration for webhook destination.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"method": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of authorization.",
									},
									"username": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Username for basic authorization.",
									},
								},
							},
						},
						"aws_account_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "AWS Account ID for EventBridge destination.",
						},
						"aws_region": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "AWS Region for EventBridge destination.",
						},
						"aws_partner_event_source": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "AWS Partner Event Source for EventBridge destination.",
						},
						"action_id": {
							Type:        schema.TypeString,
							Computed:    true,
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

func readEventStreamForDataSource(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	eventStreamID := data.Get("id").(string)

	eventStream, err := api.EventStream.Read(ctx, eventStreamID)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	data.SetId(eventStreamID)

	return diag.FromErr(flattenEventStream(data, eventStream))
}
