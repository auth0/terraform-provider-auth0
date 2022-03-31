package auth0

import (
	"net/http"
	"regexp"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

var ruleNameRegexp = regexp.MustCompile("^[^\\s-][\\w -]+[^\\s-]$")

func newRule() *schema.Resource {
	return &schema.Resource{
		Create: createRule,
		Read:   readRule,
		Update: updateRule,
		Delete: deleteRule,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringMatch(
					ruleNameRegexp,
					"Can only contain alphanumeric characters, spaces and '-'. "+
						"Can neither start nor end with '-' or spaces."),
			},
			"script": {
				Type:     schema.TypeString,
				Required: true,
			},
			"order": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func createRule(d *schema.ResourceData, m interface{}) error {
	rule := buildRule(d)
	api := m.(*management.Management)
	if err := api.Rule.Create(rule); err != nil {
		return err
	}

	d.SetId(auth0.StringValue(rule.ID))

	return readRule(d, m)
}

func readRule(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	rule, err := api.Rule.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", rule.Name)
	d.Set("script", rule.Script)
	d.Set("order", rule.Order)
	d.Set("enabled", rule.Enabled)

	return nil
}

func updateRule(d *schema.ResourceData, m interface{}) error {
	rule := buildRule(d)
	api := m.(*management.Management)
	if err := api.Rule.Update(d.Id(), rule); err != nil {
		return err
	}

	return readRule(d, m)
}

func deleteRule(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.Rule.Delete(d.Id()); err != nil {
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

func buildRule(d *schema.ResourceData) *management.Rule {
	return &management.Rule{
		Name:    String(d, "name"),
		Script:  String(d, "script"),
		Order:   Int(d, "order"),
		Enabled: Bool(d, "enabled"),
	}
}
