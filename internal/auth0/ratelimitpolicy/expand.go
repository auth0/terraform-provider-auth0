package ratelimitpolicy

import (
	"github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandRateLimitPolicyCreate(data *schema.ResourceData) *management.CreateRateLimitPolicyRequestContent {
	action, limit, redirectURI, _ := readConfiguration(data.GetRawConfig().GetAttr("configuration"))

	return &management.CreateRateLimitPolicyRequestContent{
		Resource:         management.RateLimitPolicyResourceEnum(data.Get("resource").(string)),
		Consumer:         management.RateLimitPolicyConsumerEnum(data.Get("consumer").(string)),
		ConsumerSelector: data.Get("consumer_selector").(string),
		Configuration:    expandConfigurationUnion(action, limit, redirectURI),
	}
}

func expandRateLimitPolicyPatch(data *schema.ResourceData) *management.PatchRateLimitPolicyRequestContent {
	patch := &management.PatchRateLimitPolicyRequestContent{}

	// configuration is the only updatable field; only send it when it actually changed.
	if data.HasChange("configuration") {
		action, limit, redirectURI, _ := readConfiguration(data.GetRawConfig().GetAttr("configuration"))
		patch.Configuration = expandPatchConfigurationUnion(action, limit, redirectURI)
	}

	return patch
}

// readConfiguration extracts the configuration block as configured. limit and redirectURI are
// nil when the user did not set them (distinct from limit = 0), so callers can forward exactly
// what was configured.
func readConfiguration(list cty.Value) (action string, limit *int, redirectURI *string, ok bool) {
	if list.IsNull() || list.LengthInt() == 0 {
		return "", nil, nil, false
	}

	cfg := list.AsValueSlice()[0]
	return cfg.GetAttr("action").AsString(), value.Int(cfg.GetAttr("limit")), value.String(cfg.GetAttr("redirect_uri")), true
}

// expandConfigurationUnion routes the configured fields into the SDK's configuration union. The
// variant is chosen purely by which fields the user set — not by the action value — so the API,
// not the provider, decides which action/field combinations are valid. The action string is
// forwarded verbatim; an invalid combination (e.g. action "allow" with a limit) is sent as-is
// and surfaced as an API error rather than silently dropped.
func expandConfigurationUnion(action string, limit *int, redirectURI *string) *management.RateLimitPolicyConfiguration {
	if action == "" && limit == nil && redirectURI == nil {
		return nil
	}

	if redirectURI != nil {
		return &management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationAction: &management.RateLimitPolicyConfigurationAction{
				Action:      management.RateLimitPolicyConfigurationActionAction(action),
				Limit:       intOrZero(limit),
				RedirectURI: *redirectURI,
			},
		}
	}

	if limit != nil {
		return &management.RateLimitPolicyConfiguration{
			RateLimitPolicyConfigurationOne: &management.RateLimitPolicyConfigurationOne{
				Action: management.RateLimitPolicyConfigurationOneAction(action),
				Limit:  *limit,
			},
		}
	}

	return &management.RateLimitPolicyConfiguration{
		RateLimitPolicyConfigurationZero: &management.RateLimitPolicyConfigurationZero{
			Action: management.RateLimitPolicyConfigurationZeroAction(action),
		},
	}
}

// expandPatchConfigurationUnion mirrors expandConfigurationUnion for the PATCH-only union type.
func expandPatchConfigurationUnion(action string, limit *int, redirectURI *string) *management.PatchRateLimitPolicyConfigurationRequestContent {
	if action == "" && limit == nil && redirectURI == nil {
		return nil
	}

	if redirectURI != nil {
		return &management.PatchRateLimitPolicyConfigurationRequestContent{
			PatchRateLimitPolicyConfigurationRequestContentAction: &management.PatchRateLimitPolicyConfigurationRequestContentAction{
				Action:      management.PatchRateLimitPolicyConfigurationRequestContentActionAction(action),
				Limit:       intOrZero(limit),
				RedirectURI: *redirectURI,
			},
		}
	}

	if limit != nil {
		return &management.PatchRateLimitPolicyConfigurationRequestContent{
			PatchRateLimitPolicyConfigurationRequestContentOne: &management.PatchRateLimitPolicyConfigurationRequestContentOne{
				Action: management.PatchRateLimitPolicyConfigurationRequestContentOneAction(action),
				Limit:  *limit,
			},
		}
	}

	return &management.PatchRateLimitPolicyConfigurationRequestContent{
		PatchRateLimitPolicyConfigurationRequestContentZero: &management.PatchRateLimitPolicyConfigurationRequestContentZero{
			Action: management.PatchRateLimitPolicyConfigurationRequestContentZeroAction(action),
		},
	}
}

func intOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
