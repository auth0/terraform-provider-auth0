package guardian

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenMultiFactorPolicy(ctx context.Context, api *management.Management) (string, error) {
	multiFactorPolicies, err := api.Guardian.MultiFactor.Policy(ctx)
	if err != nil {
		return "", err
	}

	flattenedPolicy := "never"
	if len(*multiFactorPolicies) > 0 {
		flattenedPolicy = (*multiFactorPolicies)[0]
	}

	return flattenedPolicy, nil
}

func flattenPhone(ctx context.Context, enabled bool, api *management.Management) ([]interface{}, error) {
	phoneData := make(map[string]interface{})
	phoneData["enabled"] = enabled

	if !enabled {
		return []interface{}{phoneData}, nil
	}

	phoneMessageTypes, err := api.Guardian.MultiFactor.Phone.MessageTypes(ctx)
	if err != nil {
		return nil, err
	}
	phoneData["message_types"] = phoneMessageTypes.GetMessageTypes()

	phoneProvider, err := api.Guardian.MultiFactor.Phone.Provider(ctx)
	if err != nil {
		return nil, err
	}
	phoneData["provider"] = phoneProvider.GetProvider()

	var phoneProviderOptions []interface{}
	switch phoneProvider.GetProvider() {
	case "twilio":
		phoneProviderOptions, err = flattenTwilioOptions(ctx, api)
		if err != nil {
			return nil, err
		}
	case "auth0", "phone-message-hook":
		phoneProviderOptions, err = flattenAuth0Options(ctx, api)
		if err != nil {
			return nil, err
		}
	}
	phoneData["options"] = phoneProviderOptions

	return []interface{}{phoneData}, nil
}

func flattenAuth0Options(ctx context.Context, api *management.Management) ([]interface{}, error) {
	m := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template(ctx)
	if err != nil {
		return nil, err
	}

	m["enrollment_message"] = template.GetEnrollmentMessage()
	m["verification_message"] = template.GetVerificationMessage()

	return []interface{}{m}, nil
}

func flattenTwilioOptions(ctx context.Context, api *management.Management) ([]interface{}, error) {
	m := make(map[string]interface{})

	template, err := api.Guardian.MultiFactor.SMS.Template(ctx)
	if err != nil {
		return nil, err
	}

	m["enrollment_message"] = template.GetEnrollmentMessage()
	m["verification_message"] = template.GetVerificationMessage()

	twilio, err := api.Guardian.MultiFactor.SMS.Twilio(ctx)
	if err != nil {
		return nil, err
	}

	m["auth_token"] = twilio.GetAuthToken()
	m["from"] = twilio.GetFrom()
	m["messaging_service_sid"] = twilio.GetMessagingServiceSid()
	m["sid"] = twilio.GetSID()

	return []interface{}{m}, nil
}

func flattenWebAuthnRoaming(ctx context.Context, enabled bool, api *management.Management) ([]interface{}, error) {
	webAuthnRoamingData := make(map[string]interface{})
	webAuthnRoamingData["enabled"] = enabled

	if !enabled {
		return []interface{}{webAuthnRoamingData}, nil
	}

	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnRoaming.Read(ctx)
	if err != nil {
		return nil, err
	}

	webAuthnRoamingData["user_verification"] = webAuthnSettings.GetUserVerification()
	webAuthnRoamingData["override_relying_party"] = webAuthnSettings.GetOverrideRelyingParty()
	webAuthnRoamingData["relying_party_identifier"] = webAuthnSettings.GetRelyingPartyIdentifier()

	return []interface{}{webAuthnRoamingData}, nil
}

func flattenWebAuthnPlatform(ctx context.Context, enabled bool, api *management.Management) ([]interface{}, error) {
	webAuthnPlatformData := make(map[string]interface{})
	webAuthnPlatformData["enabled"] = enabled

	if !enabled {
		return []interface{}{webAuthnPlatformData}, nil
	}

	webAuthnSettings, err := api.Guardian.MultiFactor.WebAuthnPlatform.Read(ctx)
	if err != nil {
		return nil, err
	}

	webAuthnPlatformData["override_relying_party"] = webAuthnSettings.GetOverrideRelyingParty()
	webAuthnPlatformData["relying_party_identifier"] = webAuthnSettings.GetRelyingPartyIdentifier()

	return []interface{}{webAuthnPlatformData}, nil
}

func flattenDUO(ctx context.Context, enabled bool, api *management.Management) ([]interface{}, error) {
	duoData := make(map[string]interface{})
	duoData["enabled"] = enabled

	if !enabled {
		return []interface{}{duoData}, nil
	}

	duoSettings, err := api.Guardian.MultiFactor.DUO.Read(ctx)
	if err != nil {
		return nil, err
	}

	duoData["integration_key"] = duoSettings.GetIntegrationKey()
	duoData["secret_key"] = duoSettings.GetSecretKey()
	duoData["hostname"] = duoSettings.GetHostname()

	return []interface{}{duoData}, nil
}

func flattenPush(ctx context.Context, data *schema.ResourceData, enabled bool, api *management.Management) ([]interface{}, error) {
	pushData := make(map[string]interface{})
	pushData["enabled"] = enabled

	if !enabled {
		return []interface{}{pushData}, nil
	}

	pushProvider, err := api.Guardian.MultiFactor.Push.Provider(ctx)
	if err != nil {
		return nil, err
	}
	pushData["provider"] = pushProvider.GetProvider()

	customApp, err := api.Guardian.MultiFactor.Push.CustomApp(ctx)
	if err != nil {
		return nil, err
	}

	pushData["custom_app"] = []interface{}{
		map[string]interface{}{
			"app_name":        customApp.GetAppName(),
			"apple_app_link":  customApp.GetAppleAppLink(),
			"google_app_link": customApp.GetGoogleAppLink(),
		},
	}

	directAPNS, err := api.Guardian.MultiFactor.Push.DirectAPNS(ctx)
	if err != nil {
		return nil, err
	}

	pushData["direct_apns"] = []interface{}{
		map[string]interface{}{
			"sandbox":   directAPNS.GetSandbox(),
			"p12":       data.Get("push.0.direct_apns.0.p12"), // Does not get read back.
			"bundle_id": directAPNS.GetBundleID(),
			"enabled":   directAPNS.GetEnabled(),
		},
	}

	pushData["direct_fcm"] = []interface{}{
		map[string]interface{}{
			"server_key": data.Get("push.0.direct_fcm.0.server_key"), // Does not get read back.
		},
	}

	if pushProvider.GetProvider() == "sns" {
		amazonSNS, err := api.Guardian.MultiFactor.Push.AmazonSNS(ctx)
		if err != nil {
			return nil, err
		}

		pushData["amazon_sns"] = []interface{}{
			map[string]interface{}{
				"aws_access_key_id":                 amazonSNS.GetAccessKeyID(),
				"aws_region":                        amazonSNS.GetRegion(),
				"aws_secret_access_key":             data.Get("push.0.amazon_sns.0.aws_secret_access_key"), // Does not get read back.
				"sns_apns_platform_application_arn": amazonSNS.GetAPNSPlatformApplicationARN(),
				"sns_gcm_platform_application_arn":  amazonSNS.GetGCMPlatformApplicationARN(),
			},
		}
	}

	return []interface{}{pushData}, nil
}
