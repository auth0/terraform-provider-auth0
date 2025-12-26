package eventstream

import (
	"errors"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenEventStream(data *schema.ResourceData, es *management.EventStream) error {
	result := multierror.Append(
		data.Set("name", es.GetName()),
		data.Set("status", es.GetStatus()),
		data.Set("created_at", es.GetCreatedAt().String()),
		data.Set("updated_at", es.GetUpdatedAt().String()),
		data.Set("subscriptions", flattenEventStreamSubscriptions(es.GetSubscriptions())),
	)
	if diags := flattenEventStreamDestination(data, es.GetDestination()); diags.HasError() {
		for _, d := range diags {
			msg := d.Summary
			if d.Detail != "" {
				msg += ": " + d.Detail
			}
			result = multierror.Append(result, errors.New(msg))
		}
	}

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

func flattenEventStreamDestination(data *schema.ResourceData, dest *management.EventStreamDestination) diag.Diagnostics {
	if dest == nil {
		return nil
	}

	destType := dest.GetEventStreamDestinationType()
	config := dest.GetEventStreamDestinationConfiguration()
	if config == nil {
		return nil
	}

	if err := data.Set("destination_type", destType); err != nil {
		return diag.FromErr(err)
	}

	switch destType {
	case "eventbridge":
		eventbridgeCfg := map[string]interface{}{
			"aws_account_id":           config["aws_account_id"],
			"aws_region":               config["aws_region"],
			"aws_partner_event_source": config["aws_partner_event_source"],
		}
		if err := data.Set("eventbridge_configuration", []interface{}{eventbridgeCfg}); err != nil {
			return diag.FromErr(err)
		}

	case "webhook":
		webhookCfg := map[string]interface{}{
			"webhook_endpoint": config["webhook_endpoint"],
		}

		if auth, ok := config["webhook_authorization"].(map[string]interface{}); ok {
			authMap := map[string]interface{}{
				"method": auth["method"],
			}
			if auth["method"] == "basic" {
				if v, ok := auth["username"]; ok && v != nil {
					authMap["username"] = v
				}

				// Password is not returned from the API, so we get it from config if available.
				if p := data.Get("webhook_configuration.0.webhook_authorization.0.password"); p != nil {
					authMap["password"] = p
				}
				// Explicitly set token_wo_version to 0 for basic auth to avoid state drift.
				authMap["token_wo_version"] = 0
			} else if auth["method"] == "bearer" {
				// Token is not returned from the API.
				// For backward compatibility, preserve regular token from config if available.
				if t := data.Get("webhook_configuration.0.webhook_authorization.0.token"); t != nil {
					authMap["token"] = t
				}

				// The token_wo is write-only and should NOT be read from API or stored in state.
				// Instead, we only preserve the version from config.
				// So no action WRT token_wo here.
			}

			// The token_wo_version is stored in state to track changes.
			if version := data.Get("webhook_configuration.0.webhook_authorization.0.token_wo_version"); version != nil {
				authMap["token_wo_version"] = version
			}

			webhookCfg["webhook_authorization"] = []interface{}{authMap}
		}

		if err := data.Set("webhook_configuration", []interface{}{webhookCfg}); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}
