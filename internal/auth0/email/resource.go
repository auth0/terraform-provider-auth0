package email

import (
	"context"
	"math"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
)

// NewResource will return a new auth0_email resource.
func NewResource() *schema.Resource {
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
							Description: "API User for your email service. This field is not accepted by the API any more so it will be removed in a future major version.",
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
			"settings": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Computed:    true,
				Description: "Specific email provider settings.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"message": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Message settings for the `mandrill` or `ses` email provider.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"view_content_link": {
										Type:     schema.TypeBool,
										Optional: true,
										Description: "Setting for the `mandrill` email provider. " +
											"Set to `true` to see the content of individual emails sent to users.",
									},
									"configuration_set_name": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "Setting for the `ses` email provider. " +
											"The name of the configuration set to apply to the sent emails.",
									},
								},
							},
						},
						"headers": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Optional:    true,
							Description: "Headers settings for the `smtp` email provider.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"x_mc_view_content_link": {
										Type:     schema.TypeString,
										Optional: true,
										ValidateFunc: validation.StringInSlice(
											[]string{"true", "false"},
											false,
										),
										Description: "Disable or enable the default View Content Link " +
											"for sensitive emails.",
									},
									"x_ses_configuration_set": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "SES Configuration set to include when sending emails.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func createEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	d.SetId(id.UniqueId())

	api := m.(*config.Config).GetAPI()
	if emailProviderIsConfigured(ctx, api) {
		return updateEmail(ctx, d, m)
	}

	email := expandEmailProvider(d.GetRawConfig())
	if err := api.EmailProvider.Create(ctx, email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func readEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	email, err := api.EmailProvider.Read(ctx)
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
		d.Set("settings", flattenEmailProviderSettings(email)),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	email := expandEmailProvider(d.GetRawConfig())
	if err := api.EmailProvider.Update(ctx, email); err != nil {
		return diag.FromErr(err)
	}

	return readEmail(ctx, d, m)
}

func deleteEmail(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.EmailProvider.Delete(ctx); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}
