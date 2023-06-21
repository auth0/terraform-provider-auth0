package okta

import (
	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/connection"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewConnectionResource will return a new auth0_connection_okta resource.
func NewConnectionResource() *schema.Resource {
	baseResource := connection.NewBaseConnectionResource(
		"This resource configure your Okta Enterprise Connection to allow your users to use their enterprise credentials to login to your app.",
		map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The strategy's client ID.",
			},
			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The strategy's client secret.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Domain name.",
			},
			"domain_aliases": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Description: "List of the domains that can be authenticated using the identity provider. " +
					"Only needed for Identifier First authentication flows.",
			},
			"scopes": {
				Type:     schema.TypeSet,
				Computed: true,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Description: "Permissions to grant to the connection. Within the Auth0 dashboard these appear " +
					"under the \"Attributes\" and \"Extended Attributes\" sections. Some examples: " +
					"`basic_profile`, `ext_profile`, `ext_nested_groups`, etc.",
			},
			"issuer": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Issuer URL, e.g. `https://auth.example.com`.",
			},
			"jwks_uri": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "JWKS URI.",
			},
			"token_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Token endpoint.",
			},
			"userinfo_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "User info endpoint.",
			},
			"authorization_endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Authorization endpoint.",
			},
			"icon_url": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Icon URL.",
			},
		},
		expandConnectionOkta,
		flattenConnectionOkta,
	)

	return baseResource
}

func flattenConnectionOkta(
	_ *schema.ResourceData,
	options *management.ConnectionOptionsOkta,
) (map[string]interface{}, diag.Diagnostics) {
	m := map[string]interface{}{
		"strategy":                 "okta",
		"client_id":                options.GetClientID(),
		"client_secret":            options.GetClientSecret(),
		"domain":                   options.GetDomain(),
		"domain_aliases":           options.GetDomainAliases(),
		"scopes":                   options.Scopes(),
		"issuer":                   options.GetIssuer(),
		"jwks_uri":                 options.GetJWKSURI(),
		"token_endpoint":           options.GetTokenEndpoint(),
		"userinfo_endpoint":        options.GetUserInfoEndpoint(),
		"authorization_endpoint":   options.GetAuthorizationEndpoint(),
		"non_persistent_attrs":     options.GetNonPersistentAttrs(),
		"set_user_root_attributes": options.GetSetUserAttributes(),
		"icon_url":                 options.GetLogoURL(),
	}

	upstreamParams, err := structure.FlattenJsonToString(options.UpstreamParams)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	m["upstream_params"] = upstreamParams

	return m, nil
}

func expandConnectionOkta(
	conn *management.Connection,
	d *schema.ResourceData,
	_ *management.Management,
) (*management.Connection, diag.Diagnostics) {
	config := d.GetRawConfig()

	options := &management.ConnectionOptionsOkta{
		ClientID:              value.String(config.GetAttr("client_id")),
		ClientSecret:          value.String(config.GetAttr("client_secret")),
		Domain:                value.String(config.GetAttr("domain")),
		DomainAliases:         value.Strings(config.GetAttr("domain_aliases")),
		AuthorizationEndpoint: value.String(config.GetAttr("authorization_endpoint")),
		Issuer:                value.String(config.GetAttr("issuer")),
		JWKSURI:               value.String(config.GetAttr("jwks_uri")),
		UserInfoEndpoint:      value.String(config.GetAttr("userinfo_endpoint")),
		TokenEndpoint:         value.String(config.GetAttr("token_endpoint")),
		NonPersistentAttrs:    value.Strings(config.GetAttr("non_persistent_attrs")),
		SetUserAttributes:     value.String(config.GetAttr("set_user_root_attributes")),
		LogoURL:               value.String(config.GetAttr("icon_url")),
	}

	expandConnectionOptionsScopes(d, options)

	var err error
	options.UpstreamParams, err = value.MapFromJSON(config.GetAttr("upstream_params"))
	if err != nil {
		return nil, diag.FromErr(err)
	}

	if d.IsNewResource() {
		conn.Strategy = auth0.String("okta")
	}

	conn.Options = options

	return conn, nil
}

type scoper interface {
	Scopes() []string
	SetScopes(enable bool, scopes ...string)
}

func expandConnectionOptionsScopes(d *schema.ResourceData, s scoper) {
	scopesList := d.Get("scopes").(*schema.Set).List()

	_, scopesToDisable := value.Difference(d, "scopes")
	for _, scope := range scopesList {
		s.SetScopes(true, scope.(string))
	}

	for _, scope := range scopesToDisable {
		s.SetScopes(false, scope.(string))
	}
}
