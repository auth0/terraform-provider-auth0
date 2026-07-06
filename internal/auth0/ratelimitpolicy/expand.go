package ratelimitpolicy

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandRateLimitPolicyCreate(data *schema.ResourceData) *management.CreateRateLimitPolicyRequestContent {
	action, limit, redirectURI := readConfiguration(data.GetRawConfig().GetAttr("configuration"))

	return &management.CreateRateLimitPolicyRequestContent{
		Resource:         management.RateLimitPolicyResourceEnum(data.Get("resource").(string)),
		Consumer:         management.RateLimitPolicyConsumerEnum(data.Get("consumer").(string)),
		ConsumerSelector: data.Get("consumer_selector").(string),
		Configuration:    expandConfigurationUnion(action, limit, redirectURI),
	}
}

func expandRateLimitPolicyPatch(data *schema.ResourceData) *management.PatchRateLimitPolicyRequestContent {
	patch := &management.PatchRateLimitPolicyRequestContent{}

	// Configuration is the only updatable field; only send it when it actually changed.
	if data.HasChange("configuration") {
		action, limit, redirectURI := readConfiguration(data.GetRawConfig().GetAttr("configuration"))
		patch.Configuration = expandPatchConfigurationUnion(action, limit, redirectURI)
	}

	return patch
}

// readConfiguration extracts the configuration block as configured. Limit and redirectURI are
// nil when the user did not set them (distinct from limit = 0), so callers can forward exactly
// what was configured.
func readConfiguration(list cty.Value) (action string, limit *int, redirectURI *string) {
	if list.IsNull() || list.LengthInt() == 0 {
		return "", nil, nil
	}

	cfg := list.AsValueSlice()[0]
	return cfg.GetAttr("action").AsString(), value.Int(cfg.GetAttr("limit")), value.String(cfg.GetAttr("redirect_uri"))
}

// expandConfigurationUnion routes the configured fields into the SDK's configuration union. The
// variant is chosen by `action`, and only the fields valid for that action are emitted. So fields
// that do not apply (e.g. a generated `limit = 0` on an `allow` policy) are dropped rather than
// sent to the API. Required fields are guaranteed present by validateRateLimitPolicyConfiguration,
// so the dereferences below are safe.
func expandConfigurationUnion(action string, limit *int, redirectURI *string) *management.RateLimitPolicyConfiguration {
	switch action {
	case string(management.RateLimitPolicyConfigurationZeroActionAllow):
		return &management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationZero: &management.RateLimitPolicyConfigurationZero{
				Action: management.RateLimitPolicyConfigurationZeroActionAllow,
			},
		}
	case string(management.RateLimitPolicyConfigurationOneActionBlock),
		string(management.RateLimitPolicyConfigurationOneActionLog):
		return &management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationOne: &management.RateLimitPolicyConfigurationOne{
				Action: management.RateLimitPolicyConfigurationOneAction(action),
				Limit:  *limit,
			},
		}
	case string(management.RateLimitPolicyConfigurationActionActionRedirect):
		return &management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationAction: &management.RateLimitPolicyConfigurationAction{
				Action:      management.RateLimitPolicyConfigurationActionActionRedirect,
				Limit:       *limit,
				RedirectURI: *redirectURI,
			},
		}
	default:
		return nil
	}
}

// expandPatchConfigurationUnion mirrors expandConfigurationUnion for the PATCH-only union type.
func expandPatchConfigurationUnion(action string, limit *int, redirectURI *string) *management.PatchRateLimitPolicyConfigurationRequestContent {
	switch action {
	case string(management.PatchRateLimitPolicyConfigurationRequestContentZeroActionAllow):
		return &management.PatchRateLimitPolicyConfigurationRequestContent{
			PatchRateLimitPolicyConfigurationRequestContentZero: &management.PatchRateLimitPolicyConfigurationRequestContentZero{
				Action: management.PatchRateLimitPolicyConfigurationRequestContentZeroActionAllow,
			},
		}
	case string(management.PatchRateLimitPolicyConfigurationRequestContentOneActionBlock),
		string(management.PatchRateLimitPolicyConfigurationRequestContentOneActionLog):
		return &management.PatchRateLimitPolicyConfigurationRequestContent{
			PatchRateLimitPolicyConfigurationRequestContentOne: &management.PatchRateLimitPolicyConfigurationRequestContentOne{
				Action: management.PatchRateLimitPolicyConfigurationRequestContentOneAction(action),
				Limit:  *limit,
			},
		}
	case string(management.PatchRateLimitPolicyConfigurationRequestContentActionActionRedirect):
		return &management.PatchRateLimitPolicyConfigurationRequestContent{
			PatchRateLimitPolicyConfigurationRequestContentAction: &management.PatchRateLimitPolicyConfigurationRequestContentAction{
				Action:      management.PatchRateLimitPolicyConfigurationRequestContentActionActionRedirect,
				Limit:       *limit,
				RedirectURI: *redirectURI,
			},
		}
	default:
		return nil
	}
}
