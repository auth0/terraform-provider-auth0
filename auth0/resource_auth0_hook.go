package auth0

import (
	"context"
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
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
				Optional:    true,
				Description: "The secrets associated with the hook",
				Elem:        schema.TypeString,
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
	hook := buildHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Create(hook); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(hook.ID))

	if err := upsertHookSecrets(d, m); err != nil {
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

	result := multierror.Append(
		d.Set("name", hook.Name),
		d.Set("dependencies", hook.Dependencies),
		d.Set("script", hook.Script),
		d.Set("trigger_id", hook.TriggerID),
		d.Set("enabled", hook.Enabled),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateHook(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	hook := buildHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Update(d.Id(), hook); err != nil {
		return diag.FromErr(err)
	}

	if err := upsertHookSecrets(d, m); err != nil {
		return diag.FromErr(err)
	}

	return readHook(ctx, d, m)
}

func upsertHookSecrets(d *schema.ResourceData, m interface{}) error {
	if d.IsNewResource() || d.HasChange("secrets") {
		secrets := Map(d, "secrets")
		api := m.(*management.Management)
		hookSecrets := toHookSecrets(secrets)
		return api.Hook.ReplaceSecrets(d.Id(), hookSecrets)
	}

	return nil
}

func toHookSecrets(val map[string]interface{}) management.HookSecrets {
	hookSecrets := management.HookSecrets{}
	for key, value := range val {
		if strVal, ok := value.(string); ok {
			hookSecrets[key] = strVal
		}
	}
	return hookSecrets
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

func buildHook(d *schema.ResourceData) *management.Hook {
	hook := &management.Hook{
		Name:      String(d, "name"),
		Script:    String(d, "script"),
		TriggerID: String(d, "trigger_id", IsNewResource()),
		Enabled:   Bool(d, "enabled"),
	}

	deps := Map(d, "dependencies")
	if deps != nil {
		hook.Dependencies = &deps
	}

	return hook
}

func validateHookName() schema.SchemaValidateDiagFunc {
	hookNameValidation := validation.StringMatch(
		regexp.MustCompile(`^[^\s-][\w -]+[^\s-]$`),
		"Can only contain alphanumeric characters, spaces and '-'. Can neither start nor end with '-' or spaces.",
	)
	return validation.ToDiagFunc(hookNameValidation)
}
