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
		rawConfig := data.GetRawConfig()
		if !rawConfig.IsNull() {
			webhookConfigRaw := rawConfig.GetAttr("webhook_configuration")
			if !webhookConfigRaw.IsNull() && webhookConfigRaw.LengthInt() > 0 {
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
							methodRaw := authCfgRaw.GetAttr("method")
							method := methodRaw.AsString()
							authMap := map[string]interface{}{
								"method": method,
							}

							if method == "basic" {
								if username := getString(authCfgRaw.GetAttr("username")); username != "" {
									authMap["username"] = username
								}
								passwordWO := getString(authCfgRaw.GetAttr("password_wo"))
								if passwordWO != "" && (data.IsNewResource() || hasPasswordWOVersionChanged(data)) {
									authMap["password"] = passwordWO
								} else if password := getString(authCfgRaw.GetAttr("password")); password != "" {
									authMap["password"] = password
								}
							} else if method == "bearer" {
								// Prefer token_wo if set.
								tokenWO := getString(authCfgRaw.GetAttr("token_wo"))
								if tokenWO != "" && (data.IsNewResource() || hasTokenWOVersionChanged(data)) {
									authMap["token"] = tokenWO
								} else if token := getString(authCfgRaw.GetAttr("token")); token != "" {
									authMap["token"] = token
								}
							}
							configMap["webhook_authorization"] = authMap
							return true // Only process the first auth element.
						})
					}
					return true // Only process the first webhook config element.
				})
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

// hasPasswordWOVersionChanged checks if the password_wo_version attribute has changed.
// This is used to determine if we need to update the password during resource updates.
func hasPasswordWOVersionChanged(data *schema.ResourceData) bool {
	return data.HasChange("webhook_configuration.0.webhook_authorization.0.password_wo_version")
}

// getStringIfNotNull returns the string value of a cty.Value if it is not null, otherwise returns an empty string.
func getString(val cty.Value) string {
	if !val.IsNull() {
		return val.AsString()
	}
	return ""
}
