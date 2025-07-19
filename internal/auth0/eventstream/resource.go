package eventstream

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

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
			"destination": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    1,
				Description: "Destination configuration block.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							Description: "Destination type (e.g., 'eventbridge', 'http'). Cannot be updated. " +
								"If altered, resource is deleted and re-created",
						},
						"configuration": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							Description: "Destination-specific configuration, as a JSON string. Cannot be updated. " +
								"If altered, resource is deleted and re-created",
							ValidateFunc:     validation.StringIsJSON,
							DiffSuppressFunc: suppressJSONDiffWithRules,
						},
					},
				},
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

func flattenJSON(prefix string, input map[string]interface{}) map[string]interface{} {
	flat := make(map[string]interface{})
	for k, v := range input {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + "." + k
		}

		switch val := v.(type) {
		case map[string]interface{}:
			for subKey, subVal := range flattenJSON(fullKey, val) {
				flat[subKey] = subVal
			}
		default:
			flat[fullKey] = val
		}
	}
	return flat
}

// suppressJSONDiffWithRules is a custom DiffSuppressFunc for Terraform that compares two JSON strings
// while intelligently ignoring certain differences based on known dynamic behaviors of the Auth0 API.
//
// This function accounts for:
//  1. Fields that are present in the Terraform configuration (user-defined) but are intentionally omitted
//     in the Auth0 API response. (e.g., secret tokens like `webhook_authorization.token`)
//  2. Fields that are added by the Auth0 API in the response, even though they were not part of the
//     original Terraform configuration. (e.g., computed fields like `aws_partner_event_source`)
//
// The comparison is performed by flattening both JSON structures into key-path maps (e.g.,
// `webhook_authorization.token`) and comparing only the shared or explicitly allowed differing keys.
// Any unexpected difference in value or unrecognized structural drift will return `false`, meaning a
// diff will be shown in the Terraform plan.
//
// This helps eliminate noisy or misleading diffs during plan/apply cycles for dynamic or
// partially opaque API schemas.
func suppressJSONDiffWithRules(_, o, n string, _ *schema.ResourceData) bool {
	var oldMap, newMap map[string]interface{}
	if err := json.Unmarshal([]byte(o), &oldMap); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(n), &newMap); err != nil {
		return false
	}

	// Rules with nested key paths.
	allowedMissingInState := map[string]bool{
		"webhook_authorization.token": true,
	}
	allowedAddedInState := map[string]bool{
		"aws_partner_event_source": true,
	}

	// Compare all key paths from new (config).
	for path, newVal := range flattenJSON("", newMap) {
		oldVal, ok := flattenJSON("", oldMap)[path]
		if !ok {
			if allowedMissingInState[path] {
				continue
			}
			return false
		}
		if !reflect.DeepEqual(oldVal, newVal) {
			return false
		}
	}

	// Check extra keys in old (API state).
	for path := range flattenJSON("", oldMap) {
		if _, ok := flattenJSON("", newMap)[path]; !ok {
			if !allowedAddedInState[path] {
				return false
			}
		}
	}

	return true
}
