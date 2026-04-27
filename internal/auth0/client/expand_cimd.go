package client

import (
	mgmtv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandCIMDClient(data *schema.ResourceData) *mgmtv2.UpdateClientRequestContent {
	config := data.GetRawConfig()
	req := &mgmtv2.UpdateClientRequestContent{}

	req.Description = value.String(config.GetAttr("description"))

	if appType := value.String(config.GetAttr("app_type")); appType != nil {
		v := mgmtv2.ClientAppTypeEnum(*appType)
		req.AppType = &v
	}

	req.OidcConformant = value.Bool(config.GetAttr("oidc_conformant"))
	req.RequireProofOfPossession = value.Bool(config.GetAttr("require_proof_of_possession"))

	req.SkipNonVerifiableCallbackURIConfirmationPrompt = value.Bool(config.GetAttr("skip_non_verifiable_callback_uri_confirmation_prompt"))

	if origins := value.Strings(config.GetAttr("allowed_origins")); origins != nil {
		req.AllowedOrigins = *origins
	}
	if webOrigins := value.Strings(config.GetAttr("web_origins")); webOrigins != nil {
		req.WebOrigins = *webOrigins
	}
	if grantTypes := value.Strings(config.GetAttr("grant_types")); grantTypes != nil {
		req.GrantTypes = *grantTypes
	}

	if orgDiscovery := config.GetAttr("organization_discovery_methods"); !orgDiscovery.IsNull() {
		if methods := value.Strings(orgDiscovery); methods != nil && len(*methods) > 0 {
			enumMethods := make([]mgmtv2.ClientOrganizationDiscoveryEnum, len(*methods))
			for i, m := range *methods {
				enumMethods[i] = mgmtv2.ClientOrganizationDiscoveryEnum(m)
			}
			req.OrganizationDiscoveryMethods = enumMethods
		}
	}

	if rp := value.String(config.GetAttr("redirection_policy")); rp != nil {
		v := mgmtv2.ClientRedirectionPolicyEnum(*rp)
		req.RedirectionPolicy = &v
	}

	req.JwtConfiguration = expandCIMDJwtConfiguration(data)
	req.RefreshToken = expandCIMDRefreshToken(data)
	req.DefaultOrganization = expandCIMDDefaultOrganization(data)
	req.ClientMetadata = expandCIMDClientMetadata(data)
	req.TokenQuota = expandCIMDTokenQuota(data)

	return req
}

func expandCIMDJwtConfiguration(data *schema.ResourceData) *mgmtv2.ClientJwtConfiguration {
	jwtConfig := data.GetRawConfig().GetAttr("jwt_configuration")
	if jwtConfig.IsNull() || jwtConfig.LengthInt() == 0 {
		return nil
	}

	var jwt mgmtv2.ClientJwtConfiguration

	jwtConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		jwt.LifetimeInSeconds = value.Int(config.GetAttr("lifetime_in_seconds"))
		jwt.SecretEncoded = value.Bool(config.GetAttr("secret_encoded"))

		if alg := value.String(config.GetAttr("alg")); alg != nil {
			v := mgmtv2.SigningAlgorithmEnum(*alg)
			jwt.Alg = &v
		}

		return stop
	})

	return &jwt
}

func expandCIMDRefreshToken(data *schema.ResourceData) *mgmtv2.ClientRefreshTokenConfiguration {
	rtConfig := data.GetRawConfig().GetAttr("refresh_token")
	if rtConfig.IsNull() || rtConfig.LengthInt() == 0 {
		return nil
	}

	var rt mgmtv2.ClientRefreshTokenConfiguration
	rtConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		if rotType := value.String(config.GetAttr("rotation_type")); rotType != nil {
			rt.RotationType = mgmtv2.RefreshTokenRotationTypeEnum(*rotType)
		}
		if expType := value.String(config.GetAttr("expiration_type")); expType != nil {
			rt.ExpirationType = mgmtv2.RefreshTokenExpirationTypeEnum(*expType)
		}
		rt.Leeway = value.Int(config.GetAttr("leeway"))
		rt.TokenLifetime = value.Int(config.GetAttr("token_lifetime"))
		rt.InfiniteTokenLifetime = value.Bool(config.GetAttr("infinite_token_lifetime"))
		rt.IdleTokenLifetime = value.Int(config.GetAttr("idle_token_lifetime"))
		rt.InfiniteIdleTokenLifetime = value.Bool(config.GetAttr("infinite_idle_token_lifetime"))

		return stop
	})

	return &rt
}

func expandCIMDDefaultOrganization(data *schema.ResourceData) *mgmtv2.ClientDefaultOrganization {
	if !data.IsNewResource() && !data.HasChange("default_organization") {
		return nil
	}

	config := data.GetRawConfig().GetAttr("default_organization")
	if config.IsNull() || config.LengthInt() == 0 {
		return nil
	}

	var defaultOrg mgmtv2.ClientDefaultOrganization

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		if orgID := value.String(cfg.GetAttr("organization_id")); orgID != nil {
			defaultOrg.OrganizationID = *orgID
		}
		if flows := value.Strings(cfg.GetAttr("flows")); flows != nil {
			enumFlows := make([]mgmtv2.ClientDefaultOrganizationFlowsEnum, len(*flows))
			for i, f := range *flows {
				enumFlows[i] = mgmtv2.ClientDefaultOrganizationFlowsEnum(f)
			}
			defaultOrg.Flows = enumFlows
		}

		return stop
	})

	return &defaultOrg
}

func expandCIMDClientMetadata(data *schema.ResourceData) *mgmtv2.ClientMetadata {
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

	if len(newMetadataMap) == 0 {
		return nil
	}

	return &newMetadataMap
}

func expandCIMDTokenQuota(data *schema.ResourceData) *mgmtv2.UpdateTokenQuota {
	config := data.GetRawConfig().GetAttr("token_quota")
	if config.IsNull() {
		return nil
	}

	var quota *mgmtv2.UpdateTokenQuota

	config.ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		clientCredsValue := cfg.GetAttr("client_credentials")
		if clientCredsValue.IsNull() {
			return false
		}

		clientCredsValue.ForEachElement(func(_ cty.Value, credsConfig cty.Value) (stop bool) {
			enforce := value.Bool(credsConfig.GetAttr("enforce"))
			perHour := value.Int(credsConfig.GetAttr("per_hour"))
			perDay := value.Int(credsConfig.GetAttr("per_day"))

			quota = &mgmtv2.UpdateTokenQuota{
				ClientCredentials: &mgmtv2.TokenQuotaClientCredentials{
					Enforce: enforce,
				},
			}

			if perHour != nil {
				quota.ClientCredentials.PerHour = perHour
			}
			if perDay != nil {
				quota.ClientCredentials.PerDay = perDay
			}

			return false
		})

		return false
	})

	return quota
}

func isEmptyRequest(req *mgmtv2.UpdateClientRequestContent) (bool, error) {
	reqBytes, err := req.MarshalJSON()
	if err != nil {
		return false, err
	}
	return string(reqBytes) == "{}", nil
}

func applyCIMDNullFields(data *schema.ResourceData, req *mgmtv2.UpdateClientRequestContent) {
	config := data.GetRawConfig()

	if data.HasChange("allowed_origins") && config.GetAttr("allowed_origins").IsNull() {
		req.SetAllowedOrigins([]string{})
	}

	if data.HasChange("web_origins") && config.GetAttr("web_origins").IsNull() {
		req.SetWebOrigins([]string{})
	}

	if data.HasChange("organization_discovery_methods") && config.GetAttr("organization_discovery_methods").IsNull() {
		req.SetOrganizationDiscoveryMethods(nil)
	}

	if data.HasChange("client_metadata") && config.GetAttr("client_metadata").IsNull() {
		req.SetClientMetadata(&map[string]any{})
	}

	if data.HasChange("default_organization") &&
		(config.GetAttr("default_organization").IsNull() || config.GetAttr("default_organization").LengthInt() == 0) {
		req.SetDefaultOrganization(nil)
	}

	if data.HasChange("skip_non_verifiable_callback_uri_confirmation_prompt") && config.GetAttr("skip_non_verifiable_callback_uri_confirmation_prompt").IsNull() {
		req.SetSkipNonVerifiableCallbackURIConfirmationPrompt(nil)
	}

	if commons.IsTokenQuotaNull(data) {
		req.SetTokenQuota(nil)
	}
}
