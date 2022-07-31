package provider

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newHook() *schema.Resource {
	return &schema.Resource{
		CreateContext: createHook,
		ReadContext:   readHook,
		UpdateContext: updateHook,
		DeleteContext: deleteHook,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateDiagFunc: validateHookName(),
				Description:      "Name of this hook",
			},
			"dependencies": {
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Optional:    true,
				Description: "Dependencies of this hook used by webtask server",
			},
			"script": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Code to be executed when this hook runs",
			},
			"trigger_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"credentials-exchange",
					"pre-user-registration",
					"post-user-registration",
					"post-change-password",
					"send-phone-message",
				}, false),
				Description: "Execution stage of this rule. Can be " +
					"credentials-exchange, pre-user-registration, " +
					"post-user-registration, post-change-password" +
					", or send-phone-message",
			},
			"secrets": {
				Type:        schema.TypeMap,
				Elem:        schema.TypeString,
				Sensitive:   true,
				Optional:    true,
				Description: "The secrets associated with the hook",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the hook is enabled, or disabled",
			},
		},
	}
}

func createHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hook := expandHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Create(hook); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(hook.GetID())

	if err := upsertHookSecrets(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, d, m)
}

func readHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	hook, err := api.Hook.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	diagnostics := checkForUntrackedHookSecrets(ctx, d, m)

	result := multierror.Append(
		d.Set("name", hook.Name),
		d.Set("dependencies", hook.Dependencies),
		d.Set("script", hook.Script),
		d.Set("trigger_id", hook.TriggerID),
		d.Set("enabled", hook.Enabled),
	)

	if err = result.ErrorOrNil(); err != nil {
		diagnostics = append(diagnostics, diag.FromErr(err)...)
	}

	return diagnostics
}

func updateHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hook := expandHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Update(d.Id(), hook); err != nil {
		return diag.FromErr(err)
	}

	if err := upsertHookSecrets(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, d, m)
}

func deleteHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Hook.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	return nil
}

func checkForUntrackedHookSecrets(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	secretsFromConfig := Map(d, "secrets")

	api := m.(*management.Management)
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

func upsertHookSecrets(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	if d.IsNewResource() || d.HasChange("secrets") {
		hookSecrets := expandHookSecrets(d)
		api := m.(*management.Management)
		return api.Hook.ReplaceSecrets(d.Id(), hookSecrets)
	}

	return nil
}

func expandHook(d *schema.ResourceData) *management.Hook {
	hook := &management.Hook{
		Name:      String(d, "name"),
		Script:    String(d, "script"),
		TriggerID: String(d, "trigger_id", IsNewResource()),
		Enabled:   Bool(d, "enabled"),
	}

	if deps := Map(d, "dependencies"); deps != nil {
		hook.Dependencies = &deps
	}

	return hook
}

func expandHookSecrets(d *schema.ResourceData) management.HookSecrets {
	hookSecrets := management.HookSecrets{}
	secrets := Map(d, "secrets")

	for key, value := range secrets {
		if strVal, ok := value.(string); ok {
			hookSecrets[key] = strVal
		}
	}

	return hookSecrets
}

func validateHookName() schema.SchemaValidateDiagFunc {
	hookNameValidation := validation.StringMatch(
		regexp.MustCompile(`^[^\s-][\w -]+[^\s-]$`),
		"Can only contain alphanumeric characters, spaces and '-'. Can neither start nor end with '-' or spaces.",
	)
	return validation.ToDiagFunc(hookNameValidation)
}
