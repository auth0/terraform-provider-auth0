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
	}

	if data.IsNewResource() {
		eventStream.Destination = expandEventStreamDestination(cfg.GetAttr("destination"))
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

func expandEventStreamDestination(config cty.Value) *management.EventStreamDestination {
	var destination management.EventStreamDestination

	config.ForEachElement(func(_ cty.Value, b cty.Value) (stop bool) {
		destination = management.EventStreamDestination{
			EventStreamDestinationType: value.String(b.GetAttr("type")),
		}
		dest, err := value.MapFromJSON(b.GetAttr("configuration"))
		if err == nil {
			destination.EventStreamDestinationConfiguration = dest
		}
		return stop
	})

	return &destination
}
