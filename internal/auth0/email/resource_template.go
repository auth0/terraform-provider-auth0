package email

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewTemplateResource will return a new auth0_email_template resource.
func NewTemplateResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEmailTemplate,
		ReadContext:   readEmailTemplate,
		UpdateContext: updateEmailTemplate,
		DeleteContext: deleteEmailTemplate,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can have standard welcome, password reset, and account verification " +
			"email-based workflows built right into Auth0. This resource allows you to configure email templates " +
			"to customize the look, feel, and sender identities of emails sent by Auth0. " +
			"Used in conjunction with configured email providers.",
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
				Description: "Template name. Options include `verify_email`, `verify_email_by_code`, `reset_email`, " +
					"`welcome_email`, `blocked_account`, `stolen_credentials`, `enrollment_email`, `mfa_oob_code`, " +
					"`user_invitation`, `change_password` (legacy), or `password_reset` (legacy).",
			},
			"body": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Body of the email template. " +
					"You can include [common variables](https://auth0.com/docs/customize/email/email-templates#common-variables).",
			},
			"from": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Email address to use as the sender. " +
					"You can include [common variables](https://auth0.com/docs/customize/email/email-templates#common-variables).",
			},
			"result_url": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "URL to redirect the user to after a successful action. " +
					"[Learn more](https://auth0.com/docs/customize/email/email-templates#configure-template-fields).",
			},
			"subject": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Subject line of the email. " +
					"You can include [common variables](https://auth0.com/docs/customize/email/email-templates#common-variables).",
			},
			"syntax": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Syntax of the template body. You can use either text or HTML with Liquid syntax.",
			},
			"url_lifetime_in_seconds": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of seconds during which the link within the email will be valid.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Indicates whether the template is enabled.",
			},
			"include_email_in_redirect": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Whether the `reset_email` and `verify_email` templates should include the user's " +
					"email address as the email parameter in the `returnUrl` (true) or whether no email address " +
					"should be included in the redirect (false). Defaults to `true`.",
			},
		},
	}
}

func createEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	email := expandEmailTemplate(d.GetRawConfig())

	// The email template resource doesn't allow deleting templates, so in order
	// to avoid conflicts, we first attempt to read the template. If it exists
	// we'll try to update it, if not we'll try to create it.
	if _, err := api.EmailTemplate.Read(email.GetTemplate()); err == nil {
		// We succeeded in reading the template, this means it was created previously.
		if err := api.EmailTemplate.Update(email.GetTemplate(), email); err != nil {
			return diag.FromErr(err)
		}

		d.SetId(email.GetTemplate())

		return nil
	}

	// If we reached this point the template doesn't exist.
	// Therefore, it is safe to create it.
	if err := api.EmailTemplate.Create(email); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(email.GetTemplate())

	return readEmailTemplate(ctx, d, m)
}

func readEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	email, err := api.EmailTemplate.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(email.GetTemplate())

	result := multierror.Append(
		d.Set("template", email.GetTemplate()),
		d.Set("body", email.GetBody()),
		d.Set("from", email.GetFrom()),
		d.Set("result_url", email.GetResultURL()),
		d.Set("subject", email.GetSubject()),
		d.Set("syntax", email.GetSyntax()),
		d.Set("url_lifetime_in_seconds", email.GetURLLifetimeInSecoonds()),
		d.Set("enabled", email.GetEnabled()),
		d.Set("include_email_in_redirect", email.GetIncludeEmailInRedirect()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	email := expandEmailTemplate(d.GetRawConfig())
	if err := api.EmailTemplate.Update(d.Id(), email); err != nil {
		return diag.FromErr(err)
	}

	return readEmailTemplate(ctx, d, m)
}

func deleteEmailTemplate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()
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
