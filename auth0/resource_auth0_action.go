package auth0

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/auth0/terraform-provider-auth0/auth0/internal/hash"
)

func newAction() *schema.Resource {
	return &schema.Resource{
		Create: createAction,
		Read:   readAction,
		Update: updateAction,
		Delete: deleteAction,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of an action",
			},
			"supported_triggers": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1, // NOTE: Changes must be made together with expandAction()
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Trigger ID",
						},
						"version": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Trigger version",
						},
					},
				},
				Description: "List of triggers that this action supports. At " +
					"this time, an action can only target a single trigger at" +
					" a time",
			},
			"code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The source code of the action.",
			},
			"dependencies": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Dependency name. For example lodash",
						},
						"version": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Dependency version. For example `latest` or `4.17.21`",
						},
					},
				},
				Set:         hash.StringKey("name"),
				Description: "List of third party npm modules, and their versions, that this action depends on",
			},
			"runtime": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"node12",
					"node16",
				}, false),
				Description: "The Node runtime. For example `node16`, defaults to `node12`",
			},
			"secrets": {
				Type:     schema.TypeList,
				Optional: true,
				MinItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Secret name",
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Sensitive:   true,
							Description: "Secret value",
						},
					},
				},
				Description: "List of secrets that are included in an action or a version of an action",
			},
			"deploy": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Deploying an action will create a new immutable" +
					" version of the action. If the action is currently bound" +
					" to a trigger, then the system will begin executing the " +
					"newly deployed version of the action immediately",
			},
			"version_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version ID of the action. This value is available if `deploy` is set to true",
			},
		},
	}
}

func createAction(d *schema.ResourceData, m interface{}) error {
	action := expandAction(d)
	api := m.(*management.Management)
	if err := api.Action.Create(action); err != nil {
		return err
	}

	d.SetId(action.GetID())

	d.Partial(true)
	if err := deployAction(d, m); err != nil {
		return err
	}
	d.Partial(false)

	return readAction(d, m)
}

func readAction(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	action, err := api.Action.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", action.Name)
	d.Set("supported_triggers", flattenActionTriggers(action.SupportedTriggers))
	d.Set("code", action.Code)
	d.Set("dependencies", flattenActionDependencies(action.Dependencies))
	d.Set("runtime", action.Runtime)

	if action.DeployedVersion != nil {
		d.Set("version_id", action.DeployedVersion.GetID())
	}

	return nil
}

func updateAction(d *schema.ResourceData, m interface{}) error {
	action := expandAction(d)
	api := m.(*management.Management)
	if err := api.Action.Update(d.Id(), action); err != nil {
		return err
	}

	d.Partial(true)
	if err := deployAction(d, m); err != nil {
		return err
	}
	d.Partial(false)

	return readAction(d, m)
}

func deployAction(d *schema.ResourceData, m interface{}) error {
	if d.Get("deploy").(bool) == true {
		api := m.(*management.Management)

		err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
			action, err := api.Action.Read(d.Id())
			if err != nil {
				return resource.NonRetryableError(err)
			}

			if strings.ToLower(action.GetStatus()) != "built" {
				return resource.RetryableError(
					fmt.Errorf(`expected action status %q to equal "built"`, action.GetStatus()),
				)
			}

			return nil
		})
		if err != nil {
			return fmt.Errorf("action never reached built state: %w", err)
		}

		actionVersion, err := api.Action.Deploy(d.Id())
		if err != nil {
			return err
		}

		d.Set("version_id", actionVersion.GetID())
	}

	return nil
}

func deleteAction(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.Action.Delete(d.Id()); err != nil {
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

func expandAction(d *schema.ResourceData) *management.Action {
	action := &management.Action{
		Name:    String(d, "name"),
		Code:    String(d, "code"),
		Runtime: String(d, "runtime"),
	}

	List(d, "supported_triggers").Elem(func(d ResourceData) {
		action.SupportedTriggers = []*management.ActionTrigger{
			{
				ID:      String(d, "id"),
				Version: String(d, "version"),
			},
		}
	})

	Set(d, "dependencies").Elem(func(d ResourceData) {
		action.Dependencies = append(action.Dependencies, &management.ActionDependency{
			Name:    String(d, "name"),
			Version: String(d, "version"),
		})
	})

	List(d, "secrets").Elem(func(d ResourceData) {
		action.Secrets = append(action.Secrets, &management.ActionSecret{
			Name:  String(d, "name"),
			Value: String(d, "value"),
		})
	})

	return action
}

func flattenActionTriggers(triggers []*management.ActionTrigger) []interface{} {
	var result []interface{}
	for _, trigger := range triggers {
		result = append(result, map[string]interface{}{
			"id":      trigger.ID,
			"version": trigger.Version,
		})
	}
	return result
}

func flattenActionDependencies(dependencies []*management.ActionDependency) []interface{} {
	var result []interface{}
	for _, dependency := range dependencies {
		result = append(result, map[string]interface{}{
			"name":    dependency.Name,
			"version": dependency.Version,
		})
	}
	return result
}
