package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandEventStream(data *schema.ResourceData) *management.EventStream {
	cfg := data.GetRawConfig()

	eventStream := &management.EventStream{
		Name:          value.String(cfg.GetAttr("name")),
		Subscriptions: expandEventStreamSubscriptions(cfg.GetAttr("subscriptions")),
		Destination:   expandEventStreamDestination(data),
	}

	return eventStream
}

func expandEventStreamSubscriptions(subs cty.Value) *[]management.EventStreamSubscription {
	subscriptions := make([]management.EventStreamSubscription, 0)

	subs.ForEachElement(func(_ cty.Value, attr cty.Value) (stop bool) {
		subscriptions = append(subscriptions, management.EventStreamSubscription{
			EventStreamSubscriptionType: value.String(attr),
		})
		return stop
	})

	return &subscriptions
}

func expandEventStreamDestination(data *schema.ResourceData) *management.EventStreamDestination {
	destType := data.Get("destination_type").(string)

	destination := &management.EventStreamDestination{
		EventStreamDestinationType: &destType,
	}

	configMap := make(map[string]interface{})

	switch destType {
	case "webhook":
		webhookCfgList, ok := data.Get("webhook_configuration").([]interface{})
		if ok && len(webhookCfgList) > 0 {
			webhookCfg := webhookCfgList[0].(map[string]interface{})

			if endpoint, ok := webhookCfg["webhook_endpoint"].(string); ok && endpoint != "" {
				configMap["webhook_endpoint"] = endpoint
			}

			if authList, ok := webhookCfg["webhook_authorization"].([]interface{}); ok && len(authList) > 0 {
				auth := authList[0].(map[string]interface{})
				method := auth["method"].(string)

				authMap := map[string]interface{}{
					"method": method,
				}

				if method == "basic" {
					if v, ok := auth["username"].(string); ok {
						authMap["username"] = v
					}
					if v, ok := auth["password"].(string); ok {
						authMap["password"] = v
					}
				} else if method == "bearer" {
					if v, ok := auth["token"].(string); ok {
						authMap["token"] = v
					}
				}

				configMap["webhook_authorization"] = authMap
			}
		}

	case "eventbridge":
		if !data.IsNewResource() {
			return nil
		}
		bridgeCfgList, ok := data.Get("eventbridge_configuration").([]interface{})
		if ok && len(bridgeCfgList) > 0 {
			bridgeCfg := bridgeCfgList[0].(map[string]interface{})
			configMap["aws_account_id"] = bridgeCfg["aws_account_id"]
			configMap["aws_region"] = bridgeCfg["aws_region"]
		}
	}

	destination.EventStreamDestinationConfiguration = configMap
	return destination
}
