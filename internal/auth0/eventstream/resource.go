package eventstream

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

var webhookConfig = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"webhook_endpoint": {
			Type: schema.TypeString,
			Description: "The HTTPS endpoint that will receive the webhook events. " +
				"Must be a valid, publicly accessible URL.",
			Required: true,
		},
		"webhook_authorization": {
			Type:     schema.TypeList,
			Required: true,
			Description: "Authorization details for the webhook endpoint. " +
				"Supports `basic` authentication using `username` and `password`, or " +
				"`bearer` authentication using a `token`. " +
				"The appropriate fields must be set based on the chosen method.",
			MaxItems: 1,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"method": {
						Type:         schema.TypeString,
						Required:     true,
						ValidateFunc: validation.StringInSlice([]string{"basic", "bearer"}, false),
						Description:  "The authorization method used to secure the webhook endpoint. Can be either `basic` or `bearer`.",
					},
					"username": {
						Type:        schema.TypeString,
						Optional:    true,
						Description: "The username for `basic` authentication. Required when `method` is set to `basic`.",
					},
					"password": {
						Type:        schema.TypeString,
						Optional:    true,
						Sensitive:   true,
						Description: "The password for `basic` authentication. Required when `method` is set to `basic`.",
					},
					"token": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
						Description: "The token used for `bearer` authentication. Required when `method` is set to `bearer`. " +
							"**Note:** For better security, consider using `token_wo` instead to prevent storing the token in Terraform state.",
					},
					"token_wo": {
						Type:      schema.TypeString,
						Optional:  true,
						Sensitive: true,
						WriteOnly: true,
						Description: "The token used for `bearer` authentication (write-only). " +
							"This value is only available during resource creation and update, and is **not** stored in Terraform state. " +
							"To update the token, increment the `token_wo_version` attribute. " +
							"Required when `method` is set to `bearer` and `token` is not provided.",
					},
					"token_wo_version": {
						Type:        schema.TypeInt,
						Optional:    true,
						Default:     1,
						Description: "Version number for token changes. Increment this value to trigger a token update when using `token_wo`.",
					},
				},
			},
		},
	},
}

var eventBridgeConfig = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"aws_account_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"aws_region": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
		},
		"aws_partner_event_source": {
			Type:     schema.TypeString,
			Computed: true,
		},
	},
}

// NewResource returns the auth0_event_stream resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEventStream,
		ReadContext:   readEventStream,
		UpdateContext: updateEventStream,
		DeleteContext: deleteEventStream,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Allows you to manage Auth0 Event Streams.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the event stream.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The current status of the event stream.",
			},
			"subscriptions": {
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of event types this stream is subscribed to.",
			},
			"destination_type": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "The type of event stream destination (either 'eventbridge' or 'webhook').",
				ValidateFunc: validation.StringInSlice([]string{"eventbridge", "webhook"}, false),
			},
			"eventbridge_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Description: "Configuration for the EventBridge destination. " +
					"This block is only applicable when `destination_type` is set to `eventbridge`. " +
					"EventBridge configurations **cannot** be updated after creation. " +
					"Any change to this block will force the resource to be recreated.",
				ExactlyOneOf: []string{"eventbridge_configuration", "webhook_configuration"},
				Elem:         eventBridgeConfig,
			},
			"webhook_configuration": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Description: "Configuration for the Webhook destination. " +
					"This block is only applicable when `destination_type` is set to `webhook`. " +
					"Webhook configurations **can** be updated after creation, including the " +
					"endpoint and authorization fields.",
				ExactlyOneOf: []string{"eventbridge_configuration", "webhook_configuration"},
				Elem:         webhookConfig,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 timestamp when the stream was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 timestamp when the stream was last updated.",
			},
		},
	}
}

func createEventStream(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	es := expandEventStream(data)
	if err := api.EventStream.Create(ctx, es); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(es.GetID())
	return readEventStream(ctx, data, m)
}

func readEventStream(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	es, err := api.EventStream.Read(ctx, d.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(d, err))
	}

	return diag.FromErr(flattenEventStream(d, es))
}

func updateEventStream(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
	es := expandEventStream(data)

	if err := api.EventStream.Update(ctx, data.Id(), es); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readEventStream(ctx, data, m)
}

func deleteEventStream(ctx context.Context, data *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.EventStream.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
