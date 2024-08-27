package form

import (
	"encoding/json"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func flattenForm(data *schema.ResourceData, form *management.Form) error {
	formStyle, err := structure.FlattenJsonToString(form.GetStyle())
	if err != nil {
		return err
	}

	formTranslations, err := structure.FlattenJsonToString(form.GetTranslations())
	if err != nil {
		return err
	}

	result := multierror.Append(
		data.Set("name", form.Name),
	)
	if form.Languages != nil {
		result = multierror.Append(result, data.Set("languages", flattenFormLanguages(form.Languages)))
	}
	if formStyle != "" {
		result = multierror.Append(result, data.Set("style", formStyle))
	}
	if form.Messages != nil {
		result = multierror.Append(result, data.Set("messages", flattenFormMessages(form.Messages)))
	}
	if formTranslations != "" {
		result = multierror.Append(result, data.Set("translations", formTranslations))
	}
	if form.Start != nil {
		result = multierror.Append(result, data.Set("start", flattenFormStart(form.Start)))
	}
	if len(form.Nodes) > 0 {
		result = multierror.Append(result, data.Set("nodes", flattenFormNodes(form.Nodes)))
	}
	if form.Ending != nil {
		result = multierror.Append(result, data.Set("ending", flattenFormEnding(form.Ending)))
	}

	return result.ErrorOrNil()
}

func flattenFormLanguages(formLanguages *management.FormLanguages) []interface{} {
	if formLanguages == nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"primary": formLanguages.Primary,
			"default": formLanguages.Default,
		},
	}
}

func flattenFormMessages(formMessages *management.FormMessages) []interface{} {
	if formMessages == nil {
		return nil
	}

	formMessagesError, err := structure.FlattenJsonToString(formMessages.GetErrors())
	if err != nil {
		return nil
	}

	formMessagesCustom, err := structure.FlattenJsonToString(formMessages.GetCustom())
	if err != nil {
		return nil
	}

	return []interface{}{
		map[string]interface{}{
			"errors": formMessagesError,
			"custom": formMessagesCustom,
		},
	}
}

func flattenFormStart(formStart *map[string]interface{}) string {
	if formStart == nil {
		return ""
	}

	formBytes, err := json.Marshal(formStart)
	if err != nil {
		return ""
	}

	return string(formBytes)
}

func flattenFormNodes(formNodes []interface{}) string {
	if formNodes == nil {
		return ""
	}

	nodeBytes, err := json.Marshal(formNodes)
	if err != nil {
		return ""
	}

	return string(nodeBytes)
}

func flattenFormEnding(formEnding *map[string]interface{}) string {
	if formEnding == nil {
		return ""
	}

	formBytes, err := json.Marshal(formEnding)
	if err != nil {
		return ""
	}

	return string(formBytes)
}
