package client

import (
	"context"
	"fmt"
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

// validateIPAddressOrCIDR accepts either a bare IPv4/IPv6 address or a CIDR
// range, which is what token_vault_privileged_access.ip_allowlist takes. The SDK
// ships a validator for each form but none for the union, so try both and only
// report an error when neither matches. The errors from the failed attempts are
// discarded in favour of a single message naming both accepted forms.
func validateIPAddressOrCIDR(i interface{}, k string) ([]string, []error) {
	if _, errs := validation.IsIPAddress(i, k); len(errs) == 0 {
		return nil, nil
	}

	if _, errs := validation.IsCIDR(i, k); len(errs) == 0 {
		return nil, nil
	}

	return nil, []error{fmt.Errorf(
		"expected %s to be a valid IPv4/IPv6 address or CIDR range, got %v", k, i,
	)}
}

// tokenVaultPrivilegedAccessMaxTotalScopes is the cap the API enforces across
// every grant on the block, as opposed to the per-grant scope count, which is
// unbounded on its own.
const tokenVaultPrivilegedAccessMaxTotalScopes = 20

// validateTokenVaultPrivilegedAccess enforces the two constraints on
// token_vault_privileged_access.grants that span more than one element and so
// cannot be expressed with MaxItems: the total scope count across all grants, and
// the absence of a repeated connection. Both are rejected by the API with a 400,
// and catching them at plan time turns a failed apply into a plan error.
func validateTokenVaultPrivilegedAccess(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	grants, ok := diff.Get("token_vault_privileged_access.0.grants").(*schema.Set)
	if !ok {
		return nil
	}

	totalScopes := 0
	seenConnections := make(map[string]bool, grants.Len())

	for _, raw := range grants.List() {
		grant, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}

		connection, _ := grant["connection"].(string)
		if seenConnections[connection] {
			return fmt.Errorf(
				"`token_vault_privileged_access.grants` must not repeat a connection, "+
					"but %q appears more than once", connection,
			)
		}
		seenConnections[connection] = true

		if scopes, ok := grant["scopes"].(*schema.Set); ok {
			totalScopes += scopes.Len()
		}
	}

	if totalScopes > tokenVaultPrivilegedAccessMaxTotalScopes {
		return fmt.Errorf(
			"`token_vault_privileged_access.grants` supports a maximum of %d scopes in "+
				"total across all grants, but %d were configured",
			tokenVaultPrivilegedAccessMaxTotalScopes, totalScopes,
		)
	}

	return nil
}

// NewCredentialsResource will return a new auth0_client_credentials resource.
func NewCredentialsResource() *schema.Resource {
	return &schema.Resource{
		CustomizeDiff: validateTokenVaultPrivilegedAccess,
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
					"client_secret_wo",
					"private_key_jwt",
					"tls_client_auth",
					"self_signed_tls_client_auth",
				},
				Description: "Secret for the client when using `client_secret_post` or `client_secret_basic` " +
					"authentication method. Keep this private. To access this attribute you need to add either " +
					"`read:client_keys` or `read:client_credentials` scope to the Terraform client. Otherwise, the attribute will contain an " +
					"empty string. The attribute will also be an empty string in case `private_key_jwt` is selected " +
					"as an authentication method. **Note:** For better security, consider using `client_secret_wo` instead.",
			},
			"client_secret_wo": {
				Type:      schema.TypeString,
				Optional:  true,
				WriteOnly: true,
				Sensitive: true,
				ConflictsWith: []string{
					"client_secret",
					"private_key_jwt",
					"tls_client_auth",
					"self_signed_tls_client_auth",
				},
				RequiredWith: []string{"client_secret_wo_version"},
				Description: "Secret for the client when using `client_secret_post` or `client_secret_basic` " +
					"authentication method (write-only). This value is **not** stored in Terraform state. " +
					"Bump `client_secret_wo_version` to rotate it. Requires Terraform 1.11+.",
			},
			"client_secret_wo_version": {
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"client_secret_wo"},
				ValidateFunc: validation.IntAtLeast(1),
				Description: "Version counter for `client_secret_wo`. Must be a positive integer (starting at `1`). " +
					"Increment this value to trigger a client secret change when using `client_secret_wo`.",
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
										Type:     schema.TypeString,
										Optional: true,
										// Computed because this one is Optional: the API records
										// x509_cert whether or not the configuration said so, and
										// without this the read writes a value the configuration does
										// not have, which ForceNew then reads as a removal and plans
										// a replacement on every run.
										Computed:     true,
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
			"token_vault_privileged_access": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Description: "Configures the client as a Token Vault privileged worker, allowing it to " +
					"request Token Vault tokens on behalf of other users. " +
					"This is an Early Access feature and must be enabled for your tenant.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"credentials": {
							Type:     schema.TypeSet,
							Set:      privateKeyJWTCredentialSetHash,
							MaxItems: 2,
							Required: true,
							Description: "Credentials the privileged worker may authenticate with. " +
								"A maximum of 2 client credentials can be set.",
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
						"ip_allowlist": {
							Type:     schema.TypeSet,
							Required: true,
							MaxItems: 10,
							Description: "Permitted IPv4 or IPv6 addresses, or CIDR ranges, from which the " +
								"privileged worker may request tokens. A maximum of 10 entries can be set. " +
								"Set to `[]` to clear the entries already configured.",
							Elem: &schema.Schema{
								Type:         schema.TypeString,
								ValidateFunc: validateIPAddressOrCIDR,
							},
						},
						"grants": {
							Type: schema.TypeSet,
							// Optional rather than Required, even though the API demands the
							// field on every write. helper/schema has no "required" flag for a
							// nested block, so it fakes one by deriving MinItems: 1, and that
							// makes declaring zero grants impossible. Since the API accepts an
							// empty array as the way to clear them, that has to stay
							// expressible. Omitting the block entirely is equivalent to
							// clearing it; the expander sends [] either way.
							Optional: true,
							MaxItems: 5,
							Description: "Pins the connections, and the scopes within them, that the privileged " +
								"worker may request tokens for. A maximum of 5 grants can be set, with a maximum " +
								"of 20 scopes in total across all of them. Omit every `grants` block to clear the " +
								"grants already configured.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"connection": {
										Type:     schema.TypeString,
										Required: true,
										Description: "Name of the connection the grant applies to. The connection " +
											"does not need to exist when the grant is configured; it is validated " +
											"at runtime.",
									},
									"scopes": {
										Type:        schema.TypeSet,
										Required:    true,
										MinItems:    1,
										Description: "Scopes the privileged worker may request on the connection.",
										Elem:        &schema.Schema{Type: schema.TypeString},
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
