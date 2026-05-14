package client

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func privateKeyJWTCredentialSetHash(v interface{}) int {
	m := v.(map[string]interface{})

	var buf strings.Builder
	buf.WriteString(m["credential_type"].(string))
	buf.WriteString(m["pem"].(string))

	if algo, ok := m["algorithm"].(string); ok && algo != "" {
		buf.WriteString(algo)
	} else {
		buf.WriteString("RS256")
	}

	if name, ok := m["name"].(string); ok && name != "" {
		buf.WriteString(name)
	}

	if expiresAt, ok := m["expires_at"].(string); ok && expiresAt != "" {
		buf.WriteString(expiresAt)
	}

	return schema.HashString(buf.String())
}

// NewCredentialsResource will return a new auth0_client_credentials resource.
func NewCredentialsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the client for which to configure the authentication method.",
			},
			"authentication_method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					"client_secret_post",
					"client_secret_basic",
					"private_key_jwt",
					"tls_client_auth",
					"self_signed_tls_client_auth",
				}, false),
				Description: "Configure the method to use when making requests to " +
					"any endpoint that requires this client to authenticate. " +
					"Options include `none` (public client without a client secret), " +
					"`client_secret_post` (confidential client using HTTP POST parameters), " +
					"`client_secret_basic` (confidential client using HTTP Basic), " +
					"`private_key_jwt` (confidential client using a Private Key JWT), " +
					"`tls_client_auth` (confidential client using CA-based mTLS authentication), " +
					"`self_signed_tls_client_auth` (confidential client using mTLS authentication utilizing a self-signed certificate).",
			},
			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				ConflictsWith: []string{
					"private_key_jwt",
					"tls_client_auth",
					"self_signed_tls_client_auth",
				},
				Description: "Secret for the client when using `client_secret_post` or `client_secret_basic` " +
					"authentication method. Keep this private. To access this attribute you need to add either " +
					"`read:client_keys` or `read:client_credentials` scope to the Terraform client. Otherwise, the attribute will contain an " +
					"empty string. The attribute will also be an empty string in case `private_key_jwt` is selected " +
					"as an authentication method.",
			},
			"private_key_jwt": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ConflictsWith: []string{
					"client_secret",
					"tls_client_auth",
					"self_signed_tls_client_auth",
				},
				Description: "Defines `private_key_jwt` client authentication method.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeSet,
							Set:      privateKeyJWTCredentialSetHash,
							MaxItems: 2,
							Required: true,
							Description: "Client credentials available for use when Private Key JWT is in use as " +
								"the client authentication method. A maximum of 2 client credentials can be set.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the client credential.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Friendly name for a credential.",
									},
									"key_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The key identifier of the credential, generated on creation.",
									},
									"credential_type": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: validation.StringInSlice([]string{"public_key"}, false),
										Description:  "Credential type. Supported types: `public_key`.",
									},
									"pem": {
										Type:     schema.TypeString,
										Required: true,
										Description: "PEM-formatted public key (SPKI and PKCS1) or X509 certificate. " +
											"Must be JSON escaped.",
									},
									"algorithm": {
										Type:         schema.TypeString,
										Optional:     true,
										ValidateFunc: validation.StringInSlice([]string{"RS256", "RS384", "PS256"}, false),
										Default:      "RS256",
										Description: "Algorithm which will be used with the credential. " +
											"Can be one of `RS256`, `RS384`, `PS256`. If not specified, " +
											"`RS256` will be used.",
									},
									"parse_expiry_from_cert": {
										Type:     schema.TypeBool,
										Optional: true,
										Description: "Parse expiry from x509 certificate. " +
											"If true, attempts to parse the expiry date from the provided PEM. " +
											"If also the `expires_at` is set the credential expiry will be set to " +
											"the explicit `expires_at` value.",
									},
									"created_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was created.",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was updated.",
									},
									"expires_at": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsRFC3339Time,
										Description: "The ISO 8601 formatted date representing " +
											"the expiration of the credential. It is not possible to set this to " +
											"never expire after it has been set. Recreate the certificate if needed.",
									},
								},
							},
						},
					},
				},
			},
			"tls_client_auth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ConflictsWith: []string{
					"client_secret",
					"private_key_jwt",
					"self_signed_tls_client_auth",
				},
				Description: "Defines `tls_client_auth` client authentication method.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Credentials that will be enabled on the client for CA-based mTLS authentication.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the client credential.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Friendly name for a credential.",
									},
									"credential_type": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"cert_subject_dn"}, false),
										Description:  "Credential type. Supported types: `cert_subject_dn`.",
									},
									"subject_dn": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										Computed:     true,
										ValidateFunc: validation.StringLenBetween(1, 256),
										Description:  "Subject Distinguished Name. Mutually exlusive with `pem` property.",
									},
									"pem": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 4096),
										Description: "PEM-formatted X509 certificate. Must be JSON escaped. " +
											"Mutually exlusive with `subject_dn` property.",
									},
									"created_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was created.",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was updated.",
									},
								},
							},
						},
					},
				},
			},
			"self_signed_tls_client_auth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				ConflictsWith: []string{
					"client_secret",
					"private_key_jwt",
					"tls_client_auth",
				},
				Description: "Defines `tls_client_auth` client authentication method.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeList,
							Required: true,
							Description: "Credentials that will be enabled on the client for mTLS " +
								"authentication utilizing self-signed certificates.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the client credential.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Friendly name for a credential.",
									},
									"credential_type": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"x509_cert"}, false),
										Description:  "Credential type. Supported types: `x509_cert`.",
									},
									"pem": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringLenBetween(1, 4096),
										Description:  "PEM-formatted X509 certificate. Must be JSON escaped. ",
									},
									"thumbprint_sha256": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The X509 certificate's SHA256 thumbprint.",
									},
									"created_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was created.",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was updated.",
									},
									"expires_at": {
										Type:     schema.TypeString,
										Computed: true,
										Description: "The ISO 8601 formatted date representing " +
											"the expiration of the credential.",
									},
								},
							},
						},
					},
				},
			},
			"signed_request_object": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
				Description: "Configuration for JWT-secured Authorization Requests(JAR).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"required": {
							Type:        schema.TypeBool,
							Optional:    true,
							Computed:    true,
							Description: "Require JWT-secured authorization requests.",
						},
						"credentials": {
							Type:        schema.TypeList,
							Required:    true,
							Description: "Client credentials for use with JWT-secured authorization requests.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the client credential.",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										ForceNew:    true,
										Description: "Friendly name for a credential.",
									},
									"key_id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The key identifier of the credential, generated on creation.",
									},
									"credential_type": {
										Type:         schema.TypeString,
										Required:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"public_key"}, false),
										Description:  "Credential type. Supported types: `public_key`.",
									},
									"pem": {
										Type:     schema.TypeString,
										Required: true,
										ForceNew: true,
										Description: "PEM-formatted public key (SPKI and PKCS1) or X509 certificate. " +
											"Must be JSON escaped.",
									},
									"algorithm": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"RS256", "RS384", "PS256"}, false),
										Default:      "RS256",
										Description: "Algorithm which will be used with the credential. " +
											"Can be one of `RS256`, `RS384`, `PS256`. If not specified, " +
											"`RS256` will be used.",
									},
									"parse_expiry_from_cert": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Description: "Parse expiry from x509 certificate. " +
											"If true, attempts to parse the expiry date from the provided PEM. " +
											"If also the `expires_at` is set the credential expiry will be set to " +
											"the explicit `expires_at` value.",
									},
									"created_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was created.",
									},
									"updated_at": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ISO 8601 formatted date the credential was updated.",
									},
									"expires_at": {
										Type:         schema.TypeString,
										Optional:     true,
										Computed:     true,
										ValidateFunc: validation.IsRFC3339Time,
										Description: "The ISO 8601 formatted date representing " +
											"the expiration of the credential. It is not possible to set this to " +
											"never expire after it has been set. Recreate the certificate if needed.",
									},
								},
							},
						},
					},
				},
			},
		},
		CreateContext: createClientCredentials,
		ReadContext:   readClientCredentials,
		UpdateContext: updateClientCredentials,
		DeleteContext: deleteClientCredentials,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    credentialsResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: upgradeCredentialsResourceStateV0toV1,
				Version: 0,
			},
		},
		Description: "With this resource, you can configure the method to use when making requests to any endpoint " +
			"that requires this client to authenticate.",
	}
}

// credentialsResourceV0 returns the V0 schema (before TypeList->TypeSet migration).
func credentialsResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"authentication_method": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
			},
			"private_key_jwt": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeList,
							MaxItems: 2,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id":                     {Type: schema.TypeString, Computed: true},
									"name":                   {Type: schema.TypeString, Optional: true},
									"key_id":                 {Type: schema.TypeString, Computed: true},
									"credential_type":        {Type: schema.TypeString, Required: true},
									"pem":                    {Type: schema.TypeString, Required: true},
									"algorithm":              {Type: schema.TypeString, Optional: true, Default: "RS256"},
									"parse_expiry_from_cert": {Type: schema.TypeBool, Optional: true},
									"created_at":             {Type: schema.TypeString, Computed: true},
									"updated_at":             {Type: schema.TypeString, Computed: true},
									"expires_at":             {Type: schema.TypeString, Optional: true, Computed: true},
								},
							},
						},
					},
				},
			},
			"tls_client_auth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id":              {Type: schema.TypeString, Computed: true},
									"name":            {Type: schema.TypeString, Optional: true},
									"credential_type": {Type: schema.TypeString, Optional: true},
									"pem":             {Type: schema.TypeString, Optional: true},
									"subject_dn":      {Type: schema.TypeString, Optional: true},
									"created_at":      {Type: schema.TypeString, Computed: true},
									"updated_at":      {Type: schema.TypeString, Computed: true},
									"expires_at":      {Type: schema.TypeString, Optional: true, Computed: true},
								},
							},
						},
					},
				},
			},
			"self_signed_tls_client_auth": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id":                {Type: schema.TypeString, Computed: true},
									"name":              {Type: schema.TypeString, Optional: true},
									"credential_type":   {Type: schema.TypeString, Optional: true},
									"pem":               {Type: schema.TypeString, Required: true},
									"thumbprint_sha256": {Type: schema.TypeString, Computed: true},
									"created_at":        {Type: schema.TypeString, Computed: true},
									"updated_at":        {Type: schema.TypeString, Computed: true},
									"expires_at":        {Type: schema.TypeString, Computed: true},
								},
							},
						},
					},
				},
			},
			"signed_request_object": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"required": {Type: schema.TypeBool, Optional: true},
						"credentials": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id":                     {Type: schema.TypeString, Computed: true},
									"name":                   {Type: schema.TypeString, Optional: true},
									"key_id":                 {Type: schema.TypeString, Computed: true},
									"credential_type":        {Type: schema.TypeString, Required: true},
									"pem":                    {Type: schema.TypeString, Required: true},
									"algorithm":              {Type: schema.TypeString, Optional: true, Default: "RS256"},
									"parse_expiry_from_cert": {Type: schema.TypeBool, Optional: true},
									"created_at":             {Type: schema.TypeString, Computed: true},
									"updated_at":             {Type: schema.TypeString, Computed: true},
									"expires_at":             {Type: schema.TypeString, Optional: true, Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func upgradeCredentialsResourceStateV0toV1(
	_ context.Context,
	rawState map[string]interface{},
	_ interface{},
) (map[string]interface{}, error) {
	return rawState, nil
}
