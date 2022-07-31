package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func newEmailTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEmailTemplate,
		ReadContext:   readEmailTemplate,
		UpdateContext: updateEmailTemplate,
		DeleteContext: deleteEmailTemplate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"template": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"verify_email",
					"verify_email_by_code",
					"reset_email",
					"welcome_email",
					"blocked_account",
					"stolen_credentials",
					"enrollment_email",
					"change_password",
					"password_reset",
					"mfa_oob_code",
					"user_invitation",
				}, true),
			},
			"body": {
				Type:     schema.TypeString,
				Required: true,
			},
			"from": {
				Type:     schema.TypeString,
				Required: true,
			},
			"result_url": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"subject": {
				Type:     schema.TypeString,
				Required: true,
			},
			"syntax": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url_lifetime_in_seconds": {
				Type:     schema.TypeInt,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Required: true,
			},
			"include_email_in_redirect": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether the `reset_email` and `verify_email` templates should include the user's email address as the email parameter in the `returnUrl` (true) or whether no email address should be included in the redirect (false). Defaults to true.",
			},
		},
	}
}

func createEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	email := expandEmailTemplate(d)
	api := m.(*management.Management)

	// The email template resource doesn't allow deleting templates, so in order
	// to avoid conflicts, we first attempt to read the template. If it exists
	// we'll try to update it, if not we'll try to create it.
	if _, err := api.EmailTemplate.Read(auth0.StringValue(email.Template)); err == nil {
		// We succeeded in reading the template, this means it was created previously.
		if err := api.EmailTemplate.Update(auth0.StringValue(email.Template), email); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(auth0.StringValue(email.Template))

		return nil
	}

	// If we reached this point the template doesn't exist.
	// Therefore, it is safe to create it.
	if err := api.EmailTemplate.Create(email); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(email.Template))

	return nil
}

func readEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	email, err := api.EmailTemplate.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	d.SetId(auth0.StringValue(email.Template))

	result := multierror.Append(
		d.Set("template", email.Template),
		d.Set("body", email.Body),
		d.Set("from", email.From),
		d.Set("result_url", email.ResultURL),
		d.Set("subject", email.Subject),
		d.Set("syntax", email.Syntax),
		d.Set("url_lifetime_in_seconds", email.URLLifetimeInSecoonds),
		d.Set("enabled", email.Enabled),
		d.Set("include_email_in_redirect", email.GetIncludeEmailInRedirect()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	email := expandEmailTemplate(d)
	api := m.(*management.Management)
	if err := api.EmailTemplate.Update(d.Id(), email); err != nil {
		return diag.FromErr(err)
	}

	return readEmailTemplate(ctx, d, m)
}

func deleteEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	emailTemplate := &management.EmailTemplate{
		Template: auth0.String(d.Id()),
		Enabled:  auth0.Bool(false),
	}
	if err := api.EmailTemplate.Update(d.Id(), emailTemplate); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func expandEmailTemplate(d *schema.ResourceData) *management.EmailTemplate {
	emailTemplate := &management.EmailTemplate{
		Template:               String(d, "template"),
		Body:                   String(d, "body"),
		From:                   String(d, "from"),
		ResultURL:              String(d, "result_url"),
		Subject:                String(d, "subject"),
		Syntax:                 String(d, "syntax"),
		URLLifetimeInSecoonds:  Int(d, "url_lifetime_in_seconds"),
		Enabled:                Bool(d, "enabled"),
		IncludeEmailInRedirect: Bool(d, "include_email_in_redirect"),
	}

	return emailTemplate
}
