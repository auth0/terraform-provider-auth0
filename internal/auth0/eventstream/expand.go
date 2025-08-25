package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandEventStream(data *schema.ResourceData) *management.EventStream {
	config := data.GetRawConfig()

	eventStream := &management.EventStream{
		Name:          value.String(config.GetAttr("name")),
		Status:        value.String(config.GetAttr("status")),
		Subscriptions: expandEventStreamSubscriptions(data),
		Destination:   expandEventStreamDestination(data),
	}

	return eventStream
}

func expandEventStreamSubscriptions(data *schema.ResourceData) *[]management.EventStreamSubscription {
	subscriptionsSet := data.Get("subscriptions").(*schema.Set)
	subscriptions := make([]management.EventStreamSubscription, 0, subscriptionsSet.Len())

	for _, subscription := range subscriptionsSet.List() {
		subscriptionMap := subscription.(map[string]interface{})
		eventType := subscriptionMap["event_type"].(string)
		subscriptions = append(subscriptions, management.EventStreamSubscription{
			EventStreamSubscriptionType: &eventType,
		})
	}

	return &subscriptions
}

func expandEventStreamDestination(data *schema.ResourceData) *management.EventStreamDestination {
	destinationList := data.Get("destination").([]interface{})
	if len(destinationList) == 0 {
		return nil
	}

	destinationMap := destinationList[0].(map[string]interface{})
	destinationType := destinationMap["type"].(string)

	destination := &management.EventStreamDestination{
		EventStreamDestinationType: &destinationType,
	}

	switch destinationType {
	case "webhook":
		destination.EventStreamDestinationConfiguration = expandEventStreamWebhookConfig(destinationMap)
	case "eventbridge":
		destination.EventStreamDestinationConfiguration = expandEventStreamEventBridgeConfig(destinationMap)
	case "action":
		destination.EventStreamDestinationConfiguration = expandEventStreamActionConfig(destinationMap)
	}

	return destination
}

func expandEventStreamWebhookConfig(destinationMap map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{}

	if endpoint, ok := destinationMap["webhook_endpoint"]; ok && endpoint != "" {
		config["webhook_endpoint"] = endpoint
	}

	if authList, ok := destinationMap["webhook_authorization"].([]interface{}); ok && len(authList) > 0 {
		authMap := authList[0].(map[string]interface{})
		method := authMap["method"].(string)

		auth := map[string]interface{}{
			"method": method,
		}

		switch method {
		case "basic":
			if username, ok := authMap["username"]; ok {
				auth["username"] = username
			}
			if password, ok := authMap["password"]; ok {
				auth["password"] = password
			}
		case "bearer":
			if token, ok := authMap["token"]; ok {
				auth["token"] = token
			}
		}

		config["webhook_authorization"] = auth
	}

	return config
}

func expandEventStreamEventBridgeConfig(destinationMap map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{}

	if accountID, ok := destinationMap["aws_account_id"]; ok && accountID != "" {
		config["aws_account_id"] = accountID
	}

	if region, ok := destinationMap["aws_region"]; ok && region != "" {
		config["aws_region"] = region
	}

	return config
}

func expandEventStreamActionConfig(destinationMap map[string]interface{}) map[string]interface{} {
	config := map[string]interface{}{}

	if actionID, ok := destinationMap["action_id"]; ok && actionID != "" {
		config["action_id"] = actionID
	}

	return config
}
