package experimentfeatureflag

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenFeatureFlag(data *schema.ResourceData, flag *management.GetFeatureFlagResponseContent, variations []*management.Variation) error {
	parameters, err := flattenParameters(flag.GetParameters())
	if err != nil {
		return err
	}

	flatVariations, err := flattenVariations(variations)
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("name", flag.GetName()),
		data.Set("description", auth0.StringValue(flag.Description)),
		data.Set("type", string(flag.GetType())),
		data.Set("status", string(flag.GetStatus())),
		data.Set("parameters", parameters),
		data.Set("variation", flatVariations),
		data.Set("created_at", flag.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", flag.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

func flattenParameters(parameters map[string]*management.FeatureFlagConfigParam) ([]interface{}, error) {
	result := make([]interface{}, 0, len(parameters))

	for name, cfg := range parameters {
		value, err := typedToString(cfg.GetValue())
		if err != nil {
			return nil, err
		}

		result = append(result, map[string]interface{}{
			"name":        name,
			"type":        string(cfg.GetType()),
			"value":       value,
			"description": cfg.GetDescription(),
		})
	}

	return result, nil
}

// flattenVariations emits the variations in the order they were fetched, which mirrors the
// order Terraform tracks them in state (and therefore configuration).
func flattenVariations(variations []*management.Variation) ([]interface{}, error) {
	result := make([]interface{}, 0, len(variations))

	for _, v := range variations {
		flat, err := flattenVariation(v)
		if err != nil {
			return nil, err
		}
		result = append(result, flat)
	}

	return result, nil
}

func flattenVariation(v *management.Variation) (map[string]interface{}, error) {
	overrides := make(map[string]interface{}, len(v.GetOverrides()))
	for name, override := range v.GetOverrides() {
		str, err := typedToString(override)
		if err != nil {
			return nil, err
		}
		overrides[name] = str
	}

	return map[string]interface{}{
		"id":          v.GetID(),
		"name":        v.GetName(),
		"description": auth0.StringValue(v.Description),
		"overrides":   overrides,
		"created_at":  v.GetCreatedAt().Format(time.RFC3339),
		"updated_at":  v.GetUpdatedAt().Format(time.RFC3339),
	}, nil
}

// typedToString renders a raw typed API value (bool, number, string, or object) as the string
// the Terraform schema stores. Objects are rendered as compact JSON.
func typedToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case bool:
		return strconv.FormatBool(v), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case string:
		return v, nil
	default:
		encoded, err := json.Marshal(v)
		if err != nil {
			return "", err
		}
		return string(encoded), nil
	}
}
