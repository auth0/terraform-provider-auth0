package hook

import (
	"context"
	"fmt"
	"regexp"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandHook(d *schema.ResourceData) *management.Hook {
	config := d.GetRawConfig()

	hook := &management.Hook{
		Name:         value.String(config.GetAttr("name")),
		Script:       value.String(config.GetAttr("script")),
		Enabled:      value.Bool(config.GetAttr("enabled")),
		Dependencies: value.MapOfStrings(config.GetAttr("dependencies")),
	}

	if d.IsNewResource() {
		hook.TriggerID = value.String(config.GetAttr("trigger_id"))
	}

	return hook
}

func checkForUntrackedHookSecrets(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	secretsFromConfig := d.Get("secrets").(map[string]interface{})

	api := m.(*config.Config).GetAPI()
	secretsFromAPI, err := api.Hook.Secrets(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	var warnings diag.Diagnostics
	for key := range secretsFromAPI {
		if _, ok := secretsFromConfig[key]; !ok {
			warnings = append(warnings, diag.Diagnostic{
				Severity: diag.Warning,
				Summary:  "Unexpected Hook Secrets",
				Detail: fmt.Sprintf("Found unexpected hook secrets with key: %s. ", key) +
					"To prevent issues, manage them through terraform. If you've just imported this resource " +
					"(and your secrets match), to make this warning disappear, run a terraform apply.",
				AttributePath: cty.Path{cty.GetAttrStep{Name: "secrets"}},
			})
		}
	}

	return warnings
}

func validateHookName() schema.SchemaValidateDiagFunc {
	hookNameValidation := validation.StringMatch(
		regexp.MustCompile(`^[^\s-][\w -]+[^\s-]$`),
		"Can only contain alphanumeric characters, spaces and '-'. Can neither start nor end with '-' or spaces.",
	)
	return validation.ToDiagFunc(hookNameValidation)
}
