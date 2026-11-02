package experimentsegment

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/auth0/go-auth0"
	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandSegment(data *schema.ResourceData) (*management.CreateSegmentRequestContent, error) {
	rules, err := expandRules(data.Get("rules").([]interface{}))
	if err != nil {
		return nil, err
	}

	cfg := data.GetRawConfig()

	return &management.CreateSegmentRequestContent{
		Name:        auth0.StringValue(value.String(cfg.GetAttr("name"))),
		Description: value.String(cfg.GetAttr("description")),
		Rules:       rules,
	}, nil
}

// expandSegmentUpdate returns nil when no patchable field changed, because the API rejects an
// empty PATCH body.
func expandSegmentUpdate(data *schema.ResourceData) (*management.UpdateSegmentRequestContent, error) {
	cfg := data.GetRawConfig()
	request := &management.UpdateSegmentRequestContent{}

	// The Set* helpers mark each field explicit so a removed description serializes as null.
	if data.HasChange("name") {
		request.SetName(value.String(cfg.GetAttr("name")))
	}

	if data.HasChange("description") {
		request.SetDescription(value.String(cfg.GetAttr("description")))
	}

	// Rules are replaced as a whole on PATCH, so always send the full list.
	if data.HasChange("rules") {
		rules, err := expandRules(data.Get("rules").([]interface{}))
		if err != nil {
			return nil, err
		}

		request.SetRules(rules)
	}

	if status := data.Get("status").(string); data.HasChange("status") && status != "" {
		segmentStatus := management.SegmentStatusEnum(status)
		request.SetStatus(&segmentStatus)
	}

	return request, nil
}

func expandRules(rawRules []interface{}) ([]*management.SegmentRule, error) {
	rules := make([]*management.SegmentRule, 0, len(rawRules))

	for i, rawRule := range rawRules {
		rule := &management.SegmentRule{}

		ruleMap, _ := rawRule.(map[string]interface{})

		if match := firstBlock(ruleMap["match"]); match != nil {
			conditions, err := expandConditions[management.SegmentMatchConditions](match)
			if err != nil {
				return nil, fmt.Errorf("rules.%d.match: %w", i, err)
			}

			rule.Match = conditions
		}

		if notMatch := firstBlock(ruleMap["not_match"]); notMatch != nil {
			conditions, err := expandConditions[management.SegmentNotMatchConditions](notMatch)
			if err != nil {
				return nil, fmt.Errorf("rules.%d.not_match: %w", i, err)
			}

			rule.NotMatch = conditions
		}

		rules = append(rules, rule)
	}

	return rules, nil
}

// expandConditions builds the conditions object keyed by attribute name, then decodes it into the
// SDK type. Decoding moves namespaced anonymous session metadata key attributes into
// ExtraProperties, which the typed fields cannot hold.
func expandConditions[T management.SegmentMatchConditions | management.SegmentNotMatchConditions](
	raw map[string]interface{},
) (*T, error) {
	conditions := make(map[string]*management.SegmentMatchExpression)

	for _, attribute := range matchAttributes {
		if expression := firstBlock(raw[attribute]); expression != nil {
			conditions[attribute] = expandExpression(expression)
		}
	}

	if metadata, ok := raw[metadataBlockName].(*schema.Set); ok {
		for _, rawEntry := range metadata.List() {
			entry := rawEntry.(map[string]interface{})
			conditions[metadataKeyPrefix+entry["key"].(string)] = expandExpression(entry)
		}
	}

	encoded, err := json.Marshal(conditions)
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// expandExpression maps the configured operator to its union branch. The validateSegmentRules
// check guarantees exactly one operator is set, so a block with no list operator is the `exists` shape.
func expandExpression(raw map[string]interface{}) *management.SegmentMatchExpression {
	if values := stringList(raw["one_of"]); len(values) > 0 {
		return &management.SegmentMatchExpression{StringList: values}
	}

	if values := stringList(raw["contains"]); len(values) > 0 {
		return &management.SegmentMatchExpression{
			SegmentContainsExpression: &management.SegmentContainsExpression{Contains: values},
		}
	}

	if values := stringList(raw["starts_with"]); len(values) > 0 {
		return &management.SegmentMatchExpression{
			SegmentStartsWithExpression: &management.SegmentStartsWithExpression{StartsWith: values},
		}
	}

	if values := stringList(raw["ends_with"]); len(values) > 0 {
		return &management.SegmentMatchExpression{
			SegmentEndsWithExpression: &management.SegmentEndsWithExpression{EndsWith: values},
		}
	}

	// Only an expression with all four lists empty reaches this line. The validateSegmentRules check
	// rejects that expression when `exists` is also null (no operator set), so `exists` is "true" or "false".
	exists, _ := strconv.ParseBool(raw["exists"].(string))

	return &management.SegmentMatchExpression{
		SegmentExistsExpression: &management.SegmentExistsExpression{Exists: exists},
	}
}

// firstBlock returns the single element of a MaxItems 1 block, or nil when it is not set.
func firstBlock(raw interface{}) map[string]interface{} {
	list, ok := raw.([]interface{})
	if !ok || len(list) == 0 {
		return nil
	}

	block, _ := list[0].(map[string]interface{})

	return block
}

func stringList(raw interface{}) []string {
	list, _ := raw.([]interface{})
	values := make([]string, 0, len(list))

	for _, item := range list {
		text, _ := item.(string)
		values = append(values, text)
	}

	return values
}
