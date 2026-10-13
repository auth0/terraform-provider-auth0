package experimentsegment

import (
	"context"
	"fmt"
	"strings"

	management "github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

const (
	metadataBlockName = "anonymous_session_metadata"
	metadataKeyPrefix = metadataBlockName + "."
)

// matchAttributes are the fixed attribute keys a segment rule condition can match on. Beyond
// these, a namespaced anonymous session metadata key attribute is modeled by its own block.
var matchAttributes = []string{
	"client_id",
	"connection",
	"connection_type",
	"organization_id",
	"domain",
	"device_type",
	"browser",
	"platform",
	"user_agent",
	"country",
	"region",
}

// expressionOperators are the mutually exclusive keys of a match expression.
var expressionOperators = []string{"one_of", "contains", "starts_with", "ends_with", "exists"}

// statusValues are the allowed lifecycle states of a segment.
var statusValues = []string{
	string(management.SegmentStatusEnumActive),
	string(management.SegmentStatusEnumArchived),
}

// maxConditionAttributes is the API limit on attributes per `match`/`not_match` object.
const maxConditionAttributes = 10

// NewResource will return a new auth0_experiment_segment resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createSegment,
		ReadContext:   readSegment,
		UpdateContext: updateSegment,
		DeleteContext: deleteSegment,
		CustomizeDiff: customdiff.All(
			validateSegmentRules,
			validateSegmentStatusOnCreate,
		),
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "Create and manage Experiment Center segments (EA only).",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(3, 255),
				Description:  "The name of the segment. Unique per tenant. Between 3 and 255 characters",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(3, 1024),
				Description:  "Description of the segment. Between 3 and 1024 characters when present.",
			},
			"rules": {
				Type:        schema.TypeList,
				Required:    true,
				MaxItems:    20,
				Description: "The rules of the segment. Between 1 and 20 rules.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"match": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "Attribute conditions that must match. Set between 1 and 10 attributes, including each `" +
								metadataBlockName + "` block. At least one of `match` or `not_match` is required per rule.",
							Elem: conditionsResource(),
						},
						"not_match": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "Attribute conditions that must not match. Set between 1 and 10 attributes, including each `" +
								metadataBlockName + "` block. At least one of `match` or `not_match` is required per rule.",
							Elem: conditionsResource(),
						},
					},
				},
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice(statusValues, false),
				Description: fmt.Sprintf(
					"The lifecycle status of the segment. One of %q. On creation, omit this attribute or set it "+
						"to `%s`. Archiving is terminal: an archived segment cannot return to `%s`, and its `name`, "+
						"`description` and `rules` can no longer change. An archived segment can still be deleted "+
						"and keeps its name reserved until then.",
					strings.Join(statusValues, ", "), management.SegmentStatusEnumActive, management.SegmentStatusEnumActive,
				),
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The type of the segment.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the segment was created.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ISO 8601 formatted date the segment was updated.",
			},
		},
	}
}

// conditionsResource returns the schema shared by the `match` and `not_match` blocks: one optional
// block per fixed attribute (each carrying a match expression) plus a repeatable
// anonymous session metadata block for namespaced attributes.
func conditionsResource() *schema.Resource {
	conditions := make(map[string]*schema.Schema, len(matchAttributes)+1)

	for _, attribute := range matchAttributes {
		conditions[attribute] = &schema.Schema{
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Condition on the `" + attribute + "` attribute. Set exactly one operator.",
			Elem:        expressionResource(),
		}
	}

	conditions[metadataBlockName] = &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Description: "Condition on a namespaced `" + metadataKeyPrefix + "<key>` attribute. " +
			"Repeat the block once per key. Each key can appear only once per `match/not_match` block.",
		Elem: anonymousSessionMetadataResource(),
	}

	return &schema.Resource{Schema: conditions}
}

// expressionSchema returns the five mutually exclusive match-expression fields. Exactly one must be
// set per attribute (enforced by validateSegmentRules). Reused for every attribute block and for
// anonymous session metadata.
func expressionSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"one_of": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    50,
			Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringLenBetween(0, 1024)},
			Description: "Values for the `one_of` operator. Between 1 and 50 values, each up to 1024 characters. An empty list counts as not set.",
		},
		"contains": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    50,
			Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringLenBetween(0, 1024)},
			Description: "Values for the `contains` operator. Between 1 and 50 values, each up to 1024 characters. An empty list counts as not set.",
		},
		"starts_with": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    50,
			Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringLenBetween(0, 1024)},
			Description: "Values for the `starts_with` operator. Between 1 and 50 values, each up to 1024 characters. An empty list counts as not set.",
		},
		"ends_with": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    50,
			Elem:        &schema.Schema{Type: schema.TypeString, ValidateFunc: validation.StringLenBetween(0, 1024)},
			Description: "Values for the `ends_with` operator. Between 1 and 50 values, each up to 1024 characters. An empty list counts as not set.",
		},
		// A string, not a bool: SDK v2 state stores an unset nested bool as false, so generated
		// import config writes `exists = false` into every block and an unset value cannot be told
		// apart from a real `exists = false` condition. Generated config writes an unset string as null.
		"exists": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringInSlice([]string{"true", "false"}, false),
			Description: "Value for the `exists` operator: `true` or `false`. Bare boolean values are " +
				"accepted.",
		},
	}
}

// expressionResource wraps expressionSchema for use as a nested attribute block.
func expressionResource() *schema.Resource {
	return &schema.Resource{Schema: expressionSchema()}
}

// anonymousSessionMetadataResource is the expression schema plus the required namespaced key.
func anonymousSessionMetadataResource() *schema.Resource {
	schemaWithKey := expressionSchema()

	schemaWithKey["key"] = &schema.Schema{
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.All(validation.StringLenBetween(1, 255), validation.StringIsNotWhiteSpace),
		Description: "The metadata key, without the `" + metadataKeyPrefix + "` prefix. Between 1 and 255 " +
			"characters, not blank.",
	}

	return &schema.Resource{Schema: schemaWithKey}
}

// validateSegmentRules rejects rule shapes the API refuses with a generic oneOf error, so they
// fail at plan time with a field-level message. Unknown values are skipped.
func validateSegmentRules(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	config := diff.GetRawConfig()
	if config.IsNull() {
		return nil
	}

	rules := config.GetAttr("rules")
	if !rules.IsKnown() || rules.IsNull() {
		return nil
	}

	for i, rule := range rules.AsValueSlice() {
		if !rule.IsKnown() {
			continue
		}

		match, notMatch := rule.GetAttr("match"), rule.GetAttr("not_match")
		if isAbsentBlock(match) && isAbsentBlock(notMatch) {
			return fmt.Errorf("rules.%d: at least one of `match` or `not_match` must be set", i)
		}

		for _, side := range []string{"match", "not_match"} {
			if err := validateConditions(rule.GetAttr(side)); err != nil {
				return fmt.Errorf("rules.%d.%s: %w", i, side, err)
			}
		}
	}

	return nil
}

func validateConditions(conditions cty.Value) error {
	if !conditions.IsKnown() || isAbsentBlock(conditions) {
		return nil
	}

	block := conditions.Index(cty.NumberIntVal(0))
	if !block.IsKnown() {
		return nil
	}

	count := 0

	for _, attribute := range matchAttributes {
		expression := block.GetAttr(attribute)
		if isAbsentBlock(expression) {
			continue
		}

		count++

		if !expression.IsKnown() {
			continue
		}

		if err := validateExpression(expression.Index(cty.NumberIntVal(0))); err != nil {
			return fmt.Errorf("%s: %w", attribute, err)
		}
	}

	metadata := block.GetAttr(metadataBlockName)
	if !isAbsentBlock(metadata) && metadata.IsKnown() {
		seen := make(map[string]bool)

		for iterator := metadata.ElementIterator(); iterator.Next(); {
			_, entry := iterator.Element()
			count++

			if !entry.IsKnown() {
				continue
			}

			key := entry.GetAttr("key")
			if key.IsKnown() && !key.IsNull() {
				if seen[key.AsString()] {
					return fmt.Errorf("%s: key %q is set more than once", metadataBlockName, key.AsString())
				}

				seen[key.AsString()] = true
			}

			if err := validateExpression(entry); err != nil {
				return fmt.Errorf("%s: %w", metadataBlockName, err)
			}
		}
	}

	if count == 0 {
		return fmt.Errorf("at least one attribute must be set")
	}

	if count > maxConditionAttributes {
		return fmt.Errorf("at most %d attributes can be set, got %d", maxConditionAttributes, count)
	}

	return nil
}

// validateExpression counts an empty list as not set. Generated import config writes each unset
// list operator as an empty list.
func validateExpression(expression cty.Value) error {
	if !expression.IsKnown() {
		return nil
	}

	set := 0

	for _, operator := range expressionOperators {
		if isUnsetOperator(expression.GetAttr(operator)) {
			continue
		}

		set++
	}

	if set != 1 {
		return fmt.Errorf("exactly one of %q must be set", strings.Join(expressionOperators, ", "))
	}

	return nil
}

func isUnsetOperator(value cty.Value) bool {
	switch {
	case value.IsNull():
		return true
	case !value.IsKnown():
		return false
	case value.Type().IsListType():
		return value.LengthInt() == 0
	default:
		return false
	}
}

func validateSegmentStatusOnCreate(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	// ResourceDiff has no IsNewResource; an empty ID marks a create or replacement plan.
	if diff.Id() != "" {
		return nil
	}
	config := diff.GetRawConfig()
	if config.IsNull() {
		return nil
	}

	status := config.GetAttr("status")
	if !status.IsKnown() || status.IsNull() || status.AsString() == string(management.SegmentStatusEnumActive) {
		return nil
	}

	return fmt.Errorf("cannot set status %q when creating a segment; "+
		"omit it or use %q. You can update the segment after creation",
		status.AsString(), management.SegmentStatusEnumActive)
}

// isAbsentBlock reports whether a nested block is not configured. Unknown blocks count as present.
func isAbsentBlock(block cty.Value) bool {
	return block.IsNull() || (block.IsKnown() && block.LengthInt() == 0)
}

func createSegment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandSegment(data)
	if err != nil {
		return diag.FromErr(err)
	}

	created, err := api.Experimentation.Segments.Create(ctx, request)
	if err != nil {
		return diag.FromErr(err)
	}

	data.SetId(created.GetID())

	return readSegment(ctx, data, meta)
}

func readSegment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	segment, err := api.Experimentation.Segments.Get(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_experiment_segment", data, err)
	}

	return diag.FromErr(flattenSegment(data, segment))
}

func updateSegment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	request, err := expandSegmentUpdate(data)
	if err != nil {
		return append(diag.FromErr(err), readSegment(ctx, data, meta)...)
	}

	_, err = api.Experimentation.Segments.Update(ctx, data.Id(), request)
	return append(diag.FromErr(internalError.HandleAPIError(data, err)), readSegment(ctx, data, meta)...)
}

func deleteSegment(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPIV3()

	err := api.Experimentation.Segments.Delete(ctx, data.Id())
	return diag.FromErr(internalError.HandleAPIError(data, err))
}
