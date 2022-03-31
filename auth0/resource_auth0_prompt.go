package auth0

import (
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func newPrompt() *schema.Resource {
	return &schema.Resource{
		Create: createPrompt,
		Read:   readPrompt,
		Update: updatePrompt,
		Delete: deletePrompt,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"universal_login_experience": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"new", "classic",
				}, false),
			},
			"identifier_first": {
				Type:     schema.TypeBool,
				Optional: true,
			},
		},
	}
}

func createPrompt(d *schema.ResourceData, m interface{}) error {
	d.SetId(resource.UniqueId())
	return updatePrompt(d, m)
}

func readPrompt(d *schema.ResourceData, m interface{}) error {
	api := m.(*management.Management)
	prompt, err := api.Prompt.Read()
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
		d.Set("universal_login_experience", prompt.UniversalLoginExperience),
		d.Set("identifier_first", prompt.IdentifierFirst),
	)

	return result.ErrorOrNil()
}

func updatePrompt(d *schema.ResourceData, m interface{}) error {
	prompt := buildPrompt(d)
	api := m.(*management.Management)
	if err := api.Prompt.Update(prompt); err != nil {
		return err
	}

	return readPrompt(d, m)
}

func deletePrompt(d *schema.ResourceData, m interface{}) error {
	d.SetId("")
	return nil
}

func buildPrompt(d *schema.ResourceData) *management.Prompt {
	return &management.Prompt{
		UniversalLoginExperience: auth0.StringValue(String(d, "universal_login_experience")),
		IdentifierFirst:          Bool(d, "identifier_first"),
	}
}
