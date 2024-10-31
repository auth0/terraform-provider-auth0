package email

import (
	"context"
	"math"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/id"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_email resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createEmailProvider,
		ReadContext:   readEmailProvider,
		UpdateContext: updateEmailProvider,
		DeleteContext: deleteEmailProvider,
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
					[]string{"azure_cs", "custom", "mailgun", "mandrill", "ms365", "sendgrid", "ses", "smtp", "sparkpost"},
					false,
				),
				Description: "Name of the email provider. " +
					"Options include `azure_cs`, `custom`, `mailgun`, `mandrill`, `ms365`, `sendgrid`, `ses`, `smtp` and `sparkpost`.",
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
						"azure_cs_connection_string": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Azure Communication Services Connection String.",
						},
						"ms365_tenant_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Microsoft 365 Tenant ID.",
						},
						"ms365_client_id": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Microsoft 365 Client ID.",
						},
						"ms365_client_secret": {
							Type:         schema.TypeString,
							Optional:     true,
							Sensitive:    true,
							ValidateFunc: validation.StringIsNotEmpty,
							Description:  "Microsoft 365 Client Secret.",
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

func createEmailProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	data.SetId(id.UniqueId())

	if emailProviderIsConfigured(ctx, api) {
		return updateEmailProvider(ctx, data, meta)
	}

	email := expandEmailProvider(data.GetRawConfig())

	if err := api.EmailProvider.Create(ctx, email); err != nil {
		return diag.FromErr(err)
	}

	return readEmailProvider(ctx, data, meta)
}

func readEmailProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	email, err := api.EmailProvider.Read(ctx)
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenEmailProvider(data, email))
}

func updateEmailProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	email := expandEmailProvider(data.GetRawConfig())

	if err := api.EmailProvider.Update(ctx, email); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readEmailProvider(ctx, data, meta)
}

func deleteEmailProvider(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.EmailProvider.Delete(ctx); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func emailProviderIsConfigured(ctx context.Context, api *management.Management) bool {
	_, err := api.EmailProvider.Read(ctx)
	return !internalError.IsStatusNotFound(err)
}
