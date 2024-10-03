package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

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
					"authentication method. Keep this private. To access this attribute you need to add the " +
					"`read:client_keys` scope to the Terraform client. Otherwise, the attribute will contain an " +
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
							Type:     schema.TypeList,
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
		Description: "With this resource, you can configure the method to use when making requests to any endpoint " +
			"that requires this client to authenticate.",
	}
}

func createClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)

	// Check that client exists.
	if _, err := api.Client.Read(ctx, clientID, management.IncludeFields("client_id")); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(clientID)

	authenticationMethod := data.Get("authentication_method").(string)
	if len(authenticationMethod) > 0 {
		switch authenticationMethod {
		case "private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth":
			if diagnostics := createAuthenticationMethodCredentials(ctx, api, data, authenticationMethod); diagnostics.HasError() {
				return diagnostics
			}
		case "client_secret_post", "client_secret_basic":
			if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}

			if err := updateSecret(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}
		case "none":
			if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
				return diag.FromErr(err)
			}
		}
	}
	if data.GetRawConfig().GetAttr("signed_request_object").LengthInt() > 0 {
		diagnostics := createSignedRequestObject(ctx, api, data)
		if diagnostics.HasError() {
			return diagnostics
		}
	}

	return readClientCredentials(ctx, data, meta)
}

func readClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id())
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return diag.FromErr(flattenClientCredentials(ctx, api, data, client))
}

func updateClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	// Check that client exists.
	if _, err := api.Client.Read(ctx, data.Id(), management.IncludeFields("client_id")); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	authenticationMethod := data.Get("authentication_method").(string)
	switch authenticationMethod {
	case "private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth":
		if diagnostics := modifyAuthenticationMethodCredentials(ctx, api, data, authenticationMethod); diagnostics.HasError() {
			return diagnostics
		}
	case "client_secret_post", "client_secret_basic":
		if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}

		if err := updateSecret(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}
	case "none":
		if err := updateTokenEndpointAuthMethod(ctx, api, data); err != nil {
			return diag.FromErr(err)
		}
	}
	if data.GetRawConfig().GetAttr("signed_request_object").LengthInt() > 0 {
		diagnostics := modifySignedRequestObject(ctx, api, data)
		if diagnostics.HasError() {
			return diagnostics
		}
	}

	return readClientCredentials(ctx, data, meta)
}

func deleteClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id(), management.IncludeFields("client_id", "app_type"))
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	tokenEndpointAuthMethod := ""
	switch client.GetAppType() {
	case "native", "spa":
		tokenEndpointAuthMethod = "none"
	case "regular_web", "non_interactive":
		tokenEndpointAuthMethod = "client_secret_post"
	default:
		tokenEndpointAuthMethod = "client_secret_basic"
	}

	credentials, err := api.Client.ListCredentials(ctx, client.GetClientID())
	if err != nil {
		return diag.FromErr(err)
	}

	if len(credentials) > 0 {
		if err := detachClientCredentials(ctx, api, client.GetClientID(), tokenEndpointAuthMethod); err != nil {
			return diag.FromErr(err)
		}

		for _, credential := range credentials {
			if err := api.Client.DeleteCredential(ctx, client.GetClientID(), credential.GetID()); err != nil {
				return diag.FromErr(err)
			}
		}

		return nil
	}

	if err := api.Client.Update(ctx, client.GetClientID(), &management.Client{
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	}); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func createAuthenticationMethodCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, authenticationMethod string) diag.Diagnostics {
	credentials, diagnostics := expandAuthenticationMethodCredentials(data.GetRawConfig(), authenticationMethod)
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	credentialsToAttach := make([]management.Credential, 0)
	for _, credential := range credentials {
		if err := api.Client.CreateCredential(ctx, clientID, credential); err != nil {
			return diag.FromErr(err)
		}

		credentialsToAttach = append(credentialsToAttach, management.Credential{
			ID: credential.ID,
		})
	}

	err := attachAuthenticationMethodCredentials(ctx, api, clientID, authenticationMethod, credentialsToAttach)

	return diag.FromErr(err)
}

func modifyAuthenticationMethodCredentials(ctx context.Context, api *management.Management, data *schema.ResourceData, authenticationMethod string) diag.Diagnostics {
	credentials, diagnostics := expandAuthenticationMethodCredentials(data.GetRawConfig(), authenticationMethod)
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	for index, credential := range credentials {
		configAddress := fmt.Sprintf("%s.0.credentials.%d", authenticationMethod, index)
		if !data.HasChange(configAddress) {
			continue
		}

		credentialID := data.Get(fmt.Sprintf("%s.id", configAddress)).(string)
		stateExpiresAt := data.Get(fmt.Sprintf("%s.expires_at", configAddress)).(string)
		if stateExpiresAt == "" {
			continue
		}

		// The error can be ignored, the schema validates the type.
		expiresAt, _ := time.Parse(time.RFC3339, stateExpiresAt)
		credential.ExpiresAt = &expiresAt

		// Limitation: Unable to update the credential to never expire. Needs to get deleted and recreated if needed.
		if err := api.Client.UpdateCredential(ctx, clientID, credentialID, credential); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func createSignedRequestObject(ctx context.Context, api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	signedRequestObject, diagnostics := expandSignedRequestObject(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	if signedRequestObject.GetCredentials() != nil {
		credentialsToAttach := make([]management.Credential, 0)
		for _, credential := range signedRequestObject.GetCredentials() {
			if err := api.Client.CreateCredential(ctx, clientID, &credential); err != nil {
				return diag.FromErr(err)
			}

			credentialsToAttach = append(credentialsToAttach, management.Credential{
				ID: credential.ID,
			})
		}

		return diag.FromErr(attachSignedRequestObjectCredentials(ctx, api, clientID, signedRequestObject.Required, credentialsToAttach))
	}

	return nil
}

func modifySignedRequestObject(ctx context.Context, api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	signedRequestObject, diagnostics := expandSignedRequestObject(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	if signedRequestObject.GetCredentials() != nil {
		for index, credential := range signedRequestObject.GetCredentials() {
			configAddress := fmt.Sprintf("signed_request_object.0.credentials.%d", index)
			if !data.HasChange(configAddress) {
				continue
			}

			credentialID := data.Get(fmt.Sprintf("%s.id", configAddress)).(string)
			stateExpiresAt := data.Get(fmt.Sprintf("%s.expires_at", configAddress)).(string)
			if stateExpiresAt == "" {
				continue
			}

			// The error can be ignored, the schema validates the type.
			expiresAt, _ := time.Parse(time.RFC3339, stateExpiresAt)
			credential.ExpiresAt = &expiresAt

			// Limitation: Unable to update the credential to never expire. Needs to get deleted and recreated if needed.
			if err := api.Client.UpdateCredential(ctx, clientID, credentialID, &credential); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if data.HasChange("signed_request_object.0.required") {
		return diag.FromErr(attachSignedRequestObjectNoCredentials(ctx, api, clientID, signedRequestObject.Required))
	}

	return nil
}

type clientWithAuthMethod struct {
	ID                          string                                  `json:"-"`
	ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
	TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
}

type clientWithSignedRequestObject struct {
	ID                  string                                `json:"-"`
	SignedRequestObject *management.ClientSignedRequestObject `json:"signed_request_object"`
}

type clientWithAuthMethodAndSignedRequestObject struct {
	ID                          string                                  `json:"-"`
	ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
	TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
	SignedRequestObject         *management.ClientSignedRequestObject   `json:"signed_request_object"`
}

func attachAuthenticationMethodCredentials(ctx context.Context, api *management.Management, clientID string, authenticationMethod string, credentials []management.Credential) error {
	client := clientWithAuthMethod{
		ID:                          clientID,
		ClientAuthenticationMethods: &management.ClientAuthenticationMethods{},
		TokenEndpointAuthMethod:     nil,
	}

	switch authenticationMethod {
	case "private_key_jwt":
		client.ClientAuthenticationMethods.PrivateKeyJWT = &management.PrivateKeyJWT{
			Credentials: &credentials,
		}
	case "tls_client_auth":
		client.ClientAuthenticationMethods.TLSClientAuth = &management.TLSClientAuth{
			Credentials: &credentials,
		}
	case "self_signed_tls_client_auth":
		client.ClientAuthenticationMethods.SelfSignedTLSClientAuth = &management.SelfSignedTLSClientAuth{
			Credentials: &credentials,
		}
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func attachSignedRequestObjectCredentials(ctx context.Context, api *management.Management, clientID string, required *bool, credentials []management.Credential) error {
	client := clientWithSignedRequestObject{
		ID: clientID,
		SignedRequestObject: &management.ClientSignedRequestObject{
			Required:    required,
			Credentials: &credentials,
		},
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func attachSignedRequestObjectNoCredentials(ctx context.Context, api *management.Management, clientID string, required *bool) error {
	client := clientWithSignedRequestObject{
		ID: clientID,
		SignedRequestObject: &management.ClientSignedRequestObject{
			Required: required,
		},
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func detachClientCredentials(ctx context.Context, api *management.Management, clientID, tokenEndpointAuthMethod string) error {
	client := clientWithAuthMethodAndSignedRequestObject{
		ID:                          clientID,
		SignedRequestObject:         nil,
		ClientAuthenticationMethods: nil,
		// API doesn't accept nil on both of these, so we temporarily set this to a default.
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	}

	return updateClientInternal(ctx, api, client.ID, client)
}

func updateClientInternal(ctx context.Context, api *management.Management, clientID string, client interface{}) error {
	request, err := api.NewRequest(ctx, http.MethodPatch, api.URI("clients", clientID), client)
	if err != nil {
		return err
	}

	response, err := api.Do(request)
	if err != nil {
		return err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode >= http.StatusBadRequest {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf("%s", string(body))
	}

	return nil
}

func updateTokenEndpointAuthMethod(ctx context.Context, api *management.Management, data *schema.ResourceData) error {
	if !data.HasChange("authentication_method") {
		return nil
	}

	clientID := data.Get("client_id").(string)
	tokenEndpointAuthMethod := data.Get("authentication_method").(string)

	return api.Client.Update(ctx, clientID, &management.Client{
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	})
}

func updateSecret(ctx context.Context, api *management.Management, data *schema.ResourceData) error {
	if !data.HasChange("client_secret") {
		return nil
	}

	clientID := data.Get("client_id").(string)
	clientSecret := data.Get("client_secret").(string)

	return api.Client.Update(ctx, clientID, &management.Client{
		ClientSecret: &clientSecret,
	})
}

func expandAuthenticationMethodCredentials(rawConfig cty.Value, authenticationMethod string) ([]*management.Credential, diag.Diagnostics) {
	credentials := make([]*management.Credential, 0)

	rawConfig.GetAttr(authenticationMethod).ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
			credentials = append(credentials, expandClientCredential(config))
			return stop
		})
		return stop
	})

	if len(credentials) == 0 {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Client Credentials Missing",
				Detail:        fmt.Sprintf("You must define client credentials when setting the authentication method as %q.", authenticationMethod),
				AttributePath: cty.Path{cty.GetAttrStep{Name: fmt.Sprintf("%s.credentials", authenticationMethod)}},
			},
		}
	} else if authenticationMethod == "tls_client_auth" {
		for _, credential := range credentials {
			if (credential.PEM != nil && credential.SubjectDN != nil) || (credential.PEM == nil && credential.SubjectDN == nil) {
				return nil, diag.Diagnostics{
					diag.Diagnostic{
						Severity:      diag.Error,
						Summary:       "Client Credentials Invalid",
						Detail:        fmt.Sprintf("Exactly one of pem and subject_dn must be set when setting the authentication method as %q.", authenticationMethod),
						AttributePath: cty.Path{cty.GetAttrStep{Name: fmt.Sprintf("%s.credentials", authenticationMethod)}},
					},
				}
			}
		}
	}

	return credentials, nil
}

func expandSignedRequestObject(rawConfig cty.Value) (*management.ClientSignedRequestObject, diag.Diagnostics) {
	signedRequestObjectConfig := rawConfig.GetAttr("signed_request_object")
	if signedRequestObjectConfig.IsNull() {
		return nil, nil
	}

	var signedRequestObject management.ClientSignedRequestObject

	signedRequestObjectConfig.ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		credentials := make([]management.Credential, 0)
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
			credentials = append(credentials, *expandClientCredential(config))
			return stop
		})
		signedRequestObject.Credentials = &credentials
		signedRequestObject.Required = value.Bool(config.GetAttr("required"))
		return stop
	})

	if signedRequestObject == (management.ClientSignedRequestObject{}) {
		return nil, nil
	}

	if len(*signedRequestObject.Credentials) == 0 {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Client Credentials Missing",
				Detail:        "You must define client credentials when using JWT-secured Authorization Requests.",
				AttributePath: cty.Path{cty.GetAttrStep{Name: "signed_request_object.credentials"}},
			},
		}
	}

	return &signedRequestObject, nil
}

func expandClientCredential(rawConfig cty.Value) *management.Credential {
	clientCredential := management.Credential{
		Name:           value.String(rawConfig.GetAttr("name")),
		CredentialType: value.String(rawConfig.GetAttr("credential_type")),
	}

	switch *clientCredential.CredentialType {
	case "public_key":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
		clientCredential.Algorithm = value.String(rawConfig.GetAttr("algorithm"))
		clientCredential.ParseExpiryFromCert = value.Bool(rawConfig.GetAttr("parse_expiry_from_cert"))
		clientCredential.ExpiresAt = value.Time(rawConfig.GetAttr("expires_at"))
	case "cert_subject_dn":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
		clientCredential.SubjectDN = value.String(rawConfig.GetAttr("subject_dn"))
	case "x509_cert":
		clientCredential.PEM = value.String(rawConfig.GetAttr("pem"))
	}

	return &clientCredential
}
