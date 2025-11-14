package client

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	internalValidation "github.com/auth0/terraform-provider-auth0/internal/validation"
)

// ValidAppTypes contains all valid values for client app_type.
var ValidAppTypes = []string{
	"native", "spa", "regular_web", "non_interactive", "resource_server", "rms",
	"box", "cloudbees", "concur", "dropbox", "mscrm", "echosign",
	"egnyte", "newrelic", "office365", "salesforce", "sentry",
	"sharepoint", "slack", "springcm", "sso_integration", "zendesk", "zoom", "express_configuration",
}

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
			"client_aliases": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "List of audiences/realms for SAML protocol. Used by the wsfed addon.",
			},
			"app_type": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice(ValidAppTypes, false),
				Description: "Type of application the client represents. Possible values are: `native`, `spa`, " +
					"`regular_web`, `non_interactive`, `resource_server`,`sso_integration`. Specific SSO integrations types accepted " +
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
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Indicates whether this client is a first-party client." +
					"Defaults to true from the API",
			},
			"is_token_endpoint_ip_header_trusted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				Description: "Indicates whether the token endpoint IP header is trusted. Requires the authentication " +
					"method to be set to `client_secret_post` or `client_secret_basic`. Setting this property when " +
					"creating the resource, will default the authentication method to `client_secret_post`. To change " +
					"the authentication method to `client_secret_basic` use the `auth0_client_credentials` resource.",
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
				Computed:    true,
				Description: "Set of URLs that are valid to call back from Auth0 for OIDC backchannel logout. Currently only one URL is allowed.",
				Deprecated: "This resource is deprecated and will be removed in the next major version. " +
					"Please use `oidc_logout` for managing OIDC backchannel logout URLs.",
			},
			"grant_types": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Optional:    true,
				Description: "Types of grants that this client is authorized to use.",
			},
			"async_approval_notification_channels": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"guardian-push",
						"email",
					}, false),
				},
				Optional: true,
				Computed: true,
				Description: "List of notification channels enabled for CIBA (Client-Initiated Backchannel Authentication) requests initiated by this client. " +
					"Valid values are `guardian-push` and `email`. The order is significant as this is the order in which notification channels will be evaluated. " +
					"Defaults to `[\"guardian-push\"]` if not specified.",
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
					"no_prompt", "pre_login_prompt", "post_login_prompt",
				}, false),
				Description: "Defines how to proceed during an authentication transaction when " +
					"`organization_usage = \"require\"`. Can be `no_prompt` (default), `pre_login_prompt` or  `post_login_prompt`.",
			},
			"organization_discovery_methods": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice([]string{
						"email", "organization_name",
					}, false),
				},
				Optional: true,
				Description: "Methods for discovering organizations during the pre_login_prompt. " +
					"Can include `email` (allows users to find their organization by entering their email address) " +
					"and/or `organization_name` (requires users to enter the organization name directly). " +
					"These methods can be combined. Setting this property requires that " +
					"`organization_require_behavior` is set to `pre_login_prompt`.",
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
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.StringInSlice([]string{
								"HS256",
								"RS256",
								"PS256",
							}, false),
							Description: "Algorithm used to sign JWTs. " +
								"Can be one of `HS256`, `RS256`, `PS256`.",
						},
					},
				},
			},
			"encryption_key": {
				Type:        schema.TypeMap,
				Optional:    true,
				Default:     nil,
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
					"or it is not allowed to make such requests (`false`).",
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
			"client_metadata": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     schema.TypeString,
				Description: "Metadata associated with the client, in the form of an object with string values " +
					"(max 255 chars). Maximum of 10 metadata properties allowed. Field names (max 255 chars) are " +
					"alphanumeric and may only include the following special characters: " +
					"`:,-+=_*?\"/\\()<>@ [Tab] [Space]`.",
			},
			"require_pushed_authorization_requests": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Makes the use of Pushed Authorization Requests mandatory for this client. This feature currently needs to be enabled on the tenant in order to make use of it.",
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
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							Description:  "Configuration settings for Android native apps.",
							AtLeastOneOf: []string{"mobile.0.ios"},
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
							Type:         schema.TypeList,
							Optional:     true,
							Computed:     true,
							MaxItems:     1,
							Description:  "Configuration settings for i0S native apps.",
							AtLeastOneOf: []string{"mobile.0.android"},
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
						"google": {
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
						"policies": {
							Type:     schema.TypeSet,
							Optional: true,
							Description: "A collection of policies governing multi-resource refresh token exchange " +
								"(MRRT), defining how refresh tokens can be used across different resource servers",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"audience": {
										Type:     schema.TypeString,
										Required: true,
										Description: "The identifier of the resource server to which the Multi " +
											"Resource Refresh Token Policy applies",
									},
									"scope": {
										Type:     schema.TypeList,
										Elem:     &schema.Schema{Type: schema.TypeString},
										Required: true,
										Description: "The resource server permissions granted under the Multi " +
											"Resource Refresh Token Policy, defining the context in which an " +
											"access token can be used",
									},
								},
							},
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
						"mscrm": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Microsoft Dynamics CRM SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Description:  "Microsoft Dynamics CRM application URL.",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
									},
								},
							},
						},
						"slack": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "Slack team or workspace name usually first segment in your Slack URL, " +
								"for example `https://acme-org.slack.com` would be `acme-org`.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"team": {
										Description: "Slack team name.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"sentry": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Sentry SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"org_slug": {
										Description: "Generated slug for your Sentry organization. Found in your " +
											"Sentry URL, for example `https://sentry.acme.com/acme-org/` would be " +
											"`acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
									"base_url": {
										Description:  "URL prefix only if running Sentry Community Edition, otherwise leave empty.",
										Type:         schema.TypeString,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
										Optional:     true,
									},
								},
							},
						},
						"echosign": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Adobe EchoSign SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Description: "Your custom domain found in your EchoSign URL, for example " +
											"`https://acme-org.echosign.com` would be `acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"egnyte": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Egnyte SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Description: "Your custom domain found in your Egnyte URL, for example " +
											"`https://acme-org.echosign.com` would be `acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"firebase": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Google Firebase addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"secret": {
										Description: "Google Firebase Secret. (SDK v2 only).",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"private_key_id": {
										Description: "Optional ID of the private key to obtain the `kid` header " +
											"claim from the issued token (SDK v3+ tokens only).",
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
									},
									"private_key": {
										Description: "Private Key for signing the token (SDK v3+ tokens only).",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"client_email": {
										Description: "ID of the Service Account you have created (shown as " +
											"`client_email` in the generated JSON file, SDK v3+ tokens only).",
										Type:     schema.TypeString,
										Optional: true,
									},
									"lifetime_in_seconds": {
										Description: "Optional expiration in seconds for the generated token. " +
											"Defaults to 3600 seconds (SDK v3+ tokens only).",
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
						},
						"newrelic": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "New Relic SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account": {
										Description: "Your New Relic Account ID found in your New Relic URL after the " +
											"`/accounts/` path, for example `https://rpm.newrelic.com/accounts/123456/query` would be `123456`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"office365": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Microsoft Office 365 SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"domain": {
										Description: "Your Office 365 domain name, for example `acme-org.com`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"connection": {
										Description: "Optional Auth0 database connection for testing an " +
											"already-configured Office 365 tenant.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"salesforce": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Salesforce SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"entity_id": {
										Description:  "Arbitrary logical URL that identifies the Saleforce resource, for example `https://acme-org.com`.",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
									},
								},
							},
						},
						"salesforce_api": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Salesforce API addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_id": {
										Description: "Consumer Key assigned by Salesforce to the Connected App.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"principal": {
										Description: "Name of the property in the user object that maps to a " +
											"Salesforce username, for example `email`.",
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
									},
									"community_name": {
										Description: "Community name.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"community_url_section": {
										Description: "Community URL section.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"salesforce_sandbox_api": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Salesforce Sandbox addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_id": {
										Description: "Consumer Key assigned by Salesforce to the Connected App.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"principal": {
										Description: "Name of the property in the user object that maps to a " +
											"Salesforce username, for example `email`.",
										Type:      schema.TypeString,
										Optional:  true,
										Sensitive: true,
									},
									"community_name": {
										Description: "Community name.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"community_url_section": {
										Description: "Community URL section.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"layer": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Layer addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"provider_id": {
										Description: "Provider ID of your Layer account.",
										Type:        schema.TypeString,
										Required:    true,
									},
									"key_id": {
										Description: "Authentication Key identifier used to sign the Layer token.",
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
									},
									"private_key": {
										Description: "Private key for signing the Layer token.",
										Type:        schema.TypeString,
										Required:    true,
										Sensitive:   true,
									},
									"principal": {
										Description: "Name of the property used as the unique user ID in Layer. " +
											"If not specified `user_id` is used.",
										Type:     schema.TypeString,
										Optional: true,
									},
									"expiration": {
										Description: "Optional expiration in minutes for the generated token. " +
											"Defaults to 5 minutes.",
										Type:         schema.TypeInt,
										Optional:     true,
										ValidateFunc: validation.IntAtLeast(0),
									},
								},
							},
						},
						"sap_api": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "SAP API addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_id": {
										Description: "If activated in the OAuth 2.0 client configuration (transaction `SOAUTH2) " +
											"the SAML attribute `client_id` must be set and equal the `client_id` form " +
											"parameter of the access token request.",
										Type:     schema.TypeString,
										Optional: true,
									},
									"username_attribute": {
										Description: "Name of the property in the user object that maps to a SAP username, for example `email`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"token_endpoint_url": {
										Description:  "The OAuth2 token endpoint URL of your SAP OData server.",
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: internalValidation.IsURLWithHTTPSorEmptyString,
									},
									"scope": {
										Description: "Requested scope for SAP APIs.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"service_password": {
										Description: "Service account password to use to authenticate API calls to the token endpoint.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
									"name_identifier_format": {
										Description: "NameID element of the Subject which can be used to express the user's identity. " +
											"Defaults to `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"sharepoint": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "SharePoint SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"url": {
										Description: "Internal SharePoint application URL.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"external_url": {
										Description: "External SharePoint application URLs if exposed to the Internet.",
										Type:        schema.TypeList,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional: true,
									},
								},
							},
						},
						"springcm": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "SpringCM SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"acs_url": {
										Description: "SpringCM ACS URL, for example `https://na11.springcm.com/atlas/sso/SSOEndpoint.ashx`.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"wams": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Windows Azure Mobile Services addon configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"master_key": {
										Description: "Your master key for Windows Azure Mobile Services.",
										Type:        schema.TypeString,
										Optional:    true,
										Sensitive:   true,
									},
								},
							},
						},
						"zendesk": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Zendesk SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account_name": {
										Description: "Zendesk account name. Usually the first segment in your Zendesk URL, " +
											"for example `https://acme-org.zendesk.com` would be `acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"zoom": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Zoom SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"account": {
										Description: "Zoom account name. Usually the first segment of your Zoom URL, for " +
											"example `https://acme-org.zoom.us` would be `acme-org`.",
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"sso_integration": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Generic SSO configuration.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Description: "SSO integration name.",
										Type:        schema.TypeString,
										Optional:    true,
									},
									"version": {
										Description: "SSO integration version installed.",
										Type:        schema.TypeString,
										Optional:    true,
									},
								},
							},
						},
						"samlp": {
							Type:        schema.TypeList,
							MaxItems:    1,
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
										Type:          schema.TypeMap,
										Optional:      true,
										Elem:          schema.TypeString,
										ConflictsWith: []string{"addons.0.samlp.0.flexible_mappings"},
										Description: "Mappings between the Auth0 user profile property " +
											"name (`name`) and the output attributes on the SAML " +
											"attribute in the assertion (`value`).",
									},
									"flexible_mappings": {
										Type:             schema.TypeString,
										Optional:         true,
										ValidateFunc:     validation.StringIsJSON,
										ConflictsWith:    []string{"addons.0.samlp.0.mappings"},
										DiffSuppressFunc: structure.SuppressJsonDiff,
										Description: "This is a supporting attribute to `mappings` field." +
											"Please note this is an experimental field. " + "" +
											"It should only be used when needed to send a map with keys as slices.",
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
										Type:     schema.TypeInt,
										Optional: true,
										Default:  3600,
										Description: "Number of seconds during which the token is valid. " +
											"Defaults to `3600` seconds.",
									},
									"sign_response": {
										Type:     schema.TypeBool,
										Optional: true,
										Description: "Indicates whether or not the SAML Response should be signed " +
											"instead of the SAML Assertion.",
									},
									"name_identifier_format": {
										Type:     schema.TypeString,
										Optional: true,
										Default:  "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified",
										Description: "Format of the name identifier. " +
											"Defaults to `urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified`.",
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
											"and `xs:anyType`. When set to `false`, all `xs:type` are `xs:anyType`. " +
											"Defaults to `true`.",
									},
									"include_attribute_name_format": {
										Type:     schema.TypeBool,
										Optional: true,
										Default:  true,
										Description: "Indicates whether or not we should infer the NameFormat " +
											"based on the attribute name. If set to `false`, the attribute " +
											"NameFormat is not set in the assertion. Defaults to `true`.",
									},
									"logout": {
										Type:        schema.TypeList,
										MaxItems:    1,
										Optional:    true,
										Description: "Configuration settings for logout.",
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"callback": {
													Description: "The service provider (client application)'s Single Logout Service URL, " +
														"where Auth0 will send logout requests and responses.",
													Type:     schema.TypeString,
													Optional: true,
												},
												"slo_enabled": {
													Description: "Controls whether Auth0 should notify service providers of session termination.",
													Type:        schema.TypeBool,
													Optional:    true,
												},
											},
										},
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
									"issuer": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Issuer of the SAML Assertion.",
									},
								},
							},
						},
						"box": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Box SSO indicator (no configuration settings needed for Box SSO).",
							Elem:        &schema.Resource{},
						},
						"cloudbees": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "CloudBees SSO indicator (no configuration settings needed for CloudBees SSO).",
							Elem:        &schema.Resource{},
						},
						"concur": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Concur SSO indicator (no configuration settings needed for Concur SSO).",
							Elem:        &schema.Resource{},
						},
						"dropbox": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Dropbox SSO indicator (no configuration settings needed for Dropbox SSO).",
							Elem:        &schema.Resource{},
						},
						"wsfed": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Description: "WS-Fed (WIF) addon indicator. Actual configuration is stored in `callback` " +
								"and `client_aliases` properties on the client.",
							Elem: &schema.Resource{},
						},
					},
				},
			},
			"default_organization": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Computed:    true,
				Description: "Configure and associate an organization with the Client",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"flows": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Definition of the flow that needs to be configured. Eg. client_credentials",
						},
						"organization_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "The unique identifier of the organization",
						},
						"disable": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "If set, the `default_organization` will be removed.",
						},
					},
				},
			},
			"token_exchange": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Allows configuration for token exchange",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allow_any_profile_of_type": {
							Required:    true,
							Type:        schema.TypeList,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of allowed profile types for token exchange",
						},
					},
				},
			},
			"compliance_level": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"none", "fapi1_adv_pkj_par", "fapi1_adv_mtls_par"}, false),
				Default:      nil,
				Description: "Defines the compliance level for this client, which may restrict it's capabilities. " +
					"Can be one of `none`, `fapi1_adv_pkj_par`, `fapi1_adv_mtls_par`.",
			},
			"require_proof_of_possession": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Makes the use of Proof-of-Possession mandatory for this client.",
			},
			"oidc_logout": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Configure OIDC logout for the Client",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"backchannel_logout_urls": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
							Required:    true,
							Description: "Set of URLs that are valid to call back from Auth0 for OIDC backchannel logout. Currently only one URL is allowed.",
						},
						"backchannel_logout_initiators": {
							Type:        schema.TypeList,
							Optional:    true,
							MaxItems:    1,
							Description: "Configure OIDC logout initiators for the Client",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"mode": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"all", "custom"}, false),
										Description:  "Determines the configuration method for enabling initiators. `custom` enables only the initiators listed in the backchannel_logout_selected_initiators set, `all` enables all current and future initiators.",
									},
									"selected_initiators": {
										Type: schema.TypeSet,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
										Optional:    true,
										Description: "Contains the list of initiators to be enabled for the given client.",
									},
								},
							},
						},
					},
				},
			},
			"session_transfer": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"can_create_session_transfer_token": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the application(Native app) can use the Token Exchange endpoint to create a session_transfer_token",
						},
						"allowed_authentication_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Computed: true,
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								Description:  "Can be either `cookie` or `query` or both.",
								ValidateFunc: validation.StringInSlice([]string{"cookie", "query"}, false),
							},
						},
						"enforce_device_binding": {
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
							Description: "Configures the level of device binding enforced when a session_transfer_token is consumed. " +
								"Can be one of `ip`, `asn` or `none`.",
							ValidateFunc: validation.StringInSlice([]string{"ip", "asn", "none"}, false),
						},
						"allow_refresh_token": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether the application is allowed to use a refresh token when using a session_transfer_token session.",
						},
						"enforce_cascade_revocation": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether Refresh Tokens created during a native-to-web session are tied to that session's lifetime. This determines if such refresh tokens should be automatically revoked when their corresponding sessions are.",
						},
						"enforce_online_refresh_tokens": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Indicates whether revoking the parent Refresh Token that initiated a Native to Web flow and was used to issue a Session Transfer Token should trigger a cascade revocation affecting its dependent child entities.",
						},
					},
				},
			},
			"token_quota": commons.TokenQuotaSchema(),
			"skip_non_verifiable_callback_uri_confirmation_prompt": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Indicates whether the confirmation prompt appears when using non-verifiable callback URIs. Set to true to skip the prompt, false to show it, or null to unset. Accepts (true/false/null) or (\"true\"/\"false\"/\"null\") ",
				ValidateFunc: validation.StringInSlice([]string{"true", "false", "null"}, false),
				DiffSuppressFunc: func(_, o, n string, _ *schema.ResourceData) bool {
					return (o == "null" && n == "") || o == n
				},
			},
			"resource_server_identifier": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The identifier of a resource server that client is associated with" +
					"This property can be sent only when app_type=resource_server." +
					"This property can not be changed, once the client is created.",
			},
			"express_configuration": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				MaxItems:    1,
				Description: "Express Configuration settings for the client. Used with OIN Express Configuration.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"initiate_login_uri_template": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The URI users should bookmark to log in to this application. Variable substitution is permitted for: organization_name, organization_id, and connection_name.",
						},
						"user_attribute_profile_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the user attribute profile to use for this application.",
						},
						"connection_profile_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the connection profile to use for this application.",
						},
						"enable_client": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "When true, all connections made via express configuration will be enabled for this application.",
						},
						"enable_organization": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "When true, all connections made via express configuration will have the associated organization enabled.",
						},
						"okta_oin_client_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier for the Okta OIN Express Configuration Client.",
						},
						"admin_login_domain": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The domain that admins are expected to log in via for authenticating for express configuration.",
						},
						"oin_submission_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The identifier of the published application in the OKTA OIN.",
						},
						"linked_clients": {
							Type:        schema.TypeList,
							Optional:    true,
							Computed:    true,
							Description: "List of client IDs that are linked to this express configuration (e.g. web or mobile clients).",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"client_id": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "The ID of the linked client.",
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

func createClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	client, err := expandClient(data)

	if err != nil {
		return diag.FromErr(err)
	}

	if err := api.Client.Create(ctx, client); err != nil {
		return diag.FromErr(err)
	}

	time.Sleep(800 * time.Millisecond)

	data.SetId(client.GetClientID())
	return readClient(ctx, data, meta)
}

func readClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	err = flattenClient(data, client)
	return diag.FromErr(err)
}

func updateClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	client, err := expandClient(data)
	if err != nil {
		return diag.FromErr(err)
	}

	nullFields := fetchNullableFields(data, client)
	if len(nullFields) != 0 {
		if err := api.Request(ctx, http.MethodPatch, api.URI("clients", data.Id()), nullFields); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := api.Client.Update(ctx, data.Id(), client); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	time.Sleep(200 * time.Millisecond)

	return readClient(ctx, data, meta)
}

func deleteClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Client.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
