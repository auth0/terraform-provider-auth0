package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func flattenEventStream(data *schema.ResourceData, es *management.EventStream) error {
	result := multierror.Append(
		data.Set("name", es.GetName()),
		data.Set("status", es.GetStatus()),
		data.Set("created_at", es.GetCreatedAt().String()),
		data.Set("updated_at", es.GetUpdatedAt().String()),
		data.Set("subscriptions", flattenEventStreamSubscriptions(es.GetSubscriptions())),
		data.Set("destination", flattenEventStreamDestination(es.GetDestination())),
	)

	return result.ErrorOrNil()
}

func flattenEventStreamSubscriptions(subs []management.EventStreamSubscription) []interface{} {
	if subs == nil {
		return nil
	}

	result := make([]interface{}, 0, len(subs))
	for _, s := range subs {
		result = append(result, s.GetEventStreamSubscriptionType())
	}
	return result
}

func flattenEventStreamDestination(dest *management.EventStreamDestination) []interface{} {
	if dest == nil {
		return nil
	}
	configurationMap, _ := structure.FlattenJsonToString(dest.GetEventStreamDestinationConfiguration())

	return []interface{}{
		map[string]interface{}{
			"type":          dest.GetEventStreamDestinationType(),
			"configuration": configurationMap,
		},
	}
}
