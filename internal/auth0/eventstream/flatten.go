package eventstream

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenEventStream(data *schema.ResourceData, es *management.EventStream) error {
	result := multierror.Append(
		data.Set("name", es.GetName()),
		data.Set("status", es.GetStatus()),
		data.Set("created_at", es.GetCreatedAt().String()),
		data.Set("updated_at", es.GetUpdatedAt().String()),
		data.Set("subscriptions", flattenEventStreamSubscriptions(es.GetSubscriptions())),
		flattenEventStreamDestination(data, es.GetDestination()),
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

func flattenEventStreamDestination(data *schema.ResourceData, dest *management.EventStreamDestination) error {
	if dest == nil {
		return nil
	}

	destType := dest.GetEventStreamDestinationType()
	config := dest.GetEventStreamDestinationConfiguration()
	if config == nil {
		return nil
	}

	if err := data.Set("destination_type", destType); err != nil {
		return err
	}

	switch destType {
	case "eventbridge":
		eventbridgeCfg := map[string]interface{}{
			"aws_account_id":           config["aws_account_id"],
			"aws_region":               config["aws_region"],
			"aws_partner_event_source": config["aws_partner_event_source"],
		}
		if err := data.Set("eventbridge_configuration", []interface{}{eventbridgeCfg}); err != nil {
			return err
		}

	case "webhook":
		webhookCfg := map[string]interface{}{
			"webhook_endpoint": config["webhook_endpoint"],
		}

		if auth, ok := config["webhook_authorization"].(map[string]interface{}); ok {
			webhookCfg["webhook_authorization"] = []interface{}{flattenWebhookAuthorization(auth, data)}
		}

		if err := data.Set("webhook_configuration", []interface{}{webhookCfg}); err != nil {
			return err
		}
	}

	return nil
}

func flattenWebhookAuthorization(auth map[string]interface{}, data *schema.ResourceData) map[string]interface{} {
	authMap := map[string]interface{}{
		"method": auth["method"],
	}
	method := auth["method"].(string)
	switch method {
	case "basic":
		if v, ok := auth["username"]; ok && v != "" {
			authMap["username"] = v
		}
		// Password is not returned from the API, so we get it from config if available.
		if password, ok := data.GetOk("webhook_configuration.0.webhook_authorization.0.password"); ok && password != "" {
			authMap["password"] = password
		}
		if version, ok := data.GetOk("webhook_configuration.0.webhook_authorization.0.password_wo_version"); ok {
			authMap["password_wo_version"] = version
		}

	case "bearer":
		// Token is not returned from the API, so we get it from config if available.
		if token, ok := data.GetOk("webhook_configuration.0.webhook_authorization.0.token"); ok && token != "" {
			authMap["token"] = token
		}
		if version, ok := data.GetOk("webhook_configuration.0.webhook_authorization.0.token_wo_version"); ok {
			authMap["token_wo_version"] = version
		}
	}

	return authMap
}
