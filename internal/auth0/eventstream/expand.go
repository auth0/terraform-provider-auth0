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
					// Prefer token_wo (write-only) over token for better security.
					if tokenWO := getTokenWO(data); tokenWO != "" {
						// For new resources, always use token_wo if provided.
						// For updates, only use token_wo if version changed (to trigger update).
						if data.IsNewResource() || hasTokenWOVersionChanged(data) {
							authMap["token"] = tokenWO
						}
						// If version didn't change, and it's an update, don't include token.
						// To avoid accidentally clearing it (since we can't read it back).
					} else if v, ok := auth["token"].(string); ok && v != "" {
						// Fall back to regular token for backward compatibility.
						authMap["token"] = v
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

// hasTokenWOVersionChanged checks if the token_wo_version attribute has changed.
// This is used to determine if we need to update the token during resource updates.
func hasTokenWOVersionChanged(data *schema.ResourceData) bool {
	return data.HasChange("webhook_configuration.0.webhook_authorization.0.token_wo_version")
}

// getTokenWO retrieves the token_wo value from the resource data's raw config.
// Returns an empty string if not set.
// This is necessary because token_wo is write-only and not stored in state,
// and can only be retrieved from the raw config.
func getTokenWO(data *schema.ResourceData) string {
	rawConfig := data.GetRawConfig()
	if rawConfig.IsNull() {
		return ""
	}

	webhookConfigRaw := rawConfig.GetAttr("webhook_configuration")
	if webhookConfigRaw.IsNull() || webhookConfigRaw.LengthInt() == 0 {
		return ""
	}

	var tokenWOStr string
	webhookConfigRaw.ForEachElement(func(_ cty.Value, webHookCfgRaw cty.Value) (stop bool) {
		authRaw := webHookCfgRaw.GetAttr("webhook_authorization")
		if !authRaw.IsNull() && authRaw.LengthInt() > 0 {
			authRaw.ForEachElement(func(_ cty.Value, authCfgRaw cty.Value) (stop bool) {
				tokenWO := authCfgRaw.GetAttr("token_wo")
				if !tokenWO.IsNull() {
					tokenWOStr = tokenWO.AsString()
				}
				return true // Stop after first.
			})
			return true // Stop after first.
		}
		return false
	})

	return tokenWOStr
}
