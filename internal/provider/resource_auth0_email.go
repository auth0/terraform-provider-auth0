package provider

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func newEmail() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEmail,
		ReadContext:   readEmail,
		UpdateContext: updateEmail,
		DeleteContext: deleteEmail,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can have standard welcome, password reset, and account verification " +
			"email-based workflows built right into Auth0. This resource allows you to configure email " +
			"providers, so you can route all emails that are part of Auth0's authentication workflows " +
			"through the supported high-volume email service of your choice.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Description: "Name of the email provider. " +
					"Options include `mailgun`, `mandrill`, `sendgrid`, `ses`, `smtp`, and `sparkpost`.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether the email provider is enabled.",
			},
			"default_from_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Email address to use as the sender when no other \"from\" address is specified.",
			},
			"credentials": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Required:    true,
				Description: "Configuration settings for the credentials for the email provider.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"api_user": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "API User for your email service.",
						},
						"api_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "API Key for your email service. Will always be encrypted in our database.",
						},
						"access_key_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "AWS Access Key ID. Used only for AWS.",
						},
						"secret_access_key": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "AWS Secret Key. Will always be encrypted in our database. Used only for AWS.",
						},
						"region": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Default region. Used only for AWS, Mailgun, and SparkPost.",
						},
						"domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Domain name.",
						},
						"smtp_host": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Hostname or IP address of your SMTP server. Used only for SMTP.",
						},
						"smtp_port": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Port used by your SMTP server. Please avoid using port 25 if " +
								"possible because many providers have limitations on this port. Used only for SMTP.",
						},
						"smtp_user": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "SMTP username. Used only for SMTP.",
						},
						"smtp_pass": {
							Type:        schema.TypeString,
							Optional:    true,
							Sensitive:   true,
							Description: "SMTP password. Used only for SMTP.",
						},
					},
				},
			},
		},
	}
}

func createEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	d.SetId(resource.UniqueId())

	if emailProviderIsConfigured(api) {
		return updateEmail(ctx, d, m)
	}

	email := expandEmail(d.GetRawConfig())
	if err := api.Email.Create(email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func readEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	email, err := api.Email.Read()
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("name", email.GetName()),
		d.Set("enabled", email.GetEnabled()),
		d.Set("default_from_address", email.GetDefaultFromAddress()),
		d.Set("credentials", flattenCredentials(email.GetCredentials(), d)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	email := expandEmail(d.GetRawConfig())
	if err := api.Email.Update(email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func deleteEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if err := api.Email.Delete(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func emailProviderIsConfigured(api *management.Management) bool {
	_, err := api.Email.Read()
	if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
		return false
	}

	return true
}

func expandEmail(config cty.Value) *management.Email {
	email := &management.Email{
		Name:               value.String(config.GetAttr("name")),
		Enabled:            value.Bool(config.GetAttr("enabled")),
		DefaultFromAddress: value.String(config.GetAttr("default_from_address")),
	}

	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		email.Credentials = &management.EmailCredentials{
			APIUser:         value.String(config.GetAttr("api_user")),
			APIKey:          value.String(config.GetAttr("api_key")),
			AccessKeyID:     value.String(config.GetAttr("access_key_id")),
			SecretAccessKey: value.String(config.GetAttr("secret_access_key")),
			Region:          value.String(config.GetAttr("region")),
			Domain:          value.String(config.GetAttr("domain")),
			SMTPHost:        value.String(config.GetAttr("smtp_host")),
			SMTPPort:        value.Int(config.GetAttr("smtp_port")),
			SMTPUser:        value.String(config.GetAttr("smtp_user")),
			SMTPPass:        value.String(config.GetAttr("smtp_pass")),
		}

		return stop
	})

	return email
}

func flattenCredentials(credentials *management.EmailCredentials, d *schema.ResourceData) []interface{} {
	if credentials == nil {
		return nil
	}

	m := map[string]interface{}{
		"api_user":          credentials.GetAPIUser(),
		"api_key":           d.Get("credentials.0.api_key").(string),
		"access_key_id":     d.Get("credentials.0.access_key_id").(string),
		"secret_access_key": d.Get("credentials.0.secret_access_key").(string),
		"region":            credentials.GetRegion(),
		"domain":            credentials.GetDomain(),
		"smtp_host":         credentials.GetSMTPHost(),
		"smtp_port":         credentials.GetSMTPPort(),
		"smtp_user":         credentials.GetSMTPUser(),
		"smtp_pass":         d.Get("credentials.0.smtp_pass").(string),
	}

	return []interface{}{m}
}
