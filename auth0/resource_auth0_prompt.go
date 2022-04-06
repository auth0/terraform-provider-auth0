package auth0

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newPrompt() *schema.Resource {
	return &schema.Resource{
		CreateContext: createPrompt,
		ReadContext:   readPrompt,
		UpdateContext: updatePrompt,
		DeleteContext: deletePrompt,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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

func createPrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(resource.UniqueId())
	return updatePrompt(ctx, d, m)
}

func readPrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	prompt, err := api.Prompt.Read()
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
		d.Set("universal_login_experience", prompt.UniversalLoginExperience),
		d.Set("identifier_first", prompt.IdentifierFirst),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updatePrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	prompt := buildPrompt(d)
	api := m.(*management.Management)
	if err := api.Prompt.Update(prompt); err != nil {
		return diag.FromErr(err)
	}

	return readPrompt(ctx, d, m)
}

func deletePrompt(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId("")
	return nil
}

func buildPrompt(d *schema.ResourceData) *management.Prompt {
	return &management.Prompt{
		UniversalLoginExperience: auth0.StringValue(String(d, "universal_login_experience")),
		IdentifierFirst:          Bool(d, "identifier_first"),
	}
}
