package experimentsegment

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenSegment(data *schema.ResourceData, segment *management.GetSegmentResponseContent) error {
	rules, err := flattenRules(segment.GetRules())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("name", segment.GetName()),
		data.Set("description", auth0.StringValue(segment.Description)),
		data.Set("rules", rules),
		data.Set("status", string(segment.GetStatus())),
		data.Set("type", string(segment.GetType())),
		data.Set("created_at", segment.GetCreatedAt().Format(time.RFC3339)),
		data.Set("updated_at", segment.GetUpdatedAt().Format(time.RFC3339)),
	)

	return result.ErrorOrNil()
}

func flattenRules(rules []*management.SegmentRule) ([]interface{}, error) {
	result := make([]interface{}, 0, len(rules))

	for i, rule := range rules {
		if rule == nil {
			continue
		}

		flattened := make(map[string]interface{})

		if rule.Match != nil {
			conditions, err := flattenConditions(rule.Match)
			if err != nil {
				return nil, fmt.Errorf("rules.%d.match: %w", i, err)
			}

			flattened["match"] = []interface{}{conditions}
		}

		if rule.NotMatch != nil {
			conditions, err := flattenConditions(rule.NotMatch)
			if err != nil {
				return nil, fmt.Errorf("rules.%d.not_match: %w", i, err)
			}

			flattened["not_match"] = []interface{}{conditions}
		}

		result = append(result, flattened)
	}

	return result, nil
}

// flattenConditions re-encodes the SDK conditions object (typed fields plus ExtraProperties) to
// read every attribute, fixed or namespaced, through one keyed map.
func flattenConditions(conditions json.Marshaler) (map[string]interface{}, error) {
	encoded, err := conditions.MarshalJSON()
	if err != nil {
		return nil, err
	}

	var expressions map[string]management.SegmentMatchExpression
	if err := json.Unmarshal(encoded, &expressions); err != nil {
		return nil, err
	}

	result := make(map[string]interface{})
	metadata := make([]interface{}, 0)

	for attribute, expression := range expressions {
		if slices.Contains(matchAttributes, attribute) {
			result[attribute] = []interface{}{flattenExpression(expression)}
			continue
		}

		if key, ok := strings.CutPrefix(attribute, metadataKeyPrefix); ok {
			entry := flattenExpression(expression)
			entry["key"] = key
			metadata = append(metadata, entry)
		}
	}

	result[metadataBlockName] = metadata

	return result, nil
}

func flattenExpression(expression management.SegmentMatchExpression) map[string]interface{} {
	switch {
	case expression.StringList != nil:
		return map[string]interface{}{"one_of": expression.StringList}
	case expression.SegmentContainsExpression != nil:
		return map[string]interface{}{"contains": expression.SegmentContainsExpression.Contains}
	case expression.SegmentStartsWithExpression != nil:
		return map[string]interface{}{"starts_with": expression.SegmentStartsWithExpression.StartsWith}
	case expression.SegmentEndsWithExpression != nil:
		return map[string]interface{}{"ends_with": expression.SegmentEndsWithExpression.EndsWith}
	case expression.SegmentExistsExpression != nil:
		return map[string]interface{}{"exists": expression.SegmentExistsExpression.Exists}
	default:
		return map[string]interface{}{}
	}
}
