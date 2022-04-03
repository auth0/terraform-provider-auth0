package auth0

import (
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newHook() *schema.Resource {
	return &schema.Resource{
		Create: createHook,
		Read:   readHook,
		Update: updateHook,
		Delete: deleteHook,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateHookNameFunc(),
				Description:  "Name of this hook",
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

func createHook(d *schema.ResourceData, m interface{}) error {
	hook := buildHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Create(hook); err != nil {
		return err
	}

	d.SetId(auth0.StringValue(hook.ID))

	if err := upsertHookSecrets(d, m); err != nil {
		return err
	}

	return readHook(d, m)
}

func readHook(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	hook, err := api.Hook.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	result := multierror.Append(
		d.Set("name", hook.Name),
		d.Set("dependencies", hook.Dependencies),
		d.Set("script", hook.Script),
		d.Set("trigger_id", hook.TriggerID),
		d.Set("enabled", hook.Enabled),
	)

	return result.ErrorOrNil()
}

func updateHook(d *schema.ResourceData, m interface{}) error {
	hook := buildHook(d)
	api := m.(*management.Management)
	if err := api.Hook.Update(d.Id(), hook); err != nil {
		return err
	}

	if err := upsertHookSecrets(d, m); err != nil {
		return err
	}

	return readHook(d, m)
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

func deleteHook(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.Hook.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
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

func validateHookNameFunc() schema.SchemaValidateFunc {
	return validation.StringMatch(
		regexp.MustCompile("^[^\\s-][\\w -]+[^\\s-]$"),
		"Can only contain alphanumeric characters, spaces and '-'. Can neither start nor end with '-' or spaces.")
}
