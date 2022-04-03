package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func newRuleConfig() *schema.Resource {
	return &schema.Resource{
		Create: createRuleConfig,
		Read:   readRuleConfig,
		Update: updateRuleConfig,
		Delete: deleteRuleConfig,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
		},
	}
}

func createRuleConfig(d *schema.ResourceData, m interface{}) error {
	ruleConfig := buildRuleConfig(d)
	key := auth0.StringValue(ruleConfig.Key)
	ruleConfig.Key = nil
	api := m.(*management.Management)
	if err := api.RuleConfig.Upsert(key, ruleConfig); err != nil {
		return err
	}

	d.SetId(auth0.StringValue(ruleConfig.Key))

	return readRuleConfig(d, m)
}

func readRuleConfig(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	ruleConfig, err := api.RuleConfig.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	return d.Set("key", ruleConfig.Key)
}

func updateRuleConfig(d *schema.ResourceData, m interface{}) error {
	ruleConfig := buildRuleConfig(d)
	ruleConfig.Key = nil
	api := m.(*management.Management)
	if err := api.RuleConfig.Upsert(d.Id(), ruleConfig); err != nil {
		return err
	}

	return readRuleConfig(d, m)
}

func deleteRuleConfig(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	if err := api.RuleConfig.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func buildRuleConfig(d *schema.ResourceData) *management.RuleConfig {
	return &management.RuleConfig{
		Key:   String(d, "key"),
		Value: String(d, "value"),
	}
}
