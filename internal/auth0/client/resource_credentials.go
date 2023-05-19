package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewCredentialsResource will return a new auth0_client_credentials resource.
func NewCredentialsResource() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"client_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the client for which to configure the authentication method.",
			},
			"authentication_method": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"none",
					"client_secret_post",
					"client_secret_basic",
					"private_key_jwt",
				}, false),
				Description: "Configure the method to use when making requests to " +
					"any endpoint that requires this client to authenticate. " +
					"Options include `none` (public client without a client secret), " +
					"`client_secret_post` (confidential client using HTTP POST parameters), " +
					"`client_secret_basic` (confidential client using HTTP Basic), " +
					"`private_key_jwt` (confidential client using a Private Key JWT).",
			},
			"client_secret": {
				Type:      schema.TypeString,
				Optional:  true,
				Computed:  true,
				Sensitive: true,
				Description: "Secret for the client when using `client_secret_post` or `client_secret_basic` " +
					"authentication method. Keep this private. To access this attribute you need to add the " +
					"`read:client_keys` scope to the Terraform client. Otherwise, the attribute will contain an " +
					"empty string. The attribute will also be an empty string in case `private_key_jwt` is selected " +
					"as an authentication method.",
			},
			"private_key_jwt": {
				Type:        schema.TypeList,
				MaxItems:    1,
				Optional:    true,
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
										Description: "PEM-formatted public key (SPKI and PKCS1) or X509 certificate. Must be JSON escaped. " +
											"Changing this will force the credential to be recreated, resulting in a new client credential ID.",
									},
									"algorithm": {
										Type:         schema.TypeString,
										Optional:     true,
										ForceNew:     true,
										ValidateFunc: validation.StringInSlice([]string{"RS256", "RS384", "PS256"}, false),
										Default:      "RS256",
										Description: "Algorithm which will be used with the credential. " +
											"Can be one of `RS256`, `RS384`, `PS256`. If not specified, `RS256` will be used.",
									},
									"parse_expiry_from_cert": {
										Type:     schema.TypeBool,
										Optional: true,
										ForceNew: true,
										Description: "Parse expiry from x509 certificate. " +
											"If true, attempts to parse the expiry date from the provided PEM. " +
											"Changing this will force the credential to be recreated, resulting in a new client credential ID.",
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
										Description:  "The ISO 8601 formatted date representing the expiration of the credential.",
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
		Description: "With this resource, you can set up applications that use Auth0 for authentication " +
			"and configure allowed callback URLs and secrets for these applications.",
	}
}

func createClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)

	authenticationMethod := data.Get("authentication_method").(string)
	switch authenticationMethod {
	case "private_key_jwt":
		if diagnostics := createPrivateKeyJWTCredentials(api, data); diagnostics.HasError() {
			return diagnostics
		}
	case "client_secret_post", "client_secret_basic":
		if err := updateTokenEndpointAuthMethod(api, data); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}

		if err := updateSecret(api, data); err != nil {
			return diag.FromErr(err)
		}
	case "none":
		if err := updateTokenEndpointAuthMethod(api, data); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}
	}

	data.SetId(clientID)

	return readClientCredentials(ctx, data, meta)
}

func readClientCredentials(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()
	clientID := data.Get("client_id").(string)

	client, err := api.Client.Read(clientID)
	if err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	privateKeyJWT, err := flattenPrivateKeyJWT(api, data, client.GetClientAuthenticationMethods())
	if err != nil {
		return diag.FromErr(err)
	}

	result := multierror.Append(
		data.Set("authentication_method", flattenAuthenticationMethod(client)),
		data.Set("client_secret", client.GetClientSecret()),
		data.Set("private_key_jwt", privateKeyJWT),
	)

	return diag.FromErr(result.ErrorOrNil())
}

func updateClientCredentials(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	authenticationMethod := data.Get("authentication_method").(string)
	switch authenticationMethod {
	case "private_key_jwt":
		if diagnostics := modifyPrivateKeyJWTCredentials(api, data); diagnostics.HasError() {
			return diagnostics
		}
	case "client_secret_post", "client_secret_basic":
		if err := updateTokenEndpointAuthMethod(api, data); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}

		if err := updateSecret(api, data); err != nil {
			return diag.FromErr(err)
		}
	case "none":
		if err := updateTokenEndpointAuthMethod(api, data); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}
	}

	return readClientCredentials(ctx, data, meta)
}

func deleteClientCredentials(_ context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	clientID := data.Get("client_id").(string)

	authenticationMethod := data.Get("authentication_method").(string)
	if authenticationMethod == "private_key_jwt" {
		credentials, err := api.Client.ListCredentials(clientID)
		if err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}

		if err := detachCredentialsFromClient(api, clientID); err != nil {
			return diag.FromErr(err)
		}

		for _, credential := range credentials {
			if err := api.Client.DeleteCredential(clientID, credential.GetID()); err != nil {
				return diag.FromErr(err)
			}
		}

		return nil
	}

	data.SetId("")

	return nil
}

func createPrivateKeyJWTCredentials(api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	credentials, diagnostics := expandPrivateKeyJWT(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	credentialsToAttach := make([]management.Credential, 0)
	for _, credential := range credentials {
		if err := api.Client.CreateCredential(clientID, credential); err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				data.SetId("")
				return nil
			}

			return diag.FromErr(err)
		}

		credentialsToAttach = append(credentialsToAttach, management.Credential{
			ID: credential.ID,
		})
	}

	if err := attachCredentialsToClient(api, clientID, credentialsToAttach); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func modifyPrivateKeyJWTCredentials(api *management.Management, data *schema.ResourceData) diag.Diagnostics {
	credentials, diagnostics := expandPrivateKeyJWT(data.GetRawConfig())
	if diagnostics.HasError() {
		return diagnostics
	}

	clientID := data.Get("client_id").(string)

	// Check that client exists.
	if _, err := api.Client.Read(clientID, management.IncludeFields("client_id")); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			data.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	for index, credential := range credentials {
		const configAddress = "private_key_jwt.0.credentials"
		if !data.HasChange(fmt.Sprintf("%s.%s", configAddress, strconv.Itoa(index))) {
			continue
		}

		credentialID := data.Get(fmt.Sprintf("%s.%s.id", configAddress, strconv.Itoa(index))).(string)
		stateExpiresAt := data.Get(fmt.Sprintf("%s.%s.expires_at", configAddress, strconv.Itoa(index))).(string)
		if stateExpiresAt == "" {
			continue
		}

		// We can ignore the error as we have the validation.IsRFC3339Time on this attribute.
		expiresAt, _ := time.Parse(time.RFC3339, stateExpiresAt)
		credential.ExpiresAt = &expiresAt

		// Limitation: Unable to update the credential to never expire. Needs to get deleted and recreated if needed.
		if err := api.Client.UpdateCredential(clientID, credentialID, credential); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func attachCredentialsToClient(api *management.Management, clientID string, credentials []management.Credential) error {
	var client = struct {
		ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
		TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
	}{
		ClientAuthenticationMethods: &management.ClientAuthenticationMethods{
			PrivateKeyJWT: &management.PrivateKeyJWT{
				Credentials: &credentials,
			},
		},
		TokenEndpointAuthMethod: nil,
	}

	request, err := api.NewRequest(http.MethodPatch, api.URI("clients", clientID), &client)
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

func detachCredentialsFromClient(api *management.Management, clientID string) error {
	var client = struct {
		ClientAuthenticationMethods *management.ClientAuthenticationMethods `json:"client_authentication_methods"`
		TokenEndpointAuthMethod     *string                                 `json:"token_endpoint_auth_method"`
	}{
		ClientAuthenticationMethods: nil,
		TokenEndpointAuthMethod:     auth0.String("client_secret_post"),
	}

	request, err := api.NewRequest(http.MethodPatch, api.URI("clients", clientID), &client)
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

func updateTokenEndpointAuthMethod(api *management.Management, data *schema.ResourceData) error {
	if !data.HasChange("authentication_method") {
		return nil
	}

	clientID := data.Get("client_id").(string)
	tokenEndpointAuthMethod := data.Get("authentication_method").(string)

	return api.Client.Update(clientID, &management.Client{
		TokenEndpointAuthMethod: &tokenEndpointAuthMethod,
	})
}

func updateSecret(api *management.Management, data *schema.ResourceData) error {
	if !data.HasChange("client_secret") {
		return nil
	}

	clientID := data.Get("client_id").(string)
	clientSecret := data.Get("client_secret").(string)

	return api.Client.Update(clientID, &management.Client{
		ClientSecret: &clientSecret,
	})
}

func expandPrivateKeyJWT(rawConfig cty.Value) ([]*management.Credential, diag.Diagnostics) {
	credentials := make([]*management.Credential, 0)

	rawConfig.GetAttr("private_key_jwt").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
		config.GetAttr("credentials").ForEachElement(func(_ cty.Value, config cty.Value) (stop bool) {
			credentials = append(credentials, expandClientCredentials(config))
			return stop
		})
		return stop
	})

	if len(credentials) == 0 {
		return nil, diag.Diagnostics{
			diag.Diagnostic{
				Severity:      diag.Error,
				Summary:       "Client Credentials Missing",
				Detail:        "You must define client credentials when setting the authentication method as Private Key JWT.",
				AttributePath: cty.Path{cty.GetAttrStep{Name: "private_key_jwt.credentials"}},
			},
		}
	}

	return credentials, nil
}

func expandClientCredentials(rawConfig cty.Value) *management.Credential {
	clientCredential := management.Credential{
		Name:                value.String(rawConfig.GetAttr("name")),
		CredentialType:      value.String(rawConfig.GetAttr("credential_type")),
		PEM:                 value.String(rawConfig.GetAttr("pem")),
		Algorithm:           value.String(rawConfig.GetAttr("algorithm")),
		ParseExpiryFromCert: value.Bool(rawConfig.GetAttr("parse_expiry_from_cert")),
	}

	if expiresAt := value.String(rawConfig.GetAttr("expires_at")); expiresAt != nil {
		// We can ignore the error as we have the validation.IsRFC3339Time on this attribute.
		expiresAt, _ := time.Parse(time.RFC3339, *expiresAt)
		clientCredential.ExpiresAt = &expiresAt
	}

	return &clientCredential
}

func flattenAuthenticationMethod(client *management.Client) string {
	if client.GetTokenEndpointAuthMethod() == "" && client.GetClientAuthenticationMethods() == nil {
		switch client.GetAppType() {
		case "native", "spa":
			return "none"
		case "regular_web", "non_interactive":
			return "client_secret_post"
		default:
			return "client_secret_basic"
		}
	}

	if client.GetTokenEndpointAuthMethod() != "" {
		return client.GetTokenEndpointAuthMethod()
	}

	return "private_key_jwt"
}

func flattenPrivateKeyJWT(
	api *management.Management,
	data *schema.ResourceData,
	clientAuthMethods *management.ClientAuthenticationMethods,
) ([]interface{}, error) {
	if clientAuthMethods == nil {
		return nil, nil
	}

	const timeRFC3339WithMilliseconds = "2006-01-02T15:04:05.000Z07:00"

	stateCredentials := make([]interface{}, 0)
	for index, credential := range clientAuthMethods.GetPrivateKeyJWT().GetCredentials() {
		credential, err := api.Client.GetCredential(data.Id(), credential.GetID())
		if err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				return nil, nil
			}

			return nil, err
		}

		stateCredentials = append(stateCredentials, map[string]interface{}{
			"id":              credential.GetID(),
			"name":            credential.GetName(),
			"key_id":          credential.GetKeyID(),
			"credential_type": credential.GetCredentialType(),
			"pem": data.Get(
				fmt.Sprintf("private_key_jwt.0.credentials.%s.pem", strconv.Itoa(index)),
			), // Doesn't get read back.
			"algorithm": credential.GetAlgorithm(),
			"parse_expiry_from_cert": data.Get(
				fmt.Sprintf("private_key_jwt.0.credentials.%s.parse_expiry_from_cert", strconv.Itoa(index)),
			), // Doesn't get read back.
			"created_at": credential.GetCreatedAt().Format(timeRFC3339WithMilliseconds),
			"updated_at": credential.GetUpdatedAt().Format(timeRFC3339WithMilliseconds),
			"expires_at": credential.GetExpiresAt().Format(timeRFC3339WithMilliseconds),
		})
	}

	return []interface{}{
		map[string]interface{}{
			"credentials": stateCredentials,
		},
	}, nil
}
