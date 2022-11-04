package provider

import (
	"context"
	"math"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

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
				ValidateFunc: validation.StringInSlice(
					[]string{"mailgun", "mandrill", "sendgrid", "ses", "smtp", "sparkpost"},
					false,
				),
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
							Deprecated:  "This field is not accepted by the API any more so it will be removed soon.",
							Description: "API User for your email service.",
						},
						"api_key": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Sensitive:    true,
							Description:  "API Key for your email service. Will always be encrypted in our database.",
						},
						"access_key_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "AWS Access Key ID. Used only for AWS.",
						},
						"secret_access_key": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "AWS Secret Key. Will always be encrypted in our database. Used only for AWS.",
						},
						"region": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Default region. Used only for AWS, Mailgun, and SparkPost.",
						},
						"domain": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(4, math.MaxInt),
							Description:  "Domain name.",
						},
						"smtp_host": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Hostname or IP address of your SMTP server. Used only for SMTP.",
						},
						"smtp_port": {
							Type:     schema.TypeInt,
							Optional: true,
							Description: "Port used by your SMTP server. Please avoid using port 25 if " +
								"possible because many providers have limitations on this port. Used only for SMTP.",
						},
						"smtp_user": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "SMTP username. Used only for SMTP.",
						},
						"smtp_pass": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "SMTP password. Used only for SMTP.",
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

	email := expandEmailProvider(d.GetRawConfig())
	if err := api.EmailProvider.Create(email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func readEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	email, err := api.EmailProvider.Read()
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
		d.Set("credentials", flattenEmailProviderCredentials(d, email)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	email := expandEmailProvider(d.GetRawConfig())
	if err := api.EmailProvider.Update(email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func deleteEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)

	if err := api.EmailProvider.Delete(); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func emailProviderIsConfigured(api *management.Management) bool {
	_, err := api.EmailProvider.Read()
	if err, ok := err.(management.Error); ok && err.Status() == http.StatusNotFound {
		return false
	}

	return true
}

func expandEmailProvider(config cty.Value) *management.EmailProvider {
	emailProvider := &management.EmailProvider{
		Name:               value.String(config.GetAttr("name")),
		Enabled:            value.Bool(config.GetAttr("enabled")),
		DefaultFromAddress: value.String(config.GetAttr("default_from_address")),
	}

	switch emailProvider.GetName() {
	case management.EmailProviderMandrill:
		expandEmailProviderMandrill(config, emailProvider)
	case management.EmailProviderSES:
		expandEmailProviderSES(config, emailProvider)
	case management.EmailProviderSendGrid:
		expandEmailProviderSendGrid(config, emailProvider)
	case management.EmailProviderSparkPost:
		expandEmailProviderSparkPost(config, emailProvider)
	case management.EmailProviderMailgun:
		expandEmailProviderMailgun(config, emailProvider)
	case management.EmailProviderSMTP:
		expandEmailProviderSmtp(config, emailProvider)
	}

	return emailProvider
}

func expandEmailProviderMandrill(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsMandrill{
			APIKey: value.String(credentials.GetAttr("api_key")),
		}
		return stop
	})
}

func expandEmailProviderSES(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsSES{
			AccessKeyID:     value.String(credentials.GetAttr("access_key_id")),
			SecretAccessKey: value.String(credentials.GetAttr("secret_access_key")),
			Region:          value.String(credentials.GetAttr("region")),
		}
		return stop
	})
}

func expandEmailProviderSendGrid(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsSendGrid{
			APIKey: value.String(credentials.GetAttr("api_key")),
		}
		return stop
	})
}

func expandEmailProviderSparkPost(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsSparkPost{
			APIKey: value.String(credentials.GetAttr("api_key")),
			Region: value.String(credentials.GetAttr("region")),
		}
		return stop
	})
}

func expandEmailProviderMailgun(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsMailgun{
			APIKey: value.String(credentials.GetAttr("api_key")),
			Domain: value.String(credentials.GetAttr("domain")),
			Region: value.String(credentials.GetAttr("region")),
		}
		return stop
	})
}

func expandEmailProviderSmtp(config cty.Value, emailProvider *management.EmailProvider) {
	config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
		emailProvider.Credentials = &management.EmailProviderCredentialsSMTP{
			SMTPHost: value.String(credentials.GetAttr("smtp_host")),
			SMTPPort: value.Int(credentials.GetAttr("smtp_port")),
			SMTPUser: value.String(credentials.GetAttr("smtp_user")),
			SMTPPass: value.String(credentials.GetAttr("smtp_pass")),
		}
		return stop
	})
}

func flattenEmailProviderCredentials(d *schema.ResourceData, emailProvider *management.EmailProvider) []interface{} {
	if emailProvider.Credentials == nil {
		return nil
	}

	var credentials interface{}
	switch credentialsType := emailProvider.Credentials.(type) {
	case *management.EmailProviderCredentialsMandrill:
		credentials = map[string]interface{}{
			"api_key": d.Get("credentials.0.api_key").(string),
		}
	case *management.EmailProviderCredentialsSES:
		credentials = map[string]interface{}{
			"access_key_id":     d.Get("credentials.0.access_key_id").(string),
			"secret_access_key": d.Get("credentials.0.secret_access_key").(string),
			"region":            credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsSendGrid:
		credentials = map[string]interface{}{
			"api_key": d.Get("credentials.0.api_key").(string),
		}
	case *management.EmailProviderCredentialsSparkPost:
		credentials = map[string]interface{}{
			"api_key": d.Get("credentials.0.api_key").(string),
			"region":  credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsMailgun:
		credentials = map[string]interface{}{
			"api_key": d.Get("credentials.0.api_key").(string),
			"domain":  credentialsType.GetDomain(),
			"region":  credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsSMTP:
		credentials = map[string]interface{}{
			"smtp_host": credentialsType.GetSMTPHost(),
			"smtp_port": credentialsType.GetSMTPPort(),
			"smtp_user": credentialsType.GetSMTPUser(),
			"smtp_pass": d.Get("credentials.0.smtp_pass").(string),
		}
	}

	return []interface{}{credentials}
}
