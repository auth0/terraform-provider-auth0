package email

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandEmailProvider(data *schema.ResourceData, config cty.Value) *management.EmailProvider {
	emailProvider := &management.EmailProvider{
		Name:               value.String(config.GetAttr("name")),
		Enabled:            value.Bool(config.GetAttr("enabled")),
		DefaultFromAddress: value.String(config.GetAttr("default_from_address")),
	}

	switch emailProvider.GetName() {
	case management.EmailProviderMandrill:
		expandEmailProviderMandrill(config, data, emailProvider)
	case management.EmailProviderSES:
		expandEmailProviderSES(config, data, emailProvider)
	case management.EmailProviderSendGrid:
		expandEmailProviderSendGrid(config, data, emailProvider)
	case management.EmailProviderSparkPost:
		expandEmailProviderSparkPost(config, data, emailProvider)
	case management.EmailProviderMailgun:
		expandEmailProviderMailgun(config, data, emailProvider)
	case management.EmailProviderSMTP:
		expandEmailProviderSMTP(config, data, emailProvider)
	case management.EmailProviderAzureCS:
		expandEmailProviderAzureCS(config, data, emailProvider)
	case management.EmailProviderMS365:
		expandEmailProviderMS365(config, data, emailProvider)
	case management.EmailProviderCustom:
		emailProvider.Credentials = &management.EmailProviderCredentialsCustom{}
	}

	return emailProvider
}

func expandEmailProviderMandrill(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsMandrill{
				APIKey: value.String(credentials.GetAttr("api_key")),
			}
			return stop
		})
	}

	config.GetAttr("settings").ForEachElement(func(_ cty.Value, settings cty.Value) (stop bool) {
		settings.GetAttr("message").ForEachElement(func(_ cty.Value, message cty.Value) (stop bool) {
			emailProvider.Settings = &management.EmailProviderSettingsMandrill{
				Message: &management.EmailProviderSettingsMandrillMessage{
					ViewContentLink: value.Bool(message.GetAttr("view_content_link")),
				},
			}
			return stop
		})
		return stop
	})
}

func expandEmailProviderSES(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsSES{
				AccessKeyID:     value.String(credentials.GetAttr("access_key_id")),
				SecretAccessKey: value.String(credentials.GetAttr("secret_access_key")),
				Region:          value.String(credentials.GetAttr("region")),
			}
			return stop
		})
	}

	config.GetAttr("settings").ForEachElement(func(_ cty.Value, settings cty.Value) (stop bool) {
		settings.GetAttr("message").ForEachElement(func(_ cty.Value, message cty.Value) (stop bool) {
			emailProvider.Settings = &management.EmailProviderSettingsSES{
				Message: &management.EmailProviderSettingsSESMessage{
					ConfigurationSetName: value.String(message.GetAttr("configuration_set_name")),
				},
			}
			return stop
		})
		return stop
	})
}

func expandEmailProviderSendGrid(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsSendGrid{
				APIKey: value.String(credentials.GetAttr("api_key")),
			}
			return stop
		})
	}
}

func expandEmailProviderSparkPost(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsSparkPost{
				APIKey: value.String(credentials.GetAttr("api_key")),
				Region: value.String(credentials.GetAttr("region")),
			}
			return stop
		})
	}
}

func expandEmailProviderMailgun(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsMailgun{
				APIKey: value.String(credentials.GetAttr("api_key")),
				Domain: value.String(credentials.GetAttr("domain")),
				Region: value.String(credentials.GetAttr("region")),
			}
			return stop
		})
	}
}

func expandEmailProviderSMTP(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
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

	config.GetAttr("settings").ForEachElement(func(_ cty.Value, settings cty.Value) (stop bool) {
		settings.GetAttr("headers").ForEachElement(func(_ cty.Value, headers cty.Value) (stop bool) {
			emailProvider.Settings = management.EmailProviderSettingsSMTP{
				Headers: &management.EmailProviderSettingsSMTPHeaders{
					XMCViewContentLink:   value.String(headers.GetAttr("x_mc_view_content_link")),
					XSESConfigurationSet: value.String(headers.GetAttr("x_ses_configuration_set")),
				},
			}
			return stop
		})
		return stop
	})
}

func expandEmailProviderAzureCS(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsAzureCS{
				ConnectionString: value.String(credentials.GetAttr("azure_cs_connection_string")),
			}
			return stop
		})
	}
}

func expandEmailProviderMS365(config cty.Value, data *schema.ResourceData, emailProvider *management.EmailProvider) {
	if data.HasChange("credentials") {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, credentials cty.Value) (stop bool) {
			emailProvider.Credentials = &management.EmailProviderCredentialsMS365{
				TenantID:     value.String(credentials.GetAttr("ms365_tenant_id")),
				ClientID:     value.String(credentials.GetAttr("ms365_client_id")),
				ClientSecret: value.String(credentials.GetAttr("ms365_client_secret")),
			}
			return stop
		})
	}
}

func expandEmailTemplate(config cty.Value) *management.EmailTemplate {
	emailTemplate := &management.EmailTemplate{
		Template:               value.String(config.GetAttr("template")),
		Body:                   value.String(config.GetAttr("body")),
		From:                   value.String(config.GetAttr("from")),
		ResultURL:              value.String(config.GetAttr("result_url")),
		Subject:                value.String(config.GetAttr("subject")),
		Syntax:                 value.String(config.GetAttr("syntax")),
		URLLifetimeInSecoonds:  value.Int(config.GetAttr("url_lifetime_in_seconds")),
		Enabled:                value.Bool(config.GetAttr("enabled")),
		IncludeEmailInRedirect: value.Bool(config.GetAttr("include_email_in_redirect")),
	}

	return emailTemplate
}
