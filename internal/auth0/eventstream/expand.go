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
					// For write-only token, read from raw config to ensure we get the value
					// even if it's not in state (since write-only fields are not stored)
					cfg := data.GetRawConfig()
					webhookCfgRaw := cfg.GetAttr("webhook_configuration")
					if !webhookCfgRaw.IsNull() && webhookCfgRaw.LengthInt() > 0 {
						authRaw := webhookCfgRaw.Index(cty.NumberIntVal(0)).GetAttr("webhook_authorization")
						if !authRaw.IsNull() && authRaw.LengthInt() > 0 {
							tokenWORaw := authRaw.Index(cty.NumberIntVal(0)).GetAttr("token_wo")
							if !tokenWORaw.IsNull() {
								if tokenWO := value.String(tokenWORaw); tokenWO != nil && *tokenWO != "" {
									authMap["token"] = *tokenWO
								}
							}
						}
					}
					// Fallback to regular token if write-only token is not provided
					if _, hasToken := authMap["token"]; !hasToken {
						if v, ok := auth["token"].(string); ok && v != "" {
							authMap["token"] = v
						}
					}
				}

				configMap["webhook_authorization"] = authMap
			}
		}

	case "eventbridge":
		// Skip returning destination configuration for existing EventBridge resources during updates
		// since EventBridge configuration cannot be updated.
		// This prevents overwriting or reconfiguring the resource unintentionally.
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
