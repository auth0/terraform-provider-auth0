package provider

import (
	"context"
	"net/http"
	"strconv"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

func newClient() *schema.Resource {
	return &schema.Resource{
		CreateContext: createClient,
		ReadContext:   readClient,
		UpdateContext: updateClient,
		DeleteContext: deleteClient,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With this resource, you can set up applications that use Auth0 for authentication " +
			"and configure allowed callback URLs and secrets for these applications.",
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the client.",
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 140),
				Description:  "Description of the purpose of the client.",
			},
			"client_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the client.",
			},
			"client_secret": {
				Type:      schema.TypeString,
				Computed:  true,
				Sensitive: true,
				Description: "Secret for the client. Keep this private. To access this attribute you need to add the " +
					"`read:client_keys` scope to the Terraform client. Otherwise, the attribute will contain an " +
					"empty string.",
			},
			"client_secret_rotation_trigger": {
				Type:     schema.TypeMap,
				Optional: true,
				Description: "Custom metadata for the rotation. " +
					"For more info: [rotate-client-secret](https://auth0.com/docs/get-started/applications/rotate-client-secret).",
			},
			"app_type": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"native", "spa", "regular_web", "non_interactive", "rms",
					"box", "cloudbees", "concur", "dropbox", "mscrm", "echosign",
					"egnyte", "newrelic", "office365", "salesforce", "sentry",
					"sharepoint", "slack", "springcm", "sso_integration", "zendesk", "zoom",
				}, false),
				Description: "Type of application the client represents.",
			},
			"logo_uri": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "URL of the logo for the client. Recommended size is 150px x 150px. " +
					"If none is set, the default badge for the application type will be shown.",
			},
			"is_first_party": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether this client is a first-party client.",
			},
			"is_token_endpoint_ip_header_trusted": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether the token endpoint IP header is trusted.",
			},
			"oidc_conformant": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether this client will conform to strict OIDC specifications.",
			},
			"callbacks": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Description: "URLs that Auth0 may call back to after a user authenticates for the client. " +
					"Make sure to specify the protocol (https://) otherwise the callback may fail in some cases. " +
					"With the exception of custom URI schemes for native clients, all callbacks should use protocol https://.",
			},
			"allowed_logout_urls": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "URLs that Auth0 may redirect to after logout.",
			},
			"grant_types": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Optional:    true,
				Description: "Types of grants that this client is authorized to use.",
			},
			"organization_usage": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"deny", "allow", "require",
				}, false),
				Description: "Defines how to proceed during an authentication transaction with " +
					"regards to an organization. Can be `deny` (default), `allow` or `require`.",
			},
			"organization_require_behavior": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"no_prompt", "pre_login_prompt",
				}, false),
				Description: "Defines how to proceed during an authentication transaction when " +
					"`organization_usage = \"require\"`. Can be `no_prompt` (default) or `pre_login_prompt`.",
			},
			"allowed_origins": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Description: "URLs that represent valid origins for cross-origin resource sharing. " +
					"By default, all your callback URLs will be allowed.",
			},
			"allowed_clients": {
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Description: "List of applications ID's that will be allowed to make delegation request. " +
					"By default, all applications will be allowed.",
			},
			"web_origins": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Description: "URLs that represent valid web origins for use with web message response mode.",
			},
			"jwt_configuration": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for the JWTs issued for this client.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"lifetime_in_seconds": {
							Type:        schema.TypeInt,
							Optional:    true,
							Computed:    true,
							Description: "Number of seconds during which the JWT will be valid.",
						},
						"secret_encoded": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							ForceNew:    true,
							Description: "Indicates whether the client secret is Base64-encoded.",
						},
						"scopes": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Permissions (scopes) included in JWTs.",
						},
						"alg": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Algorithm used to sign JWTs.",
						},
					},
				},
			},
			"encryption_key": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Encryption used for WS-Fed responses with this client.",
			},
			"sso": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Applies only to SSO clients and determines whether Auth0 will handle " +
					"Single Sign-On (true) or whether the identity provider will (false).",
			},
			"sso_disabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether or not SSO is disabled.",
			},
			"cross_origin_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "Whether this client can be used to make cross-origin authentication requests (true) " +
					"or it is not allowed to make such requests (false).",
			},
			"cross_origin_loc": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "URL of the location in your site where the cross-origin verification " +
					"takes place for the cross-origin auth flow when performing authentication in your own " +
					"domain instead of Auth0 Universal Login page.",
			},
			"custom_login_page_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Indicates whether a custom login page is to be used.",
			},
			"custom_login_page": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The content (HTML, CSS, JS) of the custom login page.",
			},
			"form_template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "HTML form template to be used for WS-Federation.",
			},
			"addons": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Addons enabled for this client and their associated configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"azure_blob": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"azure_sb": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"rms": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"mscrm": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"slack": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"sentry": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"box": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"cloudbees": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"concur": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"dropbox": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"echosign": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"egnyte": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"firebase": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"newrelic": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"office365": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"salesforce": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"salesforce_api": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"salesforce_sandbox_api": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"samlp": {
							Type:        schema.TypeList,
							MaxItems:    1,
							Computed:    true,
							Optional:    true,
							Description: "Configuration settings for a SAML add-on.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audience": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "Audience of the SAML Assertion. " +
											"Default will be the Issuer on SAMLRequest.",
									},
									"recipient": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "Recipient of the SAML Assertion (SubjectConfirmationData). " +
											"Default is `AssertionConsumerUrl` on SAMLRequest or " +
											"callback URL if no SAMLRequest was sent.",
									},
									"mappings": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem:     schema.TypeString,
										Description: "Mappings between the Auth0 user profile property " +
											"name (`name`) and the output attributes on the SAML " +
											"attribute in the assertion (`value`).",
									},
									"create_upn_claim": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether a UPN claim should be created. " +
											"Defaults to `true`.",
									},
									"passthrough_claims_with_no_mapping": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether or not to passthrough " +
											"claims that are not mapped to the common profile " +
											"in the output assertion. Defaults to `true`.",
									},
									"map_unknown_claims_as_is": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  false,
										Description: "Indicates whether to add a prefix of `http://schema.auth0.com` " +
											"to any claims that are not mapped to the common profile when passed " +
											"through in the output assertion. Defaults to `false`.",
									},
									"map_identities": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether or not to add additional identity " +
											"information in the token, such as the provider used and the " +
											"`access_token`, if available. Defaults to `true`.",
									},
									"signature_algorithm": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "rsa-sha1",
										Description: "Algorithm used to sign the SAML Assertion or response. " +
											"Options include `rsa-sha1` and `rsa-sha256`. Defaults to `rsa-sha1`.",
									},
									"digest_algorithm": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "sha1",
										Description: "Algorithm used to calculate the digest of the SAML Assertion " +
											"or response. Options include `sha1` and `sha256`. Defaults to `sha1`.",
									},
									"destination": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "Destination of the SAML Response. If not specified, " +
											"it will be `AssertionConsumerUrl` of SAMLRequest " +
											"or callback URL if there was no SAMLRequest.",
									},
									"lifetime_in_seconds": {
										Type:        schema.TypeInt,
										Optional:    true,
										Default:     3600,
										Description: "Number of seconds during which the token is valid.",
									},
									"sign_response": {
										Type:     schema.TypeBool,
										Optional: true,
										Description: "Indicates whether or not the SAML Response should be signed " +
											"instead of the SAML Assertion.",
									},
									"name_identifier_format": {
										Type:        schema.TypeString,
										Optional:    true,
										Default:     "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
										Description: "Format of the name identifier.",
									},
									"name_identifier_probes": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Optional: true,
										Description: "Attributes that can be used for Subject/NameID. " +
											"Auth0 will try each of the attributes of this array in " +
											"order and use the first value it finds.",
									},
									"authn_context_class_ref": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Class reference of the authentication context.",
									},
									"typed_attributes": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether or not we should infer the `xs:type` " +
											"of the element. Types include `xs:string`, `xs:boolean`, `xs:double`, " +
											"and `xs:anyType`. When set to false, all `xs:type` are `xs:anyType`. " +
											"Defaults to `true`.",
									},
									"include_attribute_name_format": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether or not we should infer the NameFormat " +
											"based on the attribute name. If set to false, the attribute " +
											"NameFormat is not set in the assertion. Defaults to `true`.",
									},
									"logout": {
										Type:        schema.TypeMap,
										Optional:    true,
										Description: "Configuration settings for logout.",
									},
									"binding": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Protocol binding used for SAML logout responses.",
									},
									"signing_cert": {
										Type:     schema.TypeString,
										Optional: true,
										Description: "Optionally indicates the public key certificate used to " +
											"validate SAML requests. If set, SAML requests will be required to " +
											"be signed. A sample value would be `-----BEGIN PUBLIC KEY-----\\nMIGf...bpP/t3\\n+JGNGIRMj1hF1rnb6QIDAQAB\\n-----END PUBLIC KEY-----\\n`.",
									},
								},
							},
						},
						"layer": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"sap_api": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"sharepoint": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"springcm": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"wams": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"wsfed": {
							Type:     schema.TypeMap,
							Optional: true,
							Description: "WS-Fed (WIF) addon indicator. Actual configuration is stored " +
								"in callback and `client_aliases` properties on the client.",
						},
						"zendesk": {
							Type:     schema.TypeMap,
							Optional: true,
						},
						"zoom": {
							Type:     schema.TypeMap,
							Optional: true,
						},
					},
				},
			},
			"token_endpoint_auth_method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					"client_secret_post",
					"client_secret_basic",
				}, false),
				Description: "Defines the requested authentication method for the token endpoint. " +
					"Options include `none` (public client without a client secret), " +
					"`client_secret_post` (client uses HTTP POST parameters), " +
					"`client_secret_basic` (client uses HTTP Basic).",
			},
			"client_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
				Description: "Metadata associated with the client, in the form of an object with string values " +
					"(max 255 chars). Maximum of 10 metadata properties allowed. Field names (max 255 chars) are " +
					"alphanumeric and may only include the following special characters: " +
					"`:,-+=_*?\"/\\()<>@ [Tab] [Space]`.",
			},
			"mobile": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Additional configuration for native mobile apps.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"android": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Configuration settings for Android native apps.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"app_package_name": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.android.0.app_package_name",
											"mobile.0.android.0.sha256_cert_fingerprints",
										},
									},
									"sha256_cert_fingerprints": {
										Type:     schema.TypeList,
										Optional: true,
										Elem:     &schema.Schema{Type: schema.TypeString},
										AtLeastOneOf: []string{
											"mobile.0.android.0.app_package_name",
											"mobile.0.android.0.sha256_cert_fingerprints",
										},
									},
								},
							},
						},
						"ios": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Configuration settings for i0S native apps.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"team_id": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.ios.0.team_id",
											"mobile.0.ios.0.app_bundle_identifier",
										},
									},
									"app_bundle_identifier": {
										Type:     schema.TypeString,
										Optional: true,
										AtLeastOneOf: []string{
											"mobile.0.ios.0.team_id",
											"mobile.0.ios.0.app_bundle_identifier",
										},
									},
								},
							},
						},
					},
				},
			},
			"initiate_login_uri": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.All(
					validation.IsURLWithScheme([]string{"https"}),
					internalValidation.IsURLWithNoFragment,
				),
				Description: "Initiate login URI, must be HTTPS.",
			},
			"native_social_login": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 1,
				Description: "Configuration settings to toggle native social login for mobile native applications. " +
					"Once this is set it must stay set, with both resources set to `false` in order to change the `app_type`.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"apple": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
						"facebook": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"enabled": {
										Type:     schema.TypeBool,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"refresh_token": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configuration settings for the refresh tokens issued for this client.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rotation_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"rotating",
								"non-rotating",
							}, false),
							Description: "Options include `rotating`, `non-rotating`. When `rotating`, exchanging " +
								"a refresh token will cause a new refresh token to be issued and the existing " +
								"token will be invalidated. This allows for automatic detection of token reuse " +
								"if the token is leaked.",
						},
						"expiration_type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"expiring",
								"non-expiring",
							}, false),
							Description: "Options include `expiring`, `non-expiring`. Whether a refresh token " +
								"will expire based on an absolute lifetime, after which the token can no " +
								"longer be used. If rotation is `rotating`, this must be set to `expiring`.",
						},
						"leeway": {
							Computed: true,
							Type:     schema.TypeInt,
							Optional: true,
							Description: "The amount of time in seconds in which a refresh token may be " +
								"reused without triggering reuse detection.",
						},
						"token_lifetime": {
							Computed:    true,
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The absolute lifetime of a refresh token in seconds.",
						},
						"infinite_token_lifetime": {
							Computed: true,
							Type:     schema.TypeBool,
							Optional: true,
							Description: "Whether refresh tokens should remain valid indefinitely. " +
								"If false, `token_lifetime` should also be set.",
						},
						"infinite_idle_token_lifetime": {
							Computed:    true,
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Whether inactive refresh tokens should remain valid indefinitely.",
						},
						"idle_token_lifetime": {
							Computed:    true,
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "The time in seconds after which inactive refresh tokens will expire.",
						},
					},
				},
			},
			"signing_keys": {
				Type:      schema.TypeList,
				Elem:      &schema.Schema{Type: schema.TypeMap},
				Computed:  true,
				Sensitive: true,
				Description: "List containing a map of the public cert of the signing key and the public cert " +
					"of the signing key in PKCS7.",
			},
		},
	}
}

func createClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := expandClient(d)
	if err != nil {
		return diag.FromErr(err)
	}

	api := m.(*management.Management)
	if err := api.Client.Create(client); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(auth0.StringValue(client.ClientID))
	return readClient(ctx, d, m)
}

func readClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	client, err := api.Client.Read(d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("client_id", client.ClientID),
		d.Set("client_secret", client.ClientSecret),
		d.Set("name", client.Name),
		d.Set("description", client.Description),
		d.Set("app_type", client.AppType),
		d.Set("logo_uri", client.LogoURI),
		d.Set("is_first_party", client.IsFirstParty),
		d.Set("is_token_endpoint_ip_header_trusted", client.IsTokenEndpointIPHeaderTrusted),
		d.Set("oidc_conformant", client.OIDCConformant),
		d.Set("callbacks", client.Callbacks),
		d.Set("allowed_logout_urls", client.AllowedLogoutURLs),
		d.Set("allowed_origins", client.AllowedOrigins),
		d.Set("allowed_clients", client.AllowedClients),
		d.Set("grant_types", client.GrantTypes),
		d.Set("organization_usage", client.OrganizationUsage),
		d.Set("organization_require_behavior", client.OrganizationRequireBehavior),
		d.Set("web_origins", client.WebOrigins),
		d.Set("sso", client.SSO),
		d.Set("sso_disabled", client.SSODisabled),
		d.Set("cross_origin_auth", client.CrossOriginAuth),
		d.Set("cross_origin_loc", client.CrossOriginLocation),
		d.Set("custom_login_page_on", client.CustomLoginPageOn),
		d.Set("custom_login_page", client.CustomLoginPage),
		d.Set("form_template", client.FormTemplate),
		d.Set("token_endpoint_auth_method", client.TokenEndpointAuthMethod),
		d.Set("native_social_login", flattenCustomSocialConfiguration(client.NativeSocialLogin)),
		d.Set("jwt_configuration", flattenClientJwtConfiguration(client.JWTConfiguration)),
		d.Set("refresh_token", flattenClientRefreshTokenConfiguration(client.RefreshToken)),
		d.Set("encryption_key", client.EncryptionKey),
		d.Set("addons", flattenClientAddons(client.Addons)),
		d.Set("client_metadata", client.ClientMetadata),
		d.Set("mobile", flattenClientMobile(client.Mobile)),
		d.Set("initiate_login_uri", client.InitiateLoginURI),
		d.Set("signing_keys", client.SigningKeys),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client, err := expandClient(d)
	if err != nil {
		return diag.FromErr(err)
	}

	api := m.(*management.Management)
	if clientHasChange(client) {
		if err := api.Client.Update(d.Id(), client); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(true)
	if err := rotateClientSecret(d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readClient(ctx, d, m)
}

func deleteClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*management.Management)
	if err := api.Client.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok {
			if mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
		}
	}

	return nil
}

func expandClient(d *schema.ResourceData) (*management.Client, error) {
	client := &management.Client{
		Name:                           String(d, "name"),
		Description:                    String(d, "description"),
		AppType:                        String(d, "app_type"),
		LogoURI:                        String(d, "logo_uri"),
		IsFirstParty:                   Bool(d, "is_first_party"),
		IsTokenEndpointIPHeaderTrusted: Bool(d, "is_token_endpoint_ip_header_trusted"),
		OIDCConformant:                 Bool(d, "oidc_conformant"),
		Callbacks:                      Slice(d, "callbacks"),
		AllowedLogoutURLs:              Slice(d, "allowed_logout_urls"),
		AllowedOrigins:                 Slice(d, "allowed_origins"),
		AllowedClients:                 Slice(d, "allowed_clients"),
		GrantTypes:                     Slice(d, "grant_types"),
		OrganizationUsage:              String(d, "organization_usage"),
		OrganizationRequireBehavior:    String(d, "organization_require_behavior"),
		WebOrigins:                     Slice(d, "web_origins"),
		SSO:                            Bool(d, "sso"),
		SSODisabled:                    Bool(d, "sso_disabled"),
		CrossOriginAuth:                Bool(d, "cross_origin_auth"),
		CrossOriginLocation:            String(d, "cross_origin_loc"),
		CustomLoginPageOn:              Bool(d, "custom_login_page_on"),
		CustomLoginPage:                String(d, "custom_login_page"),
		FormTemplate:                   String(d, "form_template"),
		TokenEndpointAuthMethod:        String(d, "token_endpoint_auth_method"),
		InitiateLoginURI:               String(d, "initiate_login_uri"),
	}

	List(d, "refresh_token", IsNewResource(), HasChange()).Elem(func(d ResourceData) {
		client.RefreshToken = &management.ClientRefreshToken{
			RotationType:              String(d, "rotation_type"),
			ExpirationType:            String(d, "expiration_type"),
			Leeway:                    Int(d, "leeway"),
			TokenLifetime:             Int(d, "token_lifetime"),
			InfiniteTokenLifetime:     Bool(d, "infinite_token_lifetime"),
			InfiniteIdleTokenLifetime: Bool(d, "infinite_idle_token_lifetime"),
			IdleTokenLifetime:         Int(d, "idle_token_lifetime"),
		}
	})

	List(d, "jwt_configuration").Elem(func(d ResourceData) {
		client.JWTConfiguration = &management.ClientJWTConfiguration{
			LifetimeInSeconds: Int(d, "lifetime_in_seconds"),
			SecretEncoded:     Bool(d, "secret_encoded", IsNewResource()),
			Algorithm:         String(d, "alg"),
			Scopes:            Map(d, "scopes"),
		}
	})

	if encryptionKeyMap := Map(d, "encryption_key"); encryptionKeyMap != nil {
		client.EncryptionKey = map[string]string{}
		for key, value := range encryptionKeyMap {
			client.EncryptionKey[key] = value.(string)
		}
	}

	var result *multierror.Error
	List(d, "addons").Elem(func(d ResourceData) {
		client.Addons = make(map[string]interface{})

		for _, name := range []string{
			"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
			"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
			"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
			"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
			"springcm", "wams", "wsfed", "zendesk", "zoom",
		} {
			if _, ok := d.GetOk(name); ok {
				client.Addons[name] = mapFromState(Map(d, name))
			}
		}

		List(d, "samlp").Elem(func(d ResourceData) {
			m := make(MapData)

			result = multierror.Append(
				m.Set("audience", String(d, "audience")),
				m.Set("authnContextClassRef", String(d, "authn_context_class_ref")),
				m.Set("binding", String(d, "binding")),
				m.Set("signingCert", String(d, "signing_cert")),
				m.Set("createUpnClaim", Bool(d, "create_upn_claim")),
				m.Set("destination", String(d, "destination")),
				m.Set("digestAlgorithm", String(d, "digest_algorithm")),
				m.Set("includeAttributeNameFormat", Bool(d, "include_attribute_name_format")),
				m.Set("lifetimeInSeconds", Int(d, "lifetime_in_seconds")),
				m.Set("mapIdentities", Bool(d, "map_identities")),
				m.Set("mappings", Map(d, "mappings")),
				m.Set("mapUnknownClaimsAsIs", Bool(d, "map_unknown_claims_as_is")),
				m.Set("nameIdentifierFormat", String(d, "name_identifier_format")),
				m.Set("nameIdentifierProbes", Slice(d, "name_identifier_probes")),
				m.Set("passthroughClaimsWithNoMapping", Bool(d, "passthrough_claims_with_no_mapping")),
				m.Set("recipient", String(d, "recipient")),
				m.Set("signatureAlgorithm", String(d, "signature_algorithm")),
				m.Set("signResponse", Bool(d, "sign_response")),
				m.Set("typedAttributes", Bool(d, "typed_attributes")),
				m.Set("logout", mapFromState(Map(d, "logout"))),
			)

			client.Addons["samlp"] = m
		})
	})

	if clientMetadata, ok := d.GetOk("client_metadata"); ok {
		client.ClientMetadata = make(map[string]string)
		for key, value := range clientMetadata.(map[string]interface{}) {
			client.ClientMetadata[key] = value.(string)
		}
	}

	List(d, "native_social_login").Elem(func(d ResourceData) {
		client.NativeSocialLogin = &management.ClientNativeSocialLogin{}

		List(d, "apple").Elem(func(d ResourceData) {
			m := make(MapData)
			result = multierror.Append(result, m.Set("enabled", Bool(d, "enabled")))

			client.NativeSocialLogin.Apple = m
		})

		List(d, "facebook").Elem(func(d ResourceData) {
			m := make(MapData)
			result = multierror.Append(result, m.Set("enabled", Bool(d, "enabled")))

			client.NativeSocialLogin.Facebook = m
		})
	})

	List(d, "mobile").Elem(func(d ResourceData) {
		client.Mobile = make(map[string]interface{})

		List(d, "android").Elem(func(d ResourceData) {
			m := make(MapData)
			result = multierror.Append(
				result,
				m.Set("app_package_name", String(d, "app_package_name")),
				m.Set("sha256_cert_fingerprints", Slice(d, "sha256_cert_fingerprints")),
			)

			client.Mobile["android"] = m
		})

		List(d, "ios").Elem(func(d ResourceData) {
			m := make(MapData)

			result = multierror.Append(
				result,
				m.Set("team_id", String(d, "team_id")),
				m.Set("app_bundle_identifier", String(d, "app_bundle_identifier")),
			)

			client.Mobile["ios"] = m
		})
	})

	return client, result.ErrorOrNil()
}

func mapFromState(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, value := range input {
		switch v := value.(type) {
		case string:
			if i, err := strconv.ParseInt(v, 10, 64); err == nil {
				output[key] = i
			} else if f, err := strconv.ParseFloat(v, 64); err == nil {
				output[key] = f
			} else if b, err := strconv.ParseBool(v); err == nil {
				output[key] = b
			} else {
				output[key] = v
			}
		case map[string]interface{}:
			output[key] = mapFromState(v)
		case []interface{}:
			output[key] = v
		default:
			output[key] = v
		}
	}

	return output
}

func mapToState(input map[string]interface{}) map[string]interface{} {
	output := make(map[string]interface{})

	for key, v := range input {
		switch value := v.(type) {
		case bool:
			if value {
				output[key] = "true"
			} else {
				output[key] = "false"
			}
		case float64:
			output[key] = strconv.Itoa(int(value))
		case int:
			output[key] = strconv.Itoa(value)
		default:
			output[key] = value
		}
	}

	return output
}

func rotateClientSecret(d *schema.ResourceData, m interface{}) error {
	if d.HasChange("client_secret_rotation_trigger") {
		api := m.(*management.Management)
		client, err := api.Client.RotateSecret(d.Id())
		if err != nil {
			return err
		}

		if err := d.Set("client_secret", client.ClientSecret); err != nil {
			return err
		}
	}
	return nil
}

func clientHasChange(c *management.Client) bool {
	return c.String() != "{}"
}

func flattenCustomSocialConfiguration(customSocial *management.ClientNativeSocialLogin) []interface{} {
	if customSocial == nil {
		return nil
	}

	m := make(map[string]interface{})

	if customSocial.Apple != nil {
		m["apple"] = []interface{}{
			map[string]interface{}{
				"enabled": customSocial.Apple["enabled"],
			},
		}
	}
	if customSocial.Facebook != nil {
		m["facebook"] = []interface{}{
			map[string]interface{}{
				"enabled": customSocial.Facebook["enabled"],
			},
		}
	}

	return []interface{}{m}
}

func flattenClientJwtConfiguration(jwt *management.ClientJWTConfiguration) []interface{} {
	m := make(map[string]interface{})
	if jwt != nil {
		m["lifetime_in_seconds"] = jwt.LifetimeInSeconds
		m["secret_encoded"] = jwt.SecretEncoded
		m["scopes"] = jwt.Scopes
		m["alg"] = jwt.Algorithm
	}
	return []interface{}{m}
}

func flattenClientRefreshTokenConfiguration(refreshToken *management.ClientRefreshToken) []interface{} {
	if refreshToken == nil {
		return nil
	}

	m := make(map[string]interface{})

	m["rotation_type"] = refreshToken.RotationType
	m["expiration_type"] = refreshToken.ExpirationType
	m["leeway"] = refreshToken.Leeway
	m["token_lifetime"] = refreshToken.TokenLifetime
	m["infinite_token_lifetime"] = refreshToken.InfiniteTokenLifetime
	m["infinite_idle_token_lifetime"] = refreshToken.InfiniteIdleTokenLifetime
	m["idle_token_lifetime"] = refreshToken.IdleTokenLifetime

	return []interface{}{m}
}

func flattenClientAddons(addons map[string]interface{}) []interface{} {
	if addons == nil {
		return nil
	}

	m := make(map[string]interface{})

	if value, ok := addons["samlp"]; ok {
		samlp := value.(map[string]interface{})

		samlpMap := map[string]interface{}{
			"audience":                           samlp["audience"],
			"recipient":                          samlp["recipient"],
			"mappings":                           samlp["mappings"],
			"create_upn_claim":                   samlp["createUpnClaim"],
			"passthrough_claims_with_no_mapping": samlp["passthroughClaimsWithNoMapping"],
			"map_unknown_claims_as_is":           samlp["mapUnknownClaimsAsIs"],
			"map_identities":                     samlp["mapIdentities"],
			"signature_algorithm":                samlp["signatureAlgorithm"],
			"digest_algorithm":                   samlp["digestAlgorithm"],
			"destination":                        samlp["destination"],
			"lifetime_in_seconds":                samlp["lifetimeInSeconds"],
			"sign_response":                      samlp["signResponse"],
			"name_identifier_format":             samlp["nameIdentifierFormat"],
			"name_identifier_probes":             samlp["nameIdentifierProbes"],
			"authn_context_class_ref":            samlp["authnContextClassRef"],
			"typed_attributes":                   samlp["typedAttributes"],
			"include_attribute_name_format":      samlp["includeAttributeNameFormat"],
			"binding":                            samlp["binding"],
			"signing_cert":                       samlp["signingCert"],
		}

		if logout, ok := samlp["logout"].(map[string]interface{}); ok {
			samlpMap["logout"] = mapToState(logout)
		}

		m["samlp"] = []interface{}{samlpMap}
	}

	for _, name := range []string{
		"aws", "azure_blob", "azure_sb", "rms", "mscrm", "slack", "sentry",
		"box", "cloudbees", "concur", "dropbox", "echosign", "egnyte",
		"firebase", "newrelic", "office365", "salesforce", "salesforce_api",
		"salesforce_sandbox_api", "layer", "sap_api", "sharepoint",
		"springcm", "wams", "wsfed", "zendesk", "zoom",
	} {
		if value, ok := addons[name]; ok {
			if addonType, ok := value.(map[string]interface{}); ok {
				m[name] = mapToState(addonType)
			}
		}
	}

	return []interface{}{m}
}

func flattenClientMobile(mobile map[string]interface{}) []interface{} {
	if mobile == nil {
		return nil
	}

	m := make(map[string]interface{})

	if value, ok := mobile["android"]; ok {
		android := value.(map[string]interface{})

		m["android"] = []interface{}{
			map[string]interface{}{
				"app_package_name":         android["app_package_name"],
				"sha256_cert_fingerprints": android["sha256_cert_fingerprints"],
			},
		}
	}

	if value, ok := mobile["ios"]; ok {
		ios := value.(map[string]interface{})

		m["ios"] = []interface{}{
			map[string]interface{}{
				"team_id":               ios["team_id"],
				"app_bundle_identifier": ios["app_bundle_identifier"],
			},
		}
	}

	return []interface{}{m}
}
