package client

import (
	"context"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

// NewResource will return a new auth0_client resource.
func NewResource() *schema.Resource {
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
					"empty string. Use this attribute on the `auth0_client_credentials` resource instead, to allow " +
					"managing it directly.",
				Deprecated: "Reading the client secret through this attribute is deprecated and it will be " +
					"removed in a future version. Migrate to the `auth0_client_credentials` resource to " +
					"manage a client's secret instead.",
			},
			"client_secret_rotation_trigger": {
				Type:     schema.TypeMap,
				Optional: true,
				Description: "Custom metadata for the rotation. " +
					"The contents of this map are arbitrary and are hashed by the provider. When the hash changes, a rotation is triggered. " +
					"For example, the map could contain the user making the change, the date of the change, and a text reason for the change. " +
					"For more info: [rotate-client-secret](https://auth0.com/docs/get-started/applications/rotate-client-secret). " +
					"Migrate to the `auth0_client_credentials` resource to manage a client's secret directly instead. " +
					"Refer to the [client secret rotation guide](Refer to the [client secret rotation guide](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/client_secret_rotation) " +
					"for instructions on how to rotate client secrets with zero downtime.",
				Deprecated: "Rotating a client's secret through this attribute is deprecated and it will be removed" +
					" in a future version. Migrate to the `auth0_client_credentials` resource to manage a client's " +
					"secret instead. " +
					"Refer to the [client secret rotation guide](Refer to the [client secret rotation guide](https://registry.terraform.io/providers/auth0/auth0/latest/docs/guides/client_secret_rotation) " +
					"for instructions on how to rotate client secrets with zero downtime.",
			},
			"client_aliases": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of audiences/realms for SAML protocol. Used by the wsfed addon.",
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
				Description: "Type of application the client represents. Possible values are: `native`, `spa`, " +
					"`regular_web`, `non_interactive`, `sso_integration`. Specific SSO integrations types accepted " +
					"as well are: `rms`, `box`, `cloudbees`, `concur`, `dropbox`, `mscrm`, `echosign`, `egnyte`, " +
					"`newrelic`, `office365`, `salesforce`, `sentry`, `sharepoint`, `slack`, `springcm`, `zendesk`, " +
					"`zoom`.",
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
			"oidc_backchannel_logout_urls": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "Set of URLs that are valid to call back from Auth0 for OIDC backchannel logout. Currently only one URL is allowed.",
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
				Description: "Whether this client can be used to make cross-origin authentication requests (`true`) " +
					"or it is not allowed to make such requests (`false`). Requires the `coa_toggle_enabled` " +
					"feature flag to be enabled on the tenant by the support team.",
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
				Deprecated: "Managing the authentication method through this attribute is deprecated and it will be " +
					"changed to read-only in a future version. Migrate to the `auth0_client_credentials` resource to " +
					"manage a client's authentication method instead. Check the " +
					"[MIGRATION GUIDE](https://github.com/auth0/terraform-provider-auth0/blob/main/MIGRATION_GUIDE.md) " +
					"on how to do that.",
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
							Computed:    true,
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
							Computed:    true,
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
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
				Description:  "Initiate login URI. Must be HTTPS or an empty string.",
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
							Computed: true,
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
							Computed: true,
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
			"addons": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Addons enabled for this client and their associated configurations.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "AWS Addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"principal": {
										Description: "AWS principal ARN, for example `arn:aws:iam::010616021751:saml-provider/idpname`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"role": {
										Description: "AWS role ARN, for example `arn:aws:iam::010616021751:role/foo`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"lifetime_in_seconds": {
										Description:  "AWS token lifetime in seconds.",
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntBetween(900, 43200),
										Optional:     true,
									},
								},
							},
						},
						"azure_blob": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Azure Blob Storage Addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_name": {
										Description: "Your Azure storage account name. Usually first segment in your " +
											"Azure storage URL, for example `https://acme-org.blob.core.windows.net` would " +
											"be the account name `acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
									"storage_access_key": {
										Description: "Access key associated with this storage account.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"container_name": {
										Description: "Container to request a token for, such as `my-container`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"blob_name": {
										Description: "Entity to request a token for, such as `my-blob`. If blank the " +
											"computed SAS will apply to the entire storage container.",
										Type:     schema.TypeString,
										Optional: true,
									},
									"expiration": {
										Description:  "Expiration in minutes for the generated token (default of 5 minutes).",
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntAtLeast(0),
										Optional:     true,
									},
									"signed_identifier": {
										Description: "Shared access policy identifier defined in your storage account resource.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"blob_read": {
										Description: "Indicates if the issued token has permission to read the " +
											"content, properties, metadata and block list. Use the blob as the " +
											"source of a copy operation.",
										Type:     schema.TypeBool,
										Optional: true,
									},
									"blob_write": {
										Description: "Indicates if the issued token has permission to create or " +
											"write content, properties, metadata, or block list. Snapshot or lease " +
											"the blob. Resize the blob (page blob only). Use the blob as the " +
											"destination of a copy operation within the same account.",
										Type:     schema.TypeBool,
										Optional: true,
									},
									"blob_delete": {
										Description: "Indicates if the issued token has permission to delete the blob.",
										Type:        schema.TypeBool,
										Optional:    true,
									},
									"container_read": {
										Description: "Indicates if the issued token has permission to read the " +
											"content, properties, metadata or block list of any blob in the " +
											"container. Use any blob in the container as the source of a copy operation.",
										Type:     schema.TypeBool,
										Optional: true,
									},
									"container_write": {
										Description: "Indicates that for any blob in the container if the issued " +
											"token has permission to create or write content, properties, metadata, " +
											"or block list. Snapshot or lease the blob. Resize the blob " +
											"(page blob only). Use the blob as the destination of a copy operation " +
											"within the same account.",
										Type:     schema.TypeBool,
										Optional: true,
									},
									"container_delete": {
										Description: "Indicates if issued token has permission to delete any blob in " +
											"the container.",
										Type:     schema.TypeBool,
										Optional: true,
									},
									"container_list": {
										Description: "Indicates if the issued token has permission to list blobs in the container.",
										Type:        schema.TypeBool,
										Optional:    true,
									},
								},
							},
						},
						"azure_sb": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Azure Storage Bus Addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"namespace": {
										Description: "Your Azure Service Bus namespace. Usually the first segment of " +
											"your Service Bus URL (for example `https://acme-org.servicebus.windows.net` " +
											"would be `acme-org`).",
										Type:     schema.TypeString,
										Optional: true,
									},
									"sas_key_name": {
										Description: "Your shared access policy name defined in your Service Bus entity.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"sas_key": {
										Description: "Primary Key associated with your shared access policy.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"entity_path": {
										Description: "Entity you want to request a token for, such as `my-queue`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"expiration": {
										Description:  "Optional expiration in minutes for the generated token. Defaults to 5 minutes.",
										Type:         schema.TypeInt,
										ValidateFunc: validation.IntAtLeast(0),
										Optional:     true,
									},
								},
							},
						},
						"rms": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							MaxItems:    1,
							Description: "Active Directory Rights Management Service SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Description: "URL of your Rights Management Server. It can be internal or " +
											"external, but users will have to be able to reach it.",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
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

func createClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	client := expandClient(d)
	if err := api.Client.Create(ctx, client); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(client.GetClientID())

	return readClient(ctx, d, m)
}

func readClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, d.Id())
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	result := multierror.Append(
		d.Set("client_id", client.GetClientID()),
		d.Set("client_secret", client.GetClientSecret()),
		d.Set("client_aliases", client.GetClientAliases()),
		d.Set("name", client.GetName()),
		d.Set("description", client.GetDescription()),
		d.Set("app_type", client.GetAppType()),
		d.Set("logo_uri", client.GetLogoURI()),
		d.Set("is_first_party", client.GetIsFirstParty()),
		d.Set("is_token_endpoint_ip_header_trusted", client.GetIsTokenEndpointIPHeaderTrusted()),
		d.Set("oidc_conformant", client.GetOIDCConformant()),
		d.Set("callbacks", client.GetCallbacks()),
		d.Set("allowed_logout_urls", client.GetAllowedLogoutURLs()),
		d.Set("allowed_origins", client.GetAllowedOrigins()),
		d.Set("allowed_clients", client.GetAllowedClients()),
		d.Set("grant_types", client.GetGrantTypes()),
		d.Set("organization_usage", client.GetOrganizationUsage()),
		d.Set("organization_require_behavior", client.GetOrganizationRequireBehavior()),
		d.Set("web_origins", client.GetWebOrigins()),
		d.Set("sso", client.GetSSO()),
		d.Set("sso_disabled", client.GetSSODisabled()),
		d.Set("cross_origin_auth", client.GetCrossOriginAuth()),
		d.Set("cross_origin_loc", client.GetCrossOriginLocation()),
		d.Set("custom_login_page_on", client.GetCustomLoginPageOn()),
		d.Set("custom_login_page", client.GetCustomLoginPage()),
		d.Set("form_template", client.GetFormTemplate()),
		d.Set("token_endpoint_auth_method", client.GetTokenEndpointAuthMethod()),
		d.Set("native_social_login", flattenCustomSocialConfiguration(client.GetNativeSocialLogin())),
		d.Set("jwt_configuration", flattenClientJwtConfiguration(client.GetJWTConfiguration())),
		d.Set("refresh_token", flattenClientRefreshTokenConfiguration(client.GetRefreshToken())),
		d.Set("encryption_key", client.GetEncryptionKey()),
		d.Set("addons", flattenClientAddons(client.Addons)),
		d.Set("mobile", flattenClientMobile(client.GetMobile())),
		d.Set("initiate_login_uri", client.GetInitiateLoginURI()),
		d.Set("signing_keys", client.SigningKeys),
		d.Set("client_metadata", client.ClientMetadata),
		d.Set("oidc_backchannel_logout_urls", client.OIDCBackchannelLogout.GetBackChannelLogoutURLs()),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	client := expandClient(d)
	if clientHasChange(client) {
		if err := api.Client.Update(ctx, d.Id(), client); err != nil {
			return diag.FromErr(err)
		}
	}

	d.Partial(true)
	if err := rotateClientSecret(ctx, d, m); err != nil {
		return diag.FromErr(err)
	}
	d.Partial(false)

	return readClient(ctx, d, m)
}

func deleteClient(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Client.Delete(ctx, d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
	}

	d.SetId("")
	return nil
}

func rotateClientSecret(ctx context.Context, d *schema.ResourceData, m interface{}) error {
	if !d.HasChange("client_secret_rotation_trigger") {
		return nil
	}

	api := m.(*config.Config).GetAPI()

	client, err := api.Client.RotateSecret(ctx, d.Id())
	if err != nil {
		return err
	}

	return d.Set("client_secret", client.GetClientSecret())
}
