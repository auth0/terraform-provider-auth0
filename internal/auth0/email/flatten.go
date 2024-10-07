package email

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenEmailProvider(data *schema.ResourceData, emailProvider *management.EmailProvider) error {
	result := multierror.Append(
		data.Set("name", emailProvider.GetName()),
		data.Set("enabled", emailProvider.GetEnabled()),
		data.Set("default_from_address", emailProvider.GetDefaultFromAddress()),
		data.Set("credentials", flattenEmailProviderCredentials(data, emailProvider)),
		data.Set("settings", flattenEmailProviderSettings(emailProvider)),
	)

	return result.ErrorOrNil()
}

func flattenEmailProviderCredentials(data *schema.ResourceData, emailProvider *management.EmailProvider) []interface{} {
	if emailProvider.Credentials == nil {
		return nil
	}

	var credentials interface{}
	switch credentialsType := emailProvider.Credentials.(type) {
	case *management.EmailProviderCredentialsMandrill:
		credentials = map[string]interface{}{
			"api_key": data.Get("credentials.0.api_key").(string),
		}
	case *management.EmailProviderCredentialsSES:
		credentials = map[string]interface{}{
			"access_key_id":     data.Get("credentials.0.access_key_id").(string),
			"secret_access_key": data.Get("credentials.0.secret_access_key").(string),
			"region":            credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsSendGrid:
		credentials = map[string]interface{}{
			"api_key": data.Get("credentials.0.api_key").(string),
		}
	case *management.EmailProviderCredentialsSparkPost:
		credentials = map[string]interface{}{
			"api_key": data.Get("credentials.0.api_key").(string),
			"region":  credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsMailgun:
		credentials = map[string]interface{}{
			"api_key": data.Get("credentials.0.api_key").(string),
			"domain":  credentialsType.GetDomain(),
			"region":  credentialsType.GetRegion(),
		}
	case *management.EmailProviderCredentialsSMTP:
		credentials = map[string]interface{}{
			"smtp_host": credentialsType.GetSMTPHost(),
			"smtp_port": credentialsType.GetSMTPPort(),
			"smtp_user": credentialsType.GetSMTPUser(),
			"smtp_pass": data.Get("credentials.0.smtp_pass").(string),
		}
	case *management.EmailProviderCredentialsAzureCS:
		credentials = map[string]interface{}{
			"azure_cs_connection_string": data.Get("credentials.0.azure_cs_connection_string").(string),
		}
	case *management.EmailProviderCredentialsMS365:
		credentials = map[string]interface{}{
			"ms365_tenant_id":     data.Get("credentials.0.ms365_tenant_id").(string),
			"ms365_client_id":     data.Get("credentials.0.ms365_client_id").(string),
			"ms365_client_secret": data.Get("credentials.0.ms365_client_secret").(string),
		}
	case *management.EmailProviderCredentialsCustom:
		credentials = map[string]interface{}{}
	}

	return []interface{}{credentials}
}

func flattenEmailProviderSettings(emailProvider *management.EmailProvider) []interface{} {
	if emailProvider.Settings == nil {
		return nil
	}

	var settings interface{}
	switch settingsType := emailProvider.Settings.(type) {
	case *management.EmailProviderSettingsMandrill:
		settings = map[string]interface{}{
			"message": []map[string]interface{}{
				{
					"view_content_link": settingsType.GetMessage().GetViewContentLink(),
				},
			},
		}
	case *management.EmailProviderSettingsSES:
		settings = map[string]interface{}{
			"message": []map[string]interface{}{
				{
					"configuration_set_name": settingsType.GetMessage().GetConfigurationSetName(),
				},
			},
		}
	case *management.EmailProviderSettingsSMTP:
		settings = map[string]interface{}{
			"headers": []map[string]interface{}{
				{
					"x_mc_view_content_link":  settingsType.GetHeaders().GetXMCViewContentLink(),
					"x_ses_configuration_set": settingsType.GetHeaders().GetXSESConfigurationSet(),
				},
			},
		}
	default:
		settings = nil
	}

	return []interface{}{settings}
}
