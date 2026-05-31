package auth0client

import (
	"context"
	"fmt"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// =========================================================================
// flatten: SDK response -> Terraform model
// =========================================================================

// commonClientFields is everything we currently flatten back into state. The
// Create / Get / Update response structs in go-auth0 v2 are field-identical
// for `Client`, so this struct is effectively the full superset.
type commonClientFields struct {
	// identity
	ClientID     *string
	ClientSecret *string
	Tenant       *string
	Global       *bool

	// basic
	Name        *string
	Description *string
	AppType     *mgmt.ClientAppTypeEnum
	LogoURI     *string

	// behaviour
	IsFirstParty                                   *bool
	OidcConformant                                 *bool
	SSO                                            *bool
	SSODisabled                                    *bool
	CrossOriginAuthentication                      *bool
	CustomLoginPageOn                              *bool
	IsTokenEndpointIPHeaderTrusted                 *bool
	RequirePushedAuthorizationRequests             *bool
	RequireProofOfPossession                       *bool
	SkipNonVerifiableCallbackURIConfirmationPrompt *bool

	// URLs
	Callbacks         []string
	AllowedLogoutURLs []string
	AllowedOrigins    []string
	WebOrigins        []string
	ClientAliases     []string
	AllowedClients    []string

	// OAuth/OIDC
	GrantTypes              []string
	TokenEndpointAuthMethod *mgmt.ClientTokenEndpointAuthMethodEnum
	CrossOriginLoc          *string
	InitiateLoginURI        *string
	FormTemplate            *string
	CustomLoginPage         *string
	CustomLoginPagePreview  *string
	ParRequestExpiry        *int
	ComplianceLevel         *mgmt.ClientComplianceLevelEnum
	ThirdPartySecurityMode  *mgmt.ClientThirdPartySecurityModeEnum
	RedirectionPolicy       *mgmt.ClientRedirectionPolicyEnum
	JwksURI                 *string

	// org
	OrganizationUsage            *mgmt.ClientOrganizationUsageEnum
	OrganizationRequireBehavior  *mgmt.ClientOrganizationRequireBehaviorEnum
	OrganizationDiscoveryMethods []mgmt.ClientOrganizationDiscoveryEnum

	// external / CIMD
	ResourceServerIdentifier  *string
	ExternalMetadataType      *mgmt.ClientExternalMetadataTypeEnum
	ExternalMetadataCreatedBy *mgmt.ClientExternalMetadataCreatedByEnum
	ExternalClientID          *string

	// metadata + nested
	ClientMetadata *mgmt.ClientMetadata
	SigningKeys    mgmt.ClientSigningKeys // []*ClientSigningKey
	TokenQuota     *mgmt.TokenQuota

	// Wave 2
	JwtConfiguration                  *mgmt.ClientJwtConfiguration
	RefreshToken                      *mgmt.ClientRefreshTokenConfiguration
	OidcLogout                        *mgmt.ClientOidcBackchannelLogoutSettings
	OidcBackchannelLogout             *mgmt.ClientOidcBackchannelLogoutSettings
	EncryptionKey                     *mgmt.ClientEncryptionKey
	DefaultOrganization               *mgmt.ClientDefaultOrganization
	NativeSocialLogin                 *mgmt.NativeSocialLogin
	SessionTransfer                   *mgmt.ClientSessionTransferConfiguration
	Mobile                            *mgmt.ClientMobile
	TokenExchange                     *mgmt.ClientTokenExchangeConfiguration
	MyOrganizationConfiguration       *mgmt.ClientMyOrganizationResponseConfiguration
	ExpressConfiguration              *mgmt.ExpressConfiguration
	AsyncApprovalNotificationChannels *mgmt.ClientAsyncApprovalNotificationsChannelsAPIPostConfiguration
	// signed_request_object: any so we accept both Public-Key and CredentialID variants.
	SignedRequestObject any
	Addons              *mgmt.ClientAddons
	// client_authentication_methods: pointer-or-pointer (Get returns the
	// non-Create flavour). We accept it as a generic any value.
	ClientAuthenticationMethods any
}

func flattenInto(_ context.Context, m *model, c commonClientFields, diags *diag.Diagnostics) {
	// identity
	m.ID = types.StringPointerValue(c.ClientID)
	m.ClientID = types.StringPointerValue(c.ClientID)
	m.ClientSecret = types.StringPointerValue(c.ClientSecret)
	m.Tenant = types.StringPointerValue(c.Tenant)
	m.Global = types.BoolPointerValue(c.Global)

	// basic
	m.Name = types.StringPointerValue(c.Name)
	m.Description = types.StringPointerValue(c.Description)
	m.AppType = framework.EnumPtrToString(c.AppType)
	m.LogoURI = types.StringPointerValue(c.LogoURI)

	// behaviour flags
	m.IsFirstParty = types.BoolPointerValue(c.IsFirstParty)
	m.OIDCConformant = types.BoolPointerValue(c.OidcConformant)
	m.SSO = types.BoolPointerValue(c.SSO)
	m.SSODisabled = types.BoolPointerValue(c.SSODisabled)
	m.CrossOriginAuthentication = types.BoolPointerValue(c.CrossOriginAuthentication)
	m.CustomLoginPageOn = types.BoolPointerValue(c.CustomLoginPageOn)
	m.IsTokenEndpointIPHeaderTrusted = types.BoolPointerValue(c.IsTokenEndpointIPHeaderTrusted)
	m.RequirePushedAuthorizationRequests = types.BoolPointerValue(c.RequirePushedAuthorizationRequests)
	m.RequireProofOfPossession = types.BoolPointerValue(c.RequireProofOfPossession)
	m.SkipNonVerifiableCallbackURIConfirmationPrompt = types.BoolPointerValue(c.SkipNonVerifiableCallbackURIConfirmationPrompt)

	// URLs
	m.Callbacks = framework.StringSliceToList(c.Callbacks)
	m.AllowedLogoutURLs = framework.StringSliceToList(c.AllowedLogoutURLs)
	m.AllowedOrigins = framework.StringSliceToList(c.AllowedOrigins)
	m.WebOrigins = framework.StringSliceToList(c.WebOrigins)
	m.ClientAliases = framework.StringSliceToList(c.ClientAliases)
	m.AllowedClients = framework.StringSliceToList(c.AllowedClients)

	// OAuth/OIDC
	m.GrantTypes = framework.StringSliceToList(c.GrantTypes)
	m.TokenEndpointAuthMethod = framework.EnumPtrToString(c.TokenEndpointAuthMethod)
	m.CrossOriginLoc = types.StringPointerValue(c.CrossOriginLoc)
	m.InitiateLoginURI = types.StringPointerValue(c.InitiateLoginURI)
	m.FormTemplate = types.StringPointerValue(c.FormTemplate)
	m.CustomLoginPage = types.StringPointerValue(c.CustomLoginPage)
	m.CustomLoginPagePreview = types.StringPointerValue(c.CustomLoginPagePreview)
	m.ParRequestExpiry = framework.IntPtrToInt64(c.ParRequestExpiry)
	m.ComplianceLevel = framework.EnumPtrToString(c.ComplianceLevel)
	m.ThirdPartySecurityMode = framework.EnumPtrToString(c.ThirdPartySecurityMode)
	m.RedirectionPolicy = framework.EnumPtrToString(c.RedirectionPolicy)
	m.JwksURI = types.StringPointerValue(c.JwksURI)

	// org
	m.OrganizationUsage = framework.EnumPtrToString(c.OrganizationUsage)
	m.OrganizationRequireBehavior = framework.EnumPtrToString(c.OrganizationRequireBehavior)
	m.OrganizationDiscoveryMethods = framework.EnumSliceToList(c.OrganizationDiscoveryMethods)

	// external / CIMD
	m.ResourceServerIdentifier = types.StringPointerValue(c.ResourceServerIdentifier)
	m.ExternalMetadataType = framework.EnumPtrToString(c.ExternalMetadataType)
	m.ExternalMetadataCreatedBy = framework.EnumPtrToString(c.ExternalMetadataCreatedBy)
	m.ExternalClientID = types.StringPointerValue(c.ExternalClientID)

	// metadata
	m.ClientMetadata = clientMetadataToTfMap(c.ClientMetadata)

	// nested
	m.SigningKeys = flattenSigningKeys(c.SigningKeys, diags)
	m.TokenQuota = flattenTokenQuota(c.TokenQuota, diags)

	// Wave 2 nested objects
	m.JwtConfiguration = flattenJwtConfiguration(c.JwtConfiguration, diags)
	m.RefreshToken = flattenRefreshToken(c.RefreshToken, diags)
	m.OidcLogout = flattenOidcLogout(c.OidcLogout, diags)
	m.OidcBackchannelLogout = flattenOidcLogout(c.OidcBackchannelLogout, diags)
	m.EncryptionKey = flattenEncryptionKey(c.EncryptionKey, diags)
	m.DefaultOrganization = flattenDefaultOrganization(c.DefaultOrganization, diags)
	m.NativeSocialLogin = flattenNativeSocialLogin(c.NativeSocialLogin, diags)
	m.SessionTransfer = flattenSessionTransfer(c.SessionTransfer, diags)
	m.Mobile = flattenMobile(c.Mobile, diags)
	m.TokenExchange = flattenTokenExchange(c.TokenExchange, diags)
	m.MyOrganizationConfiguration = flattenMyOrgConfig(c.MyOrganizationConfiguration, diags)
	m.ExpressConfiguration = flattenExpressConfiguration(c.ExpressConfiguration, diags)

	// list-of-enum and JSON-string fields
	if c.AsyncApprovalNotificationChannels == nil {
		m.AsyncApprovalNotificationChannels = types.ListNull(types.StringType)
	} else {
		m.AsyncApprovalNotificationChannels = framework.EnumSliceToList([]mgmt.AsyncApprovalNotificationsChannelsEnum(*c.AsyncApprovalNotificationChannels))
	}
	m.SignedRequestObject = flattenSignedRequestObject(c.SignedRequestObject, diags)
	m.Addons = flattenAddons(c.Addons, diags)
	m.ClientAuthenticationMethods = framework.FlattenJSONToString(c.ClientAuthenticationMethods, diags)
}

func flattenCreate(ctx context.Context, m *model, c *mgmt.CreateClientResponseContent, diags *diag.Diagnostics) {
	flattenInto(ctx, m, commonClientFields{
		ClientID: c.ClientID, ClientSecret: c.ClientSecret, Tenant: c.Tenant, Global: c.Global,
		Name: c.Name, Description: c.Description, AppType: c.AppType, LogoURI: c.LogoURI,
		IsFirstParty: c.IsFirstParty, OidcConformant: c.OidcConformant,
		SSO: c.SSO, SSODisabled: c.SSODisabled, CrossOriginAuthentication: c.CrossOriginAuthentication,
		CustomLoginPageOn: c.CustomLoginPageOn, IsTokenEndpointIPHeaderTrusted: c.IsTokenEndpointIPHeaderTrusted,
		RequirePushedAuthorizationRequests:             c.RequirePushedAuthorizationRequests,
		RequireProofOfPossession:                       c.RequireProofOfPossession,
		SkipNonVerifiableCallbackURIConfirmationPrompt: c.SkipNonVerifiableCallbackURIConfirmationPrompt,
		Callbacks: c.Callbacks, AllowedLogoutURLs: c.AllowedLogoutURLs, AllowedOrigins: c.AllowedOrigins,
		WebOrigins: c.WebOrigins, ClientAliases: c.ClientAliases, AllowedClients: c.AllowedClients,
		GrantTypes:              c.GrantTypes,
		TokenEndpointAuthMethod: c.TokenEndpointAuthMethod,
		CrossOriginLoc:          c.CrossOriginLoc, InitiateLoginURI: c.InitiateLoginURI,
		FormTemplate: c.FormTemplate, CustomLoginPage: c.CustomLoginPage,
		CustomLoginPagePreview: c.CustomLoginPagePreview, ParRequestExpiry: c.ParRequestExpiry,
		ComplianceLevel: c.ComplianceLevel, ThirdPartySecurityMode: c.ThirdPartySecurityMode,
		RedirectionPolicy: c.RedirectionPolicy, JwksURI: c.JwksURI,
		OrganizationUsage: c.OrganizationUsage, OrganizationRequireBehavior: c.OrganizationRequireBehavior,
		OrganizationDiscoveryMethods: c.OrganizationDiscoveryMethods,
		ResourceServerIdentifier:     c.ResourceServerIdentifier,
		ExternalMetadataType:         c.ExternalMetadataType,
		ExternalMetadataCreatedBy:    c.ExternalMetadataCreatedBy,
		ExternalClientID:             c.ExternalClientID,
		ClientMetadata:               c.ClientMetadata,
		SigningKeys:                  derefSigningKeys(c.SigningKeys),
		TokenQuota:                   c.TokenQuota,
		// Wave 2
		JwtConfiguration:                  c.JwtConfiguration,
		RefreshToken:                      c.RefreshToken,
		OidcLogout:                        c.OidcLogout,
		EncryptionKey:                     c.EncryptionKey,
		DefaultOrganization:               c.DefaultOrganization,
		SessionTransfer:                   c.SessionTransfer,
		Mobile:                            c.Mobile,
		TokenExchange:                     c.TokenExchange,
		MyOrganizationConfiguration:       c.MyOrganizationConfiguration,
		ExpressConfiguration:              c.ExpressConfiguration,
		AsyncApprovalNotificationChannels: c.AsyncApprovalNotificationChannels,
		SignedRequestObject:               c.SignedRequestObject,
		Addons:                            c.Addons,
		ClientAuthenticationMethods:       c.ClientAuthenticationMethods,
	}, diags)
}

func flattenGet(ctx context.Context, m *model, c *mgmt.GetClientResponseContent, diags *diag.Diagnostics) {
	flattenInto(ctx, m, commonClientFields{
		ClientID: c.ClientID, ClientSecret: c.ClientSecret, Tenant: c.Tenant, Global: c.Global,
		Name: c.Name, Description: c.Description, AppType: c.AppType, LogoURI: c.LogoURI,
		IsFirstParty: c.IsFirstParty, OidcConformant: c.OidcConformant,
		SSO: c.SSO, SSODisabled: c.SSODisabled, CrossOriginAuthentication: c.CrossOriginAuthentication,
		CustomLoginPageOn: c.CustomLoginPageOn, IsTokenEndpointIPHeaderTrusted: c.IsTokenEndpointIPHeaderTrusted,
		RequirePushedAuthorizationRequests:             c.RequirePushedAuthorizationRequests,
		RequireProofOfPossession:                       c.RequireProofOfPossession,
		SkipNonVerifiableCallbackURIConfirmationPrompt: c.SkipNonVerifiableCallbackURIConfirmationPrompt,
		Callbacks: c.Callbacks, AllowedLogoutURLs: c.AllowedLogoutURLs, AllowedOrigins: c.AllowedOrigins,
		WebOrigins: c.WebOrigins, ClientAliases: c.ClientAliases, AllowedClients: c.AllowedClients,
		GrantTypes:              c.GrantTypes,
		TokenEndpointAuthMethod: c.TokenEndpointAuthMethod,
		CrossOriginLoc:          c.CrossOriginLoc, InitiateLoginURI: c.InitiateLoginURI,
		FormTemplate: c.FormTemplate, CustomLoginPage: c.CustomLoginPage,
		CustomLoginPagePreview: c.CustomLoginPagePreview, ParRequestExpiry: c.ParRequestExpiry,
		ComplianceLevel: c.ComplianceLevel, ThirdPartySecurityMode: c.ThirdPartySecurityMode,
		RedirectionPolicy: c.RedirectionPolicy, JwksURI: c.JwksURI,
		OrganizationUsage: c.OrganizationUsage, OrganizationRequireBehavior: c.OrganizationRequireBehavior,
		OrganizationDiscoveryMethods: c.OrganizationDiscoveryMethods,
		ResourceServerIdentifier:     c.ResourceServerIdentifier,
		ExternalMetadataType:         c.ExternalMetadataType,
		ExternalMetadataCreatedBy:    c.ExternalMetadataCreatedBy,
		ExternalClientID:             c.ExternalClientID,
		ClientMetadata:               c.ClientMetadata,
		SigningKeys:                  derefSigningKeys(c.SigningKeys),
		TokenQuota:                   c.TokenQuota,
		// Wave 2
		JwtConfiguration:                  c.JwtConfiguration,
		RefreshToken:                      c.RefreshToken,
		OidcLogout:                        c.OidcLogout,
		EncryptionKey:                     c.EncryptionKey,
		DefaultOrganization:               c.DefaultOrganization,
		SessionTransfer:                   c.SessionTransfer,
		Mobile:                            c.Mobile,
		TokenExchange:                     c.TokenExchange,
		MyOrganizationConfiguration:       c.MyOrganizationConfiguration,
		ExpressConfiguration:              c.ExpressConfiguration,
		AsyncApprovalNotificationChannels: c.AsyncApprovalNotificationChannels,
		SignedRequestObject:               c.SignedRequestObject,
		Addons:                            c.Addons,
		ClientAuthenticationMethods:       c.ClientAuthenticationMethods,
	}, diags)
}

func flattenUpdate(ctx context.Context, m *model, c *mgmt.UpdateClientResponseContent, diags *diag.Diagnostics) {
	flattenInto(ctx, m, commonClientFields{
		ClientID: c.ClientID, ClientSecret: c.ClientSecret, Tenant: c.Tenant, Global: c.Global,
		Name: c.Name, Description: c.Description, AppType: c.AppType, LogoURI: c.LogoURI,
		IsFirstParty: c.IsFirstParty, OidcConformant: c.OidcConformant,
		SSO: c.SSO, SSODisabled: c.SSODisabled, CrossOriginAuthentication: c.CrossOriginAuthentication,
		CustomLoginPageOn: c.CustomLoginPageOn, IsTokenEndpointIPHeaderTrusted: c.IsTokenEndpointIPHeaderTrusted,
		RequirePushedAuthorizationRequests:             c.RequirePushedAuthorizationRequests,
		RequireProofOfPossession:                       c.RequireProofOfPossession,
		SkipNonVerifiableCallbackURIConfirmationPrompt: c.SkipNonVerifiableCallbackURIConfirmationPrompt,
		Callbacks: c.Callbacks, AllowedLogoutURLs: c.AllowedLogoutURLs, AllowedOrigins: c.AllowedOrigins,
		WebOrigins: c.WebOrigins, ClientAliases: c.ClientAliases, AllowedClients: c.AllowedClients,
		GrantTypes: c.GrantTypes,
		// NB: UpdateClientResponseContent.TokenEndpointAuthMethod is the …Enum
		// (not the `OrNullEnum`), so this assignment matches.
		TokenEndpointAuthMethod: c.TokenEndpointAuthMethod,
		CrossOriginLoc:          c.CrossOriginLoc, InitiateLoginURI: c.InitiateLoginURI,
		FormTemplate: c.FormTemplate, CustomLoginPage: c.CustomLoginPage,
		CustomLoginPagePreview: c.CustomLoginPagePreview, ParRequestExpiry: c.ParRequestExpiry,
		ComplianceLevel: c.ComplianceLevel, ThirdPartySecurityMode: c.ThirdPartySecurityMode,
		RedirectionPolicy: c.RedirectionPolicy, JwksURI: c.JwksURI,
		OrganizationUsage: c.OrganizationUsage, OrganizationRequireBehavior: c.OrganizationRequireBehavior,
		OrganizationDiscoveryMethods: c.OrganizationDiscoveryMethods,
		ResourceServerIdentifier:     c.ResourceServerIdentifier,
		ExternalMetadataType:         c.ExternalMetadataType,
		ExternalMetadataCreatedBy:    c.ExternalMetadataCreatedBy,
		ExternalClientID:             c.ExternalClientID,
		ClientMetadata:               c.ClientMetadata,
		SigningKeys:                  derefSigningKeys(c.SigningKeys),
		TokenQuota:                   c.TokenQuota,
		// Wave 2
		JwtConfiguration:                  c.JwtConfiguration,
		RefreshToken:                      c.RefreshToken,
		OidcLogout:                        c.OidcLogout,
		EncryptionKey:                     c.EncryptionKey,
		DefaultOrganization:               c.DefaultOrganization,
		SessionTransfer:                   c.SessionTransfer,
		Mobile:                            c.Mobile,
		TokenExchange:                     c.TokenExchange,
		MyOrganizationConfiguration:       c.MyOrganizationConfiguration,
		ExpressConfiguration:              c.ExpressConfiguration,
		AsyncApprovalNotificationChannels: c.AsyncApprovalNotificationChannels,
		SignedRequestObject:               c.SignedRequestObject,
		Addons:                            c.Addons,
		ClientAuthenticationMethods:       c.ClientAuthenticationMethods,
	}, diags)
}

// =========================================================================
// flatten helpers (top-level + nested)
// =========================================================================

// clientMetadataToTfMap stringifies any value coming back so we can store it
// in a `map[string]string`. Non-string scalars are coerced via fmt.Sprintf.
func clientMetadataToTfMap(in *mgmt.ClientMetadata) types.Map {
	if in == nil {
		return types.MapNull(types.StringType)
	}
	asString := make(map[string]string, len(*in))
	for k, v := range *in {
		switch t := v.(type) {
		case nil:
			asString[k] = ""
		case string:
			asString[k] = t
		default:
			asString[k] = fmt.Sprintf("%v", t)
		}
	}
	els := make(map[string]attr.Value, len(asString))
	for k, v := range asString {
		els[k] = types.StringValue(v)
	}
	mv, _ := types.MapValue(types.StringType, els)
	return mv
}

func flattenTokenQuota(in *mgmt.TokenQuota, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(tokenQuotaAttrTypes())
	}
	ccVal := types.ObjectNull(tokenQuotaCCAttrTypes())
	if in.ClientCredentials != nil {
		cc := in.ClientCredentials
		v, d := types.ObjectValue(tokenQuotaCCAttrTypes(), map[string]attr.Value{
			"enforce":  types.BoolPointerValue(cc.Enforce),
			"per_day":  framework.IntPtrToInt64(cc.PerDay),
			"per_hour": framework.IntPtrToInt64(cc.PerHour),
		})
		diags.Append(d...)
		ccVal = v
	}
	out, d := types.ObjectValue(tokenQuotaAttrTypes(), map[string]attr.Value{
		"client_credentials": ccVal,
	})
	diags.Append(d...)
	return out
}

// derefSigningKeys turns the SDK's *ClientSigningKeys (pointer to slice alias)
// into a plain slice so flattenSigningKeys has one shape to work with.
func derefSigningKeys(in *mgmt.ClientSigningKeys) mgmt.ClientSigningKeys {
	if in == nil {
		return nil
	}
	return *in
}

func flattenSigningKeys(in mgmt.ClientSigningKeys, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: signingKeyAttrTypes()}
	if in == nil {
		return types.ListNull(objType)
	}
	vals := make([]attr.Value, 0, len(in))
	for _, k := range in {
		if k == nil {
			continue
		}
		v, d := types.ObjectValue(signingKeyAttrTypes(), map[string]attr.Value{
			"pkcs7":   types.StringPointerValue(k.Pkcs7),
			"cert":    types.StringPointerValue(k.Cert),
			"subject": types.StringPointerValue(k.Subject),
		})
		diags.Append(d...)
		vals = append(vals, v)
	}
	lv, d := types.ListValue(objType, vals)
	diags.Append(d...)
	return lv
}
