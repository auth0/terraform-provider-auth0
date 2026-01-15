package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/auth0/terraform-provider-auth0/internal/value"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		rawConfig := data.GetRawConfig()
		if rawConfig.IsNull() {
			break
		}

		webhookConfigRaw := rawConfig.GetAttr("webhook_configuration")
		if webhookConfigRaw.IsNull() || webhookConfigRaw.LengthInt() == 0 {
			break
		}

		webhookConfigRaw.ForEachElement(func(_ cty.Value, webhookCfgRaw cty.Value) (stop bool) {
			// Webhook_endpoint.
			endpointRaw := webhookCfgRaw.GetAttr("webhook_endpoint")
			if !endpointRaw.IsNull() && endpointRaw.AsString() != "" {
				configMap["webhook_endpoint"] = endpointRaw.AsString()
			}

			// Webhook_authorization.
			authRaw := webhookCfgRaw.GetAttr("webhook_authorization")
			if !authRaw.IsNull() && authRaw.LengthInt() > 0 {
				authRaw.ForEachElement(func(_ cty.Value, authCfgRaw cty.Value) (stop bool) {
					configMap["webhook_authorization"] = extractWebhookAuth(authCfgRaw, data)
					return true // Only process the first auth element.
				})
			}
			return true // Only process the first webhook config element.
		})

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

func extractWebhookAuth(authCfgRaw cty.Value, data *schema.ResourceData) map[string]interface{} {
	methodRaw := authCfgRaw.GetAttr("method")
	method := methodRaw.AsString()
	authMap := map[string]interface{}{
		"method": method,
	}

	switch method {
	case "basic":
		if username := authCfgRaw.GetAttr("username"); !nullOrEmptyString(username) {
			authMap["username"] = username.AsString()
		}
		passwordWO := authCfgRaw.GetAttr("password_wo")
		if !nullOrEmptyString(passwordWO) && (data.IsNewResource() || hasPasswordWOVersionChanged(data)) {
			authMap["password"] = passwordWO.AsString()
		} else if password := authCfgRaw.GetAttr("password"); !nullOrEmptyString(password) {
			authMap["password"] = password.AsString()
		}
	case "bearer":
		tokenWO := authCfgRaw.GetAttr("token_wo")
		if !nullOrEmptyString(tokenWO) && (data.IsNewResource() || hasTokenWOVersionChanged(data)) {
			authMap["token"] = tokenWO.AsString()
		} else if token := authCfgRaw.GetAttr("token"); !nullOrEmptyString(token) {
			authMap["token"] = token.AsString()
		}
	}
	return authMap
}

func hasTokenWOVersionChanged(data *schema.ResourceData) bool {
	return data.HasChange("webhook_configuration.0.webhook_authorization.0.token_wo_version")
}

func hasPasswordWOVersionChanged(data *schema.ResourceData) bool {
	return data.HasChange("webhook_configuration.0.webhook_authorization.0.password_wo_version")
}

func nullOrEmptyString(s cty.Value) bool {
	return s.IsNull() || s.AsString() == ""
}
