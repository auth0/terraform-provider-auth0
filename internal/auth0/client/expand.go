package client

import (
	"fmt"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandClient(data *schema.ResourceData) (*management.Client, error) {
	config := data.GetRawConfig()

	client := &management.Client{
		Name:                               value.String(config.GetAttr("name")),
		Description:                        value.String(config.GetAttr("description")),
		AppType:                            value.String(config.GetAttr("app_type")),
		LogoURI:                            value.String(config.GetAttr("logo_uri")),
		IsFirstParty:                       value.Bool(config.GetAttr("is_first_party")),
		OIDCConformant:                     value.Bool(config.GetAttr("oidc_conformant")),
		ClientAliases:                      value.Strings(config.GetAttr("client_aliases")),
		Callbacks:                          value.Strings(config.GetAttr("callbacks")),
		AllowedLogoutURLs:                  value.Strings(config.GetAttr("allowed_logout_urls")),
		AllowedOrigins:                     value.Strings(config.GetAttr("allowed_origins")),
		AllowedClients:                     value.Strings(config.GetAttr("allowed_clients")),
		GrantTypes:                         value.Strings(config.GetAttr("grant_types")),
		AsyncApprovalNotificationChannels:  value.Strings(config.GetAttr("async_approval_notification_channels")),
		OrganizationUsage:                  value.String(config.GetAttr("organization_usage")),
		OrganizationRequireBehavior:        value.String(config.GetAttr("organization_require_behavior")),
		OrganizationDiscoveryMethods:       value.Strings(config.GetAttr("organization_discovery_methods")),
		WebOrigins:                         value.Strings(config.GetAttr("web_origins")),
		RequirePushedAuthorizationRequests: value.Bool(config.GetAttr("require_pushed_authorization_requests")),
		SSO:                                value.Bool(config.GetAttr("sso")),
		SSODisabled:                        value.Bool(config.GetAttr("sso_disabled")),
		CrossOriginAuth:                    value.Bool(config.GetAttr("cross_origin_auth")),
		CrossOriginLocation:                value.String(config.GetAttr("cross_origin_loc")),
		CustomLoginPageOn:                  value.Bool(config.GetAttr("custom_login_page_on")),
		CustomLoginPage:                    value.String(config.GetAttr("custom_login_page")),
		FormTemplate:                       value.String(config.GetAttr("form_template")),
		InitiateLoginURI:                   value.String(config.GetAttr("initiate_login_uri")),
		EncryptionKey:                      value.MapOfStrings(config.GetAttr("encryption_key")),
		IsTokenEndpointIPHeaderTrusted:     value.Bool(config.GetAttr("is_token_endpoint_ip_header_trusted")),
		// TODO(major): Replace OIDCBackchannelLogout with OIDCLogout when releasing v2.
		//nolint:staticcheck // SA1019 — OIDCBackchannelLogout is deprecated, retained for backward compatibility.
		OIDCBackchannelLogout:    expandOIDCBackchannelLogout(data),
		OIDCLogout:               expandOIDCLogout(data),
		ClientMetadata:           expandClientMetadata(data),
		RefreshToken:             expandClientRefreshToken(data),
		JWTConfiguration:         expandClientJWTConfiguration(data),
		Addons:                   expandClientAddons(data),
		NativeSocialLogin:        expandClientNativeSocialLogin(data),
		Mobile:                   expandClientMobile(data),
		DefaultOrganization:      expandDefaultOrganization(data),
		TokenExchange:            expandTokenExchange(data),
		RequireProofOfPossession: value.Bool(config.GetAttr("require_proof_of_possession")),
		SessionTransfer:          expandSessionTransfer(data),
		ComplianceLevel:          value.String(config.GetAttr("compliance_level")),
		TokenQuota:               commons.ExpandTokenQuota(config.GetAttr("token_quota")),
		SkipNonVerifiableCallbackURIConfirmationPrompt: value.BoolPtr(data.Get("skip_non_verifiable_callback_uri_confirmation_prompt")),
		ExpressConfiguration:                           expandExpressConfiguration(data),
	}

	if data.IsNewResource() && client.IsTokenEndpointIPHeaderTrusted != nil {
		client.TokenEndpointAuthMethod = auth0.String("client_secret_post")
	}

	if data.IsNewResource() {
		switch client.GetAppType() {
		case "native", "spa":
			client.TokenEndpointAuthMethod = auth0.String("none")
		case "regular_web", "non_interactive":
			client.TokenEndpointAuthMethod = auth0.String("client_secret_post")
		}
	}

	if data.IsNewResource() && client.ResourceServerIdentifier != nil {
		client.ResourceServerIdentifier = value.String(config.GetAttr("resource_server_identifier"))
	}

	defaultConfig := data.GetRawConfig().GetAttr("default_organization")

	for _, item := range defaultConfig.AsValueSlice() {
		disable := item.GetAttr("disable")
		organizationID := item.GetAttr("organization_id")
		flows := item.GetAttr("flows")

		if !disable.IsNull() && disable.True() {
			if (!flows.IsNull() && flows.LengthInt() > 0) || (!organizationID.IsNull() && organizationID.AsString() != "") {
				return nil, fmt.Errorf("cannot set both disable and either flows/organization_id")
			}
		}
	}

	return client, nil
}

func expandDefaultOrganization(data *schema.ResourceData) *management.ClientDefaultOrganization {
	if !data.IsNewResource() && !data.HasChange("default_organization") {
		return nil
	}
	var defaultOrg management.ClientDefaultOrganization

	config := data.GetRawConfig().GetAttr("default_organization")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		if !disable.IsNull() && disable.True() {
			stop = true
		} else {
			defaultOrg.Flows = value.Strings(cfg.GetAttr("flows"))
			defaultOrg.OrganizationID = value.String(cfg.GetAttr("organization_id"))
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return nil
	}
	if defaultOrg == (management.ClientDefaultOrganization{}) {
		return nil
	}

	return &defaultOrg
}

func expandTokenExchange(data *schema.ResourceData) *management.ClientTokenExchange {
	if !data.IsNewResource() && !data.HasChange("token_exchange") {
		return nil
	}
	var tokenExchange management.ClientTokenExchange

	config := data.GetRawConfig().GetAttr("token_exchange")
	if config.IsNull() || config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		tokenExchange.AllowAnyProfileOfType = value.Strings(cfg.GetAttr("allow_any_profile_of_type"))
		return stop
	}) {
		return nil
	}
	if tokenExchange == (management.ClientTokenExchange{}) {
		return nil
	}

	return &tokenExchange
}

// TODO(major): Replace OIDCBackchannelLogout with OIDCLogout when releasing v2.
//
//nolint:staticcheck // SA1019 — OIDCBackchannelLogout is deprecated, retained for backward compatibility.
func expandOIDCBackchannelLogout(data *schema.ResourceData) *management.OIDCBackchannelLogout {
	raw := data.GetRawConfig().GetAttr("oidc_backchannel_logout_urls")

	logoutUrls := value.Strings(raw)

	if logoutUrls == nil {
		return nil
	}

	// TODO(major): Replace OIDCBackchannelLogout with OIDCLogout when releasing v2.
	//nolint:staticcheck // SA1019 — OIDCBackchannelLogout is deprecated, retained for backward compatibility.
	return &management.OIDCBackchannelLogout{
		BackChannelLogoutURLs: logoutUrls,
	}
}

func expandOIDCLogout(data *schema.ResourceData) *management.OIDCLogout {
	oidcLogoutConfig := data.GetRawConfig().GetAttr("oidc_logout")
	if oidcLogoutConfig.IsNull() {
		return nil
	}

	var oidcLogout management.OIDCLogout

	oidcLogoutConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		oidcLogout.BackChannelLogoutURLs = value.Strings(config.GetAttr("backchannel_logout_urls"))
		oidcLogout.BackChannelLogoutInitiators = expandBackChannelLogoutInitiators(config.GetAttr("backchannel_logout_initiators"))
		oidcLogout.BackChannelLogoutSessionMetadata = expandBackChannelLogoutSessionMetadata(config.GetAttr("backchannel_logout_session_metadata"))
		return stop
	})

	if oidcLogout == (management.OIDCLogout{}) {
		return nil
	}

	return &oidcLogout
}

func expandBackChannelLogoutInitiators(config cty.Value) *management.BackChannelLogoutInitiators {
	if config.IsNull() {
		return nil
	}

	var initiators management.BackChannelLogoutInitiators

	config.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		initiators.Mode = value.String(config.GetAttr("mode"))
		initiators.SelectedInitiators = value.Strings(config.GetAttr("selected_initiators"))
		return stop
	})

	if initiators == (management.BackChannelLogoutInitiators{}) {
		return nil
	}

	return &initiators
}

func expandBackChannelLogoutSessionMetadata(config cty.Value) *management.BackChannelLogoutSessionMetadata {
	if config.IsNull() {
		return nil
	}

	var metadata management.BackChannelLogoutSessionMetadata

	config.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		metadata.Include = value.Bool(config.GetAttr("include"))
		return stop
	})

	if metadata == (management.BackChannelLogoutSessionMetadata{}) {
		return nil
	}

	return &metadata
}

func expandClientRefreshToken(data *schema.ResourceData) *management.ClientRefreshToken {
	refreshTokenConfig := data.GetRawConfig().GetAttr("refresh_token")
	if refreshTokenConfig.IsNull() {
		return nil
	}

	var refreshToken management.ClientRefreshToken

	refreshTokenConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		refreshToken.RotationType = value.String(config.GetAttr("rotation_type"))
		refreshToken.ExpirationType = value.String(config.GetAttr("expiration_type"))
		refreshToken.Leeway = value.Int(config.GetAttr("leeway"))
		refreshToken.TokenLifetime = value.Int(config.GetAttr("token_lifetime"))
		refreshToken.InfiniteTokenLifetime = value.Bool(config.GetAttr("infinite_token_lifetime"))
		refreshToken.InfiniteIdleTokenLifetime = value.Bool(config.GetAttr("infinite_idle_token_lifetime"))
		refreshToken.IdleTokenLifetime = value.Int(config.GetAttr("idle_token_lifetime"))
		refreshToken.Policies = expandRefreshTokenPolicies(config.GetAttr("policies"))
		return stop
	})

	if refreshToken == (management.ClientRefreshToken{}) {
		return nil
	}

	return &refreshToken
}

func expandRefreshTokenPolicies(policies cty.Value) *[]management.ClientRefreshTokenPolicy {
	clientRefreshTokenPolicy := make([]management.ClientRefreshTokenPolicy, 0)

	policies.ForEachElement(func(_ cty.Value, dep cty.Value) (stop bool) {
		clientRefreshTokenPolicy = append(clientRefreshTokenPolicy, management.ClientRefreshTokenPolicy{
			Audience: value.String(dep.GetAttr("audience")),
			Scope:    value.Strings(dep.GetAttr("scope")),
		})
		return stop
	})

	if len(clientRefreshTokenPolicy) == 0 {
		return nil
	}

	return &clientRefreshTokenPolicy
}

func expandClientJWTConfiguration(data *schema.ResourceData) *management.ClientJWTConfiguration {
	jwtConfig := data.GetRawConfig().GetAttr("jwt_configuration")
	if jwtConfig.IsNull() {
		return nil
	}

	var jwt management.ClientJWTConfiguration

	jwtConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		jwt.LifetimeInSeconds = value.Int(config.GetAttr("lifetime_in_seconds"))
		jwt.Algorithm = value.String(config.GetAttr("alg"))
		jwt.Scopes = value.MapOfStrings(config.GetAttr("scopes"))

		if data.IsNewResource() {
			jwt.SecretEncoded = value.Bool(config.GetAttr("secret_encoded"))
		}

		return stop
	})

	if jwt == (management.ClientJWTConfiguration{}) {
		return nil
	}

	return &jwt
}

func expandClientNativeSocialLogin(data *schema.ResourceData) *management.ClientNativeSocialLogin {
	nativeSocialLoginConfig := data.GetRawConfig().GetAttr("native_social_login")
	if nativeSocialLoginConfig.IsNull() {
		return nil
	}

	var nativeSocialLogin management.ClientNativeSocialLogin

	nativeSocialLoginConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		nativeSocialLogin.Apple = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("apple"))
		nativeSocialLogin.Facebook = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("facebook"))
		nativeSocialLogin.Google = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("google"))
		return stop
	})

	if nativeSocialLogin == (management.ClientNativeSocialLogin{}) {
		return nil
	}

	return &nativeSocialLogin
}

func expandClientNativeSocialLoginSupportEnabled(config cty.Value) *management.ClientNativeSocialLoginSupportEnabled {
	if config.IsNull() {
		return nil
	}

	var support management.ClientNativeSocialLoginSupportEnabled

	config.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		support.Enabled = value.Bool(config.GetAttr("enabled"))
		return stop
	})

	if support == (management.ClientNativeSocialLoginSupportEnabled{}) {
		return nil
	}

	return &support
}

func expandClientMobile(data *schema.ResourceData) *management.ClientMobile {
	mobileConfig := data.GetRawConfig().GetAttr("mobile")
	if mobileConfig.IsNull() {
		return nil
	}

	var mobile management.ClientMobile

	mobileConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		mobile.Android = expandClientMobileAndroid(config.GetAttr("android"))
		mobile.IOS = expandClientMobileIOS(config.GetAttr("ios"))
		return stop
	})

	if mobile == (management.ClientMobile{}) {
		return nil
	}

	return &mobile
}

func expandClientMobileAndroid(androidConfig cty.Value) *management.ClientMobileAndroid {
	if androidConfig.IsNull() {
		return nil
	}

	var android management.ClientMobileAndroid

	androidConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		android.AppPackageName = value.String(config.GetAttr("app_package_name"))
		android.KeyHashes = value.Strings(config.GetAttr("sha256_cert_fingerprints"))
		return stop
	})

	if android == (management.ClientMobileAndroid{}) {
		return nil
	}

	return &android
}

func expandClientMobileIOS(iosConfig cty.Value) *management.ClientMobileIOS {
	if iosConfig.IsNull() {
		return nil
	}

	var ios management.ClientMobileIOS

	iosConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		ios.TeamID = value.String(config.GetAttr("team_id"))
		ios.AppID = value.String(config.GetAttr("app_bundle_identifier"))
		return stop
	})

	if ios == (management.ClientMobileIOS{}) {
		return nil
	}

	return &ios
}

func expandClientMetadata(data *schema.ResourceData) *map[string]interface{} {
	if !data.HasChange("client_metadata") {
		return nil
	}

	oldMetadata, newMetadata := data.GetChange("client_metadata")
	oldMetadataMap := oldMetadata.(map[string]interface{})
	newMetadataMap := newMetadata.(map[string]interface{})

	for key := range oldMetadataMap {
		if _, ok := newMetadataMap[key]; !ok {
			newMetadataMap[key] = nil
		}
	}

	return &newMetadataMap
}

func expandClientAddons(data *schema.ResourceData) *management.ClientAddons {
	if !data.HasChange("addons") {
		return nil
	}

	var addons management.ClientAddons

	data.GetRawConfig().GetAttr("addons").ForEachElement(func(_ cty.Value, addonsCfg cty.Value) (stop bool) {
		addons.AWS = expandClientAddonAWS(addonsCfg.GetAttr("aws"))
		addons.AzureBlob = expandClientAddonAzureBlob(addonsCfg.GetAttr("azure_blob"))
		addons.AzureSB = expandClientAddonAzureSB(addonsCfg.GetAttr("azure_sb"))
		addons.RMS = expandClientAddonRMS(addonsCfg.GetAttr("rms"))
		addons.MSCRM = expandClientAddonMSCRM(addonsCfg.GetAttr("mscrm"))
		addons.Slack = expandClientAddonSlack(addonsCfg.GetAttr("slack"))
		addons.Sentry = expandClientAddonSentry(addonsCfg.GetAttr("sentry"))
		addons.EchoSign = expandClientAddonEchoSign(addonsCfg.GetAttr("echosign"))
		addons.Egnyte = expandClientAddonEgnyte(addonsCfg.GetAttr("egnyte"))
		addons.Firebase = expandClientAddonFirebase(addonsCfg.GetAttr("firebase"))
		addons.NewRelic = expandClientAddonNewRelic(addonsCfg.GetAttr("newrelic"))
		addons.Office365 = expandClientAddonOffice365(addonsCfg.GetAttr("office365"))
		addons.Salesforce = expandClientAddonSalesforce(addonsCfg.GetAttr("salesforce"))
		addons.SalesforceAPI = expandClientAddonSalesforceAPI(addonsCfg.GetAttr("salesforce_api"))
		addons.SalesforceSandboxAPI = expandClientAddonSalesforceSandboxAPI(addonsCfg.GetAttr("salesforce_sandbox_api"))
		addons.Layer = expandClientAddonLayer(addonsCfg.GetAttr("layer"))
		addons.SAPAPI = expandClientAddonSAPAPI(addonsCfg.GetAttr("sap_api"))
		addons.SharePoint = expandClientAddonSharepoint(addonsCfg.GetAttr("sharepoint"))
		addons.SpringCM = expandClientAddonSpringCM(addonsCfg.GetAttr("springcm"))
		addons.WAMS = expandClientAddonWAMS(addonsCfg.GetAttr("wams"))
		addons.Zendesk = expandClientAddonZendesk(addonsCfg.GetAttr("zendesk"))
		addons.Zoom = expandClientAddonZoom(addonsCfg.GetAttr("zoom"))
		addons.SSOIntegration = expandClientAddonSSOIntegration(addonsCfg.GetAttr("sso_integration"))
		addons.SAML2 = expandClientAddonSAMLP(addonsCfg.GetAttr("samlp"))

		if addonsCfg.GetAttr("box").LengthInt() == 1 {
			addons.Box = &management.BoxClientAddon{}
		}

		if addonsCfg.GetAttr("cloudbees").LengthInt() == 1 {
			addons.CloudBees = &management.CloudBeesClientAddon{}
		}

		if addonsCfg.GetAttr("concur").LengthInt() == 1 {
			addons.Concur = &management.ConcurClientAddon{}
		}

		if addonsCfg.GetAttr("dropbox").LengthInt() == 1 {
			addons.Dropbox = &management.DropboxClientAddon{}
		}

		if addonsCfg.GetAttr("wsfed").LengthInt() == 1 {
			addons.WSFED = &management.WSFEDClientAddon{}
		}

		return stop
	})

	return &addons
}

func expandClientAddonAWS(awsCfg cty.Value) *management.AWSClientAddon {
	var awsAddon management.AWSClientAddon

	awsCfg.ForEachElement(func(_ cty.Value, awsCfg cty.Value) (stop bool) {
		awsAddon = management.AWSClientAddon{
			Principal:         value.String(awsCfg.GetAttr("principal")),
			Role:              value.String(awsCfg.GetAttr("role")),
			LifetimeInSeconds: value.Int(awsCfg.GetAttr("lifetime_in_seconds")),
		}

		return stop
	})

	if awsAddon == (management.AWSClientAddon{}) {
		return nil
	}

	return &awsAddon
}

func expandClientAddonAzureBlob(azureCfg cty.Value) *management.AzureBlobClientAddon {
	var azureAddon management.AzureBlobClientAddon

	azureCfg.ForEachElement(func(_ cty.Value, azureCfg cty.Value) (stop bool) {
		azureAddon = management.AzureBlobClientAddon{
			AccountName:      value.String(azureCfg.GetAttr("account_name")),
			StorageAccessKey: value.String(azureCfg.GetAttr("storage_access_key")),
			ContainerName:    value.String(azureCfg.GetAttr("container_name")),
			BlobName:         value.String(azureCfg.GetAttr("blob_name")),
			Expiration:       value.Int(azureCfg.GetAttr("expiration")),
			SignedIdentifier: value.String(azureCfg.GetAttr("signed_identifier")),
			BlobRead:         value.Bool(azureCfg.GetAttr("blob_read")),
			BlobWrite:        value.Bool(azureCfg.GetAttr("blob_write")),
			BlobDelete:       value.Bool(azureCfg.GetAttr("blob_delete")),
			ContainerRead:    value.Bool(azureCfg.GetAttr("container_read")),
			ContainerWrite:   value.Bool(azureCfg.GetAttr("container_write")),
			ContainerDelete:  value.Bool(azureCfg.GetAttr("container_delete")),
			ContainerList:    value.Bool(azureCfg.GetAttr("container_list")),
		}

		return stop
	})

	if azureAddon == (management.AzureBlobClientAddon{}) {
		return nil
	}

	return &azureAddon
}

func expandClientAddonAzureSB(azureCfg cty.Value) *management.AzureSBClientAddon {
	var azureAddon management.AzureSBClientAddon

	azureCfg.ForEachElement(func(_ cty.Value, azureCfg cty.Value) (stop bool) {
		azureAddon = management.AzureSBClientAddon{
			Namespace:  value.String(azureCfg.GetAttr("namespace")),
			SASKeyName: value.String(azureCfg.GetAttr("sas_key_name")),
			SASKey:     value.String(azureCfg.GetAttr("sas_key")),
			EntityPath: value.String(azureCfg.GetAttr("entity_path")),
			Expiration: value.Int(azureCfg.GetAttr("expiration")),
		}

		return stop
	})

	if azureAddon == (management.AzureSBClientAddon{}) {
		return nil
	}

	return &azureAddon
}

func expandClientAddonRMS(rmsCfg cty.Value) *management.RMSClientAddon {
	var rmsAddon management.RMSClientAddon

	rmsCfg.ForEachElement(func(_ cty.Value, rmsCfg cty.Value) (stop bool) {
		rmsAddon = management.RMSClientAddon{
			URL: value.String(rmsCfg.GetAttr("url")),
		}

		return stop
	})

	if rmsAddon == (management.RMSClientAddon{}) {
		return nil
	}

	return &rmsAddon
}

func expandClientAddonMSCRM(mscrmCfg cty.Value) *management.MSCRMClientAddon {
	var mscrmAddon management.MSCRMClientAddon

	mscrmCfg.ForEachElement(func(_ cty.Value, mscrmCfg cty.Value) (stop bool) {
		mscrmAddon = management.MSCRMClientAddon{
			URL: value.String(mscrmCfg.GetAttr("url")),
		}

		return stop
	})

	if mscrmAddon == (management.MSCRMClientAddon{}) {
		return nil
	}

	return &mscrmAddon
}

func expandClientAddonSlack(slackCfg cty.Value) *management.SlackClientAddon {
	var slackAddon management.SlackClientAddon

	slackCfg.ForEachElement(func(_ cty.Value, slackCfg cty.Value) (stop bool) {
		slackAddon = management.SlackClientAddon{
			Team: value.String(slackCfg.GetAttr("team")),
		}

		return stop
	})

	if slackAddon == (management.SlackClientAddon{}) {
		return nil
	}

	return &slackAddon
}

func expandClientAddonSentry(sentryCfg cty.Value) *management.SentryClientAddon {
	var sentryAddon management.SentryClientAddon

	sentryCfg.ForEachElement(func(_ cty.Value, sentryCfg cty.Value) (stop bool) {
		sentryAddon = management.SentryClientAddon{
			OrgSlug: value.String(sentryCfg.GetAttr("org_slug")),
			BaseURL: value.String(sentryCfg.GetAttr("base_url")),
		}

		return stop
	})

	if sentryAddon == (management.SentryClientAddon{}) {
		return nil
	}

	return &sentryAddon
}

func expandClientAddonEchoSign(echoSignCfg cty.Value) *management.EchoSignClientAddon {
	var echoSignAddon management.EchoSignClientAddon

	echoSignCfg.ForEachElement(func(_ cty.Value, echoSignCfg cty.Value) (stop bool) {
		echoSignAddon = management.EchoSignClientAddon{
			Domain: value.String(echoSignCfg.GetAttr("domain")),
		}

		return stop
	})

	if echoSignAddon == (management.EchoSignClientAddon{}) {
		return nil
	}

	return &echoSignAddon
}

func expandClientAddonEgnyte(egnyteCfg cty.Value) *management.EgnyteClientAddon {
	var egnyteAddon management.EgnyteClientAddon

	egnyteCfg.ForEachElement(func(_ cty.Value, egnyteCfg cty.Value) (stop bool) {
		egnyteAddon = management.EgnyteClientAddon{
			Domain: value.String(egnyteCfg.GetAttr("domain")),
		}

		return stop
	})

	if egnyteAddon == (management.EgnyteClientAddon{}) {
		return nil
	}

	return &egnyteAddon
}

func expandClientAddonFirebase(firebaseCfg cty.Value) *management.FirebaseClientAddon {
	var firebaseAddon management.FirebaseClientAddon

	firebaseCfg.ForEachElement(func(_ cty.Value, firebaseCfg cty.Value) (stop bool) {
		firebaseAddon = management.FirebaseClientAddon{
			Secret:            value.String(firebaseCfg.GetAttr("secret")),
			PrivateKeyID:      value.String(firebaseCfg.GetAttr("private_key_id")),
			PrivateKey:        value.String(firebaseCfg.GetAttr("private_key")),
			ClientEmail:       value.String(firebaseCfg.GetAttr("client_email")),
			LifetimeInSeconds: value.Int(firebaseCfg.GetAttr("lifetime_in_seconds")),
		}

		return stop
	})

	if firebaseAddon == (management.FirebaseClientAddon{}) {
		return nil
	}

	return &firebaseAddon
}

func expandClientAddonNewRelic(newRelicCfg cty.Value) *management.NewRelicClientAddon {
	var newRelicAddon management.NewRelicClientAddon

	newRelicCfg.ForEachElement(func(_ cty.Value, newRelicCfg cty.Value) (stop bool) {
		newRelicAddon = management.NewRelicClientAddon{
			Account: value.String(newRelicCfg.GetAttr("account")),
		}

		return stop
	})

	if newRelicAddon == (management.NewRelicClientAddon{}) {
		return nil
	}

	return &newRelicAddon
}

func expandClientAddonOffice365(office365Cfg cty.Value) *management.Office365ClientAddon {
	var office365Addon management.Office365ClientAddon

	office365Cfg.ForEachElement(func(_ cty.Value, office365Cfg cty.Value) (stop bool) {
		office365Addon = management.Office365ClientAddon{
			Domain:     value.String(office365Cfg.GetAttr("domain")),
			Connection: value.String(office365Cfg.GetAttr("connection")),
		}

		return stop
	})

	if office365Addon == (management.Office365ClientAddon{}) {
		return nil
	}

	return &office365Addon
}

func expandClientAddonSalesforce(salesforceCfg cty.Value) *management.SalesforceClientAddon {
	var salesforceAddon management.SalesforceClientAddon

	salesforceCfg.ForEachElement(func(_ cty.Value, salesforceCfg cty.Value) (stop bool) {
		salesforceAddon = management.SalesforceClientAddon{
			EntityID: value.String(salesforceCfg.GetAttr("entity_id")),
		}

		return stop
	})

	if salesforceAddon == (management.SalesforceClientAddon{}) {
		return nil
	}

	return &salesforceAddon
}

func expandClientAddonSalesforceAPI(salesforceCfg cty.Value) *management.SalesforceAPIClientAddon {
	var salesforceAddon management.SalesforceAPIClientAddon

	salesforceCfg.ForEachElement(func(_ cty.Value, salesforceCfg cty.Value) (stop bool) {
		salesforceAddon = management.SalesforceAPIClientAddon{
			ClientID:            value.String(salesforceCfg.GetAttr("client_id")),
			Principal:           value.String(salesforceCfg.GetAttr("principal")),
			CommunityName:       value.String(salesforceCfg.GetAttr("community_name")),
			CommunityURLSection: value.String(salesforceCfg.GetAttr("community_url_section")),
		}

		return stop
	})

	if salesforceAddon == (management.SalesforceAPIClientAddon{}) {
		return nil
	}

	return &salesforceAddon
}

func expandClientAddonSalesforceSandboxAPI(salesforceCfg cty.Value) *management.SalesforceSandboxAPIClientAddon {
	var salesforceAddon management.SalesforceSandboxAPIClientAddon

	salesforceCfg.ForEachElement(func(_ cty.Value, salesforceCfg cty.Value) (stop bool) {
		salesforceAddon = management.SalesforceSandboxAPIClientAddon{
			ClientID:            value.String(salesforceCfg.GetAttr("client_id")),
			Principal:           value.String(salesforceCfg.GetAttr("principal")),
			CommunityName:       value.String(salesforceCfg.GetAttr("community_name")),
			CommunityURLSection: value.String(salesforceCfg.GetAttr("community_url_section")),
		}

		return stop
	})

	if salesforceAddon == (management.SalesforceSandboxAPIClientAddon{}) {
		return nil
	}

	return &salesforceAddon
}

func expandClientAddonLayer(layerCfg cty.Value) *management.LayerClientAddon {
	var layerAddon management.LayerClientAddon

	layerCfg.ForEachElement(func(_ cty.Value, layerCfg cty.Value) (stop bool) {
		layerAddon = management.LayerClientAddon{
			ProviderID: value.String(layerCfg.GetAttr("provider_id")),
			KeyID:      value.String(layerCfg.GetAttr("key_id")),
			PrivateKey: value.String(layerCfg.GetAttr("private_key")),
			Principal:  value.String(layerCfg.GetAttr("principal")),
			Expiration: value.Int(layerCfg.GetAttr("expiration")),
		}

		return stop
	})

	if layerAddon == (management.LayerClientAddon{}) {
		return nil
	}

	return &layerAddon
}

func expandClientAddonSAPAPI(sapAPICfg cty.Value) *management.SAPAPIClientAddon {
	var sapAPIAddon management.SAPAPIClientAddon

	sapAPICfg.ForEachElement(func(_ cty.Value, sapAPICfg cty.Value) (stop bool) {
		sapAPIAddon = management.SAPAPIClientAddon{
			ClientID:             value.String(sapAPICfg.GetAttr("client_id")),
			UsernameAttribute:    value.String(sapAPICfg.GetAttr("username_attribute")),
			TokenEndpointURL:     value.String(sapAPICfg.GetAttr("token_endpoint_url")),
			Scope:                value.String(sapAPICfg.GetAttr("scope")),
			ServicePassword:      value.String(sapAPICfg.GetAttr("service_password")),
			NameIdentifierFormat: value.String(sapAPICfg.GetAttr("name_identifier_format")),
		}

		return stop
	})

	if sapAPIAddon == (management.SAPAPIClientAddon{}) {
		return nil
	}

	return &sapAPIAddon
}

func expandClientAddonSharepoint(sharepointCfg cty.Value) *management.SharePointClientAddon {
	var sharepointAddon management.SharePointClientAddon

	sharepointCfg.ForEachElement(func(_ cty.Value, sharepointCfg cty.Value) (stop bool) {
		sharepointAddon = management.SharePointClientAddon{
			URL:         value.String(sharepointCfg.GetAttr("url")),
			ExternalURL: value.Strings(sharepointCfg.GetAttr("external_url")),
		}

		return stop
	})

	if sharepointAddon == (management.SharePointClientAddon{}) {
		return nil
	}

	return &sharepointAddon
}

func expandClientAddonSpringCM(springCMCfg cty.Value) *management.SpringCMClientAddon {
	var springCMAddon management.SpringCMClientAddon

	springCMCfg.ForEachElement(func(_ cty.Value, springCMCfg cty.Value) (stop bool) {
		springCMAddon = management.SpringCMClientAddon{
			ACSURL: value.String(springCMCfg.GetAttr("acs_url")),
		}

		return stop
	})

	if springCMAddon == (management.SpringCMClientAddon{}) {
		return nil
	}

	return &springCMAddon
}

func expandClientAddonWAMS(wamsCfg cty.Value) *management.WAMSClientAddon {
	var wamsAddon management.WAMSClientAddon

	wamsCfg.ForEachElement(func(_ cty.Value, wamsCfg cty.Value) (stop bool) {
		wamsAddon = management.WAMSClientAddon{
			Masterkey: value.String(wamsCfg.GetAttr("master_key")),
		}

		return stop
	})

	if wamsAddon == (management.WAMSClientAddon{}) {
		return nil
	}

	return &wamsAddon
}

func expandClientAddonZendesk(zendeskCfg cty.Value) *management.ZendeskClientAddon {
	var zendeskAddon management.ZendeskClientAddon

	zendeskCfg.ForEachElement(func(_ cty.Value, zendeskCfg cty.Value) (stop bool) {
		zendeskAddon = management.ZendeskClientAddon{
			AccountName: value.String(zendeskCfg.GetAttr("account_name")),
		}

		return stop
	})

	if zendeskAddon == (management.ZendeskClientAddon{}) {
		return nil
	}

	return &zendeskAddon
}

func expandClientAddonZoom(zoomCfg cty.Value) *management.ZoomClientAddon {
	var zoomAddon management.ZoomClientAddon

	zoomCfg.ForEachElement(func(_ cty.Value, zoomCfg cty.Value) (stop bool) {
		zoomAddon = management.ZoomClientAddon{
			Account: value.String(zoomCfg.GetAttr("account")),
		}

		return stop
	})

	if zoomAddon == (management.ZoomClientAddon{}) {
		return nil
	}

	return &zoomAddon
}

func expandClientAddonSSOIntegration(ssoCfg cty.Value) *management.SSOIntegrationClientAddon {
	var ssoAddon management.SSOIntegrationClientAddon

	ssoCfg.ForEachElement(func(_ cty.Value, ssoCfg cty.Value) (stop bool) {
		ssoAddon = management.SSOIntegrationClientAddon{
			Name:    value.String(ssoCfg.GetAttr("name")),
			Version: value.String(ssoCfg.GetAttr("version")),
		}

		return stop
	})

	if ssoAddon == (management.SSOIntegrationClientAddon{}) {
		return nil
	}

	return &ssoAddon
}

func expandClientAddonSAMLP(samlpCfg cty.Value) *management.SAML2ClientAddon {
	var samlpAddon management.SAML2ClientAddon

	samlpCfg.ForEachElement(func(_ cty.Value, samlpCfg cty.Value) (stop bool) {
		samlpAddon = management.SAML2ClientAddon{
			Mappings:                       value.MapOfStrings(samlpCfg.GetAttr("mappings")),
			Audience:                       value.String(samlpCfg.GetAttr("audience")),
			Recipient:                      value.String(samlpCfg.GetAttr("recipient")),
			CreateUPNClaim:                 value.Bool(samlpCfg.GetAttr("create_upn_claim")),
			MapUnknownClaimsAsIs:           value.Bool(samlpCfg.GetAttr("map_unknown_claims_as_is")),
			PassthroughClaimsWithNoMapping: value.Bool(samlpCfg.GetAttr("passthrough_claims_with_no_mapping")),
			MapIdentities:                  value.Bool(samlpCfg.GetAttr("map_identities")),
			SignatureAlgorithm:             value.String(samlpCfg.GetAttr("signature_algorithm")),
			DigestAlgorithm:                value.String(samlpCfg.GetAttr("digest_algorithm")),
			Issuer:                         value.String(samlpCfg.GetAttr("issuer")),
			Destination:                    value.String(samlpCfg.GetAttr("destination")),
			LifetimeInSeconds:              value.Int(samlpCfg.GetAttr("lifetime_in_seconds")),
			SignResponse:                   value.Bool(samlpCfg.GetAttr("sign_response")),
			NameIdentifierFormat:           value.String(samlpCfg.GetAttr("name_identifier_format")),
			NameIdentifierProbes:           value.Strings(samlpCfg.GetAttr("name_identifier_probes")),
			AuthnContextClassRef:           value.String(samlpCfg.GetAttr("authn_context_class_ref")),
			TypedAttributes:                value.Bool(samlpCfg.GetAttr("typed_attributes")),
			IncludeAttributeNameFormat:     value.Bool(samlpCfg.GetAttr("include_attribute_name_format")),
			Binding:                        value.String(samlpCfg.GetAttr("binding")),
			SigningCert:                    value.String(samlpCfg.GetAttr("signing_cert")),
		}

		flexibleMappings, err := value.MapFromJSON(samlpCfg.GetAttr("flexible_mappings"))
		if err == nil {
			samlpAddon.FlexibleMappings = flexibleMappings
		}

		var logout management.SAML2ClientAddonLogout

		samlpCfg.GetAttr("logout").ForEachElement(func(_ cty.Value, logoutCfg cty.Value) (stop bool) {
			logout = management.SAML2ClientAddonLogout{
				Callback:   value.String(logoutCfg.GetAttr("callback")),
				SLOEnabled: value.Bool(logoutCfg.GetAttr("slo_enabled")),
			}

			return stop
		})

		if logout != (management.SAML2ClientAddonLogout{}) {
			samlpAddon.Logout = &logout
		}

		if samlpAddon.DigestAlgorithm == nil {
			samlpAddon.DigestAlgorithm = auth0.String("sha1")
		}

		if samlpAddon.NameIdentifierFormat == nil {
			samlpAddon.NameIdentifierFormat = auth0.String("urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified")
		}

		if samlpAddon.SignatureAlgorithm == nil {
			samlpAddon.SignatureAlgorithm = auth0.String("rsa-sha1")
		}

		if samlpAddon.LifetimeInSeconds == nil {
			samlpAddon.LifetimeInSeconds = auth0.Int(3600)
		}

		if samlpAddon.CreateUPNClaim == nil {
			samlpAddon.CreateUPNClaim = auth0.Bool(true)
		}

		if samlpAddon.IncludeAttributeNameFormat == nil {
			samlpAddon.IncludeAttributeNameFormat = auth0.Bool(true)
		}

		if samlpAddon.MapIdentities == nil {
			samlpAddon.MapIdentities = auth0.Bool(true)
		}

		if samlpAddon.MapUnknownClaimsAsIs == nil {
			samlpAddon.MapUnknownClaimsAsIs = auth0.Bool(false)
		}

		if samlpAddon.PassthroughClaimsWithNoMapping == nil {
			samlpAddon.PassthroughClaimsWithNoMapping = auth0.Bool(true)
		}

		if samlpAddon.TypedAttributes == nil {
			samlpAddon.TypedAttributes = auth0.Bool(true)
		}

		return stop
	})

	return &samlpAddon
}

func clientHasChange(c *management.Client) bool {
	return c.String() != "{}"
}

func expandClientGrant(data *schema.ResourceData) *management.ClientGrant {
	cfg := data.GetRawConfig()

	clientGrant := &management.ClientGrant{}

	if data.IsNewResource() {
		clientGrant.ClientID = value.String(cfg.GetAttr("client_id"))
		clientGrant.Audience = value.String(cfg.GetAttr("audience"))
		clientGrant.SubjectType = value.String(cfg.GetAttr("subject_type"))
	}

	if data.IsNewResource() || data.HasChange("scopes") {
		clientGrant.Scope = value.Strings(cfg.GetAttr("scopes"))
	}

	if data.IsNewResource() || data.HasChange("allow_any_organization") {
		clientGrant.AllowAnyOrganization = value.Bool(cfg.GetAttr("allow_any_organization"))
	}

	if data.IsNewResource() || data.HasChange("organization_usage") {
		clientGrant.OrganizationUsage = value.String(cfg.GetAttr("organization_usage"))
	}

	if data.IsNewResource() || data.HasChange("authorization_details_types") {
		clientGrant.AuthorizationDetailsTypes = value.Strings(cfg.GetAttr("authorization_details_types"))
	}

	return clientGrant
}

func expandSessionTransfer(data *schema.ResourceData) *management.SessionTransfer {
	if !data.IsNewResource() && !data.HasChange("session_transfer") {
		return nil
	}

	sessionTransferConfig := data.GetRawConfig().GetAttr("session_transfer")
	if sessionTransferConfig.IsNull() {
		return nil
	}

	// Handles case when session_transfer is not defined.
	_, ok := data.GetOk("session_transfer")
	if !ok {
		return nil
	}

	var sessionTransfer management.SessionTransfer

	sessionTransferConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		sessionTransfer.CanCreateSessionTransferToken = value.Bool(config.GetAttr("can_create_session_transfer_token"))
		sessionTransfer.AllowedAuthenticationMethods = value.Strings(config.GetAttr("allowed_authentication_methods"))
		sessionTransfer.EnforceDeviceBinding = value.String(config.GetAttr("enforce_device_binding"))
		sessionTransfer.AllowRefreshToken = value.Bool(config.GetAttr("allow_refresh_token"))

		enforceOnlineRefreshTokens := data.Get("session_transfer.0.enforce_online_refresh_tokens").(bool)
		sessionTransfer.EnforceOnlineRefreshTokens = auth0.Bool(enforceOnlineRefreshTokens)

		enforceCascadeRevocation := data.Get("session_transfer.0.enforce_cascade_revocation").(bool)
		sessionTransfer.EnforceCascadeRevocation = auth0.Bool(enforceCascadeRevocation)

		return stop
	})

	return &sessionTransfer
}

func fetchNullableFields(data *schema.ResourceData, client *management.Client) map[string]interface{} {
	type nullCheckFunc func(*schema.ResourceData) bool

	checks := map[string]nullCheckFunc{
		"default_organization":    isDefaultOrgNull,
		"session_transfer":        isSessionTransferNull,
		"oidc_backchannel_logout": isOIDCLogoutNull,
		"cross_origin_loc":        isCrossOriginLocNull,
		"encryption_key": func(d *schema.ResourceData) bool {
			return isEncryptionKeyNull(d) && !d.IsNewResource()
		},
		"addons": func(_ *schema.ResourceData) bool {
			return clientHasChange(client) && client.GetAddons() != nil
		},
		"token_quota": commons.IsTokenQuotaNull,
		"skip_non_verifiable_callback_uri_confirmation_prompt": isSkipNonVerifiableCallbackURIConfirmationPromptNull,
	}

	nullableMap := make(map[string]interface{})

	for field, checkFunc := range checks {
		if checkFunc(data) {
			nullableMap[field] = nil
		}
	}

	return nullableMap
}

func isDefaultOrgNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("default_organization") {
		return false
	}
	empty := true
	config := data.GetRawConfig()
	defaultOrgConfig := config.GetAttr("default_organization")
	if defaultOrgConfig.IsNull() || defaultOrgConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		disable := cfg.GetAttr("disable")
		flows := cfg.GetAttr("flows")
		organizationID := cfg.GetAttr("organization_id")

		if (!disable.IsNull() && disable.True()) || (flows.IsNull() && organizationID.IsNull()) {
			stop = true
		} else {
			empty = false
		}
		return stop
	}) {
		// We forced an early return because it was disabled.
		return true
	}
	return empty
}

func isEncryptionKeyNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("encryption_key") {
		return false
	}

	config := data.GetRawConfig().GetAttr("encryption_key")

	if config.IsNull() {
		return true
	}

	empty := true
	config.ForEachElement(func(_, val cty.Value) (stop bool) {
		if !val.IsNull() && val.AsString() != "" {
			empty = false
		}
		return false
	})

	return empty
}

func isCrossOriginLocNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("cross_origin_loc") {
		return false
	}

	config := data.GetRawConfig()
	attr := config.GetAttr("cross_origin_loc")

	// If it's null, it means it was not set.
	return attr.IsNull()
}

func isSessionTransferNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("session_transfer") {
		return false
	}

	rawConfig := data.GetRawConfig().GetAttr("session_transfer")

	// If the session_transfer block is explicitly set to null.
	if rawConfig.IsNull() {
		return true
	}

	// If the session_transfer block exists, but all fields inside it are null or not set.
	empty := true
	rawConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		canCreate := cfg.GetAttr("can_create_session_transfer_token")
		enforceBinding := cfg.GetAttr("enforce_device_binding")
		allowedMethods := cfg.GetAttr("allowed_authentication_methods")

		if (!canCreate.IsNull() && canCreate.True()) ||
			(!enforceBinding.IsNull() && enforceBinding.AsString() != "") ||
			(!allowedMethods.IsNull() && allowedMethods.LengthInt() > 0) {
			empty = false
		}

		return stop
	})

	return empty
}

func isOIDCLogoutNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("oidc_logout") {
		return false
	}

	rawConfig := data.GetRawConfig().GetAttr("oidc_logout")

	// If oidc_logout is explicitly set to null.
	if rawConfig.IsNull() {
		return true
	}

	empty := true
	rawConfig.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		logoutURLs := cfg.GetAttr("backchannel_logout_urls")
		initiators := cfg.GetAttr("backchannel_logout_initiators")
		sessionMetadata := cfg.GetAttr("backchannel_logout_session_metadata")

		// backchannel_logout_urls is a TypeSet of strings
		if !logoutURLs.IsNull() && logoutURLs.LengthInt() > 0 {
			empty = false
			return stop
		}

		if !initiators.IsNull() {
			initiators.ForEachElement(func(_ cty.Value, initiatorCfg cty.Value) (stop bool) {
				mode := initiatorCfg.GetAttr("mode")
				selected := initiatorCfg.GetAttr("selected_initiators")

				if (!mode.IsNull() && mode.AsString() != "") ||
					(!selected.IsNull() && selected.LengthInt() > 0) {
					empty = false
				}
				return stop
			})
		}

		if !sessionMetadata.IsNull() {
			sessionMetadata.ForEachElement(func(_ cty.Value, metaCfg cty.Value) (stop bool) {
				include := metaCfg.GetAttr("include")
				if !include.IsNull() {
					empty = false
				}
				return stop
			})
		}

		return stop
	})

	return empty
}

func isSkipNonVerifiableCallbackURIConfirmationPromptNull(data *schema.ResourceData) bool {
	if !data.IsNewResource() && !data.HasChange("skip_non_verifiable_callback_uri_confirmation_prompt") {
		return false
	}

	rawConfig := data.GetRawConfig()
	if rawConfig.IsNull() {
		return true
	}

	return rawConfig.GetAttr("skip_non_verifiable_callback_uri_confirmation_prompt").IsNull()
}

func expandExpressConfiguration(data *schema.ResourceData) *management.ExpressConfiguration {
	config := data.GetRawConfig()
	expressConfig := config.GetAttr("express_configuration")

	if expressConfig.IsNull() || expressConfig.LengthInt() == 0 {
		return nil
	}

	var result *management.ExpressConfiguration

	expressConfig.ForEachElement(func(_ cty.Value, elem cty.Value) (stop bool) {
		result = &management.ExpressConfiguration{
			InitiateLoginURITemplate: value.String(elem.GetAttr("initiate_login_uri_template")),
			UserAttributeProfileID:   value.String(elem.GetAttr("user_attribute_profile_id")),
			ConnectionProfileID:      value.String(elem.GetAttr("connection_profile_id")),
			EnableClient:             value.Bool(elem.GetAttr("enable_client")),
			EnableOrganization:       value.Bool(elem.GetAttr("enable_organization")),
			AdminLoginDomain:         value.String(elem.GetAttr("admin_login_domain")),
			OktaOINClientID:          value.String(elem.GetAttr("okta_oin_client_id")),
			OINSubmissionID:          value.String(elem.GetAttr("okta_oin_client_id")),
		}

		linkedClientsAttr := elem.GetAttr("linked_clients")
		if !linkedClientsAttr.IsNull() && linkedClientsAttr.LengthInt() > 0 {
			linkedClients := make([]management.LinkedClient, 0)
			linkedClientsAttr.ForEachElement(func(_ cty.Value, lc cty.Value) (stop bool) {
				linkedClients = append(linkedClients, management.LinkedClient{
					ClientID: value.String(lc.GetAttr("client_id")),
				})
				return stop
			})
			result.LinkedClients = &linkedClients
		}

		return stop
	})

	return result
}
