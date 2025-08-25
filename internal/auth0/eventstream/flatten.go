package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenEventStream(data *schema.ResourceData, eventStream *management.EventStream) error {
	result := make(map[string]interface{})

	result["name"] = eventStream.GetName()
	result["status"] = eventStream.GetStatus()
	result["created_at"] = eventStream.GetCreatedAt().String()
	result["updated_at"] = eventStream.GetUpdatedAt().String()

	if eventStream.Subscriptions != nil {
		result["subscriptions"] = flattenEventStreamSubscriptions(eventStream.GetSubscriptions())
	}

	if eventStream.Destination != nil {
		result["destination"] = flattenEventStreamDestination(eventStream.GetDestination())
	}

	for key, value := range result {
		if err := data.Set(key, value); err != nil {
			return err
		}
	}

	return nil
}

func flattenEventStreamSubscriptions(subscriptions []management.EventStreamSubscription) []interface{} {
	result := make([]interface{}, len(subscriptions))

	for i, subscription := range subscriptions {
		result[i] = map[string]interface{}{
			"event_type": subscription.GetEventStreamSubscriptionType(),
		}
	}

	return result
}

func flattenEventStreamDestination(destination *management.EventStreamDestination) []interface{} {
	if destination == nil {
		return []interface{}{}
	}

	result := map[string]interface{}{
		"type": destination.GetEventStreamDestinationType(),
	}

	config := destination.GetEventStreamDestinationConfiguration()
	destinationType := destination.GetEventStreamDestinationType()

	switch destinationType {
	case "webhook":
		flattenEventStreamWebhookDestination(result, config)
	case "eventbridge":
		flattenEventStreamEventBridgeDestination(result, config)
	case "action":
		flattenEventStreamActionDestination(result, config)
	}

	return []interface{}{result}
}

func flattenEventStreamWebhookDestination(result map[string]interface{}, config map[string]interface{}) {
	if endpoint, ok := config["webhook_endpoint"]; ok {
		result["webhook_endpoint"] = endpoint
	}

	if authConfig, ok := config["webhook_authorization"].(map[string]interface{}); ok {
		auth := map[string]interface{}{}

		if method, ok := authConfig["method"]; ok {
			auth["method"] = method

			switch method {
			case "basic":
				if username, ok := authConfig["username"]; ok {
					auth["username"] = username
				}
				// Don't return password for security.
			case "bearer":
				// Don't return token for security.
			}
		}

		result["webhook_authorization"] = []interface{}{auth}
	}
}

func flattenEventStreamEventBridgeDestination(result map[string]interface{}, config map[string]interface{}) {
	if accountID, ok := config["aws_account_id"]; ok {
		result["aws_account_id"] = accountID
	}

	if region, ok := config["aws_region"]; ok {
		result["aws_region"] = region
	}

	if partnerSource, ok := config["aws_partner_event_source"]; ok {
		result["aws_partner_event_source"] = partnerSource
	}
}

func flattenEventStreamActionDestination(result map[string]interface{}, config map[string]interface{}) {
	if actionID, ok := config["action_id"]; ok {
		result["action_id"] = actionID
	}
}
