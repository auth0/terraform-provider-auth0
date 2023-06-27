package client

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandClient(d *schema.ResourceData) *management.Client {
	config := d.GetRawConfig()

	client := &management.Client{
		Name:                           value.String(config.GetAttr("name")),
		Description:                    value.String(config.GetAttr("description")),
		AppType:                        value.String(config.GetAttr("app_type")),
		LogoURI:                        value.String(config.GetAttr("logo_uri")),
		IsFirstParty:                   value.Bool(config.GetAttr("is_first_party")),
		IsTokenEndpointIPHeaderTrusted: value.Bool(config.GetAttr("is_token_endpoint_ip_header_trusted")),
		OIDCConformant:                 value.Bool(config.GetAttr("oidc_conformant")),
		ClientAliases:                  value.Strings(config.GetAttr("client_aliases")),
		Callbacks:                      value.Strings(config.GetAttr("callbacks")),
		AllowedLogoutURLs:              value.Strings(config.GetAttr("allowed_logout_urls")),
		AllowedOrigins:                 value.Strings(config.GetAttr("allowed_origins")),
		AllowedClients:                 value.Strings(config.GetAttr("allowed_clients")),
		GrantTypes:                     value.Strings(config.GetAttr("grant_types")),
		OrganizationUsage:              value.String(config.GetAttr("organization_usage")),
		OrganizationRequireBehavior:    value.String(config.GetAttr("organization_require_behavior")),
		WebOrigins:                     value.Strings(config.GetAttr("web_origins")),
		SSO:                            value.Bool(config.GetAttr("sso")),
		SSODisabled:                    value.Bool(config.GetAttr("sso_disabled")),
		CrossOriginAuth:                value.Bool(config.GetAttr("cross_origin_auth")),
		CrossOriginLocation:            value.String(config.GetAttr("cross_origin_loc")),
		CustomLoginPageOn:              value.Bool(config.GetAttr("custom_login_page_on")),
		CustomLoginPage:                value.String(config.GetAttr("custom_login_page")),
		FormTemplate:                   value.String(config.GetAttr("form_template")),
		TokenEndpointAuthMethod:        value.String(config.GetAttr("token_endpoint_auth_method")),
		InitiateLoginURI:               value.String(config.GetAttr("initiate_login_uri")),
		EncryptionKey:                  value.MapOfStrings(config.GetAttr("encryption_key")),
		OIDCBackchannelLogout:          expandOIDCBackchannelLogout(d),
		ClientMetadata:                 expandClientMetadata(d),
		RefreshToken:                   expandClientRefreshToken(d),
		JWTConfiguration:               expandClientJWTConfiguration(d),
		// Addons:                         expandClientAddons(d), TODO: DXCDT-441 Add new go-auth0 v1-beta types.
		NativeSocialLogin: expandClientNativeSocialLogin(d),
		Mobile:            expandClientMobile(d),
	}

	return client
}

func expandOIDCBackchannelLogout(d *schema.ResourceData) *management.OIDCBackchannelLogout {
	raw := d.GetRawConfig().GetAttr("oidc_backchannel_logout_urls")

	logoutUrls := value.Strings(raw)

	if logoutUrls == nil {
		return nil
	}

	return &management.OIDCBackchannelLogout{
		BackChannelLogoutURLs: logoutUrls,
	}
}

func expandClientRefreshToken(d *schema.ResourceData) *management.ClientRefreshToken {
	refreshTokenConfig := d.GetRawConfig().GetAttr("refresh_token")
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
		return stop
	})

	if refreshToken == (management.ClientRefreshToken{}) {
		return nil
	}

	return &refreshToken
}

func expandClientJWTConfiguration(d *schema.ResourceData) *management.ClientJWTConfiguration {
	jwtConfig := d.GetRawConfig().GetAttr("jwt_configuration")
	if jwtConfig.IsNull() {
		return nil
	}

	var jwt management.ClientJWTConfiguration

	jwtConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		jwt.LifetimeInSeconds = value.Int(config.GetAttr("lifetime_in_seconds"))
		jwt.Algorithm = value.String(config.GetAttr("alg"))
		jwt.Scopes = value.MapOfStrings(config.GetAttr("scopes"))

		if d.IsNewResource() {
			jwt.SecretEncoded = value.Bool(config.GetAttr("secret_encoded"))
		}

		return stop
	})

	if jwt == (management.ClientJWTConfiguration{}) {
		return nil
	}

	return &jwt
}

func expandClientNativeSocialLogin(d *schema.ResourceData) *management.ClientNativeSocialLogin {
	nativeSocialLoginConfig := d.GetRawConfig().GetAttr("native_social_login")
	if nativeSocialLoginConfig.IsNull() {
		return nil
	}

	var nativeSocialLogin management.ClientNativeSocialLogin

	nativeSocialLoginConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		nativeSocialLogin.Apple = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("apple"))
		nativeSocialLogin.Facebook = expandClientNativeSocialLoginSupportEnabled(config.GetAttr("facebook"))
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

func expandClientMobile(d *schema.ResourceData) *management.ClientMobile {
	mobileConfig := d.GetRawConfig().GetAttr("mobile")
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

func expandClientMetadata(d *schema.ResourceData) *map[string]interface{} {
	if !d.HasChange("client_metadata") {
		return nil
	}

	oldMetadata, newMetadata := d.GetChange("client_metadata")
	oldMetadataMap := oldMetadata.(map[string]interface{})
	newMetadataMap := newMetadata.(map[string]interface{})

	for key := range oldMetadataMap {
		if _, ok := newMetadataMap[key]; !ok {
			newMetadataMap[key] = nil
		}
	}

	return &newMetadataMap
}

func clientHasChange(c *management.Client) bool {
	return c.String() != "{}"
}
