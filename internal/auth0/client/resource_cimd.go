package client

import (
	"context"
	"fmt"

	mgmtv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/commons"
	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewCIMDResource returns a new auth0_client_cimd resource.
func NewCIMDResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCIMDClient,
		ReadContext:   readCIMDClient,
		UpdateContext: updateCIMDClient,
		DeleteContext: deleteCIMDClient,
		Importer: &schema.ResourceImporter{
			StateContext: importCIMDClient,
		},
		Description: "With this resource, you can register an Auth0 client from a " +
			"Client ID Metadata Document (CIMD) URL. CIMD enables tenant admins to " +
			"onboard MCP agent clients by providing a URL to an externally-hosted " +
			"metadata document instead of using Dynamic Client Registration.\n\n" +
			"Requires the `client_id_metadata_document_supported` tenant setting to be enabled.",
		Schema: cimdClientSchema(),
	}
}

func cimdClientSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"external_client_id": {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: "The HTTPS URL of the Client ID Metadata Document. " +
				"Must include a path component (e.g. `https://app.example.com/client.json`). " +
				"This value is immutable after creation.",
		},
		"external_client_id_version": {
			Type:     schema.TypeInt,
			Optional: true,
			Description: "Version number for external_client_id metadata document changes. " +
				"Update this value to sync the client with the latest values of the json metadata document.",
		},
		"client_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The ID of the client.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the client, derived from the CIMD metadata document.",
		},
		"callbacks": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Description: "URLs that Auth0 may call back after authentication. " +
				"Derived from the CIMD metadata document.",
		},
		"logo_uri": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "URL of the logo for this client, derived from the CIMD metadata document.",
		},
		"is_first_party": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether this is a first-party client. Always `false` for CIMD clients.",
		},
		"signing_keys": {
			Type:      schema.TypeList,
			Computed:  true,
			Sensitive: true,
			Elem:      &schema.Schema{Type: schema.TypeMap},
			Description: "List containing a map of the public cert of the signing key and the public cert " +
				"of the signing key in PKCS7.",
		},
		"external_metadata_type": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Type of external metadata. Always `cimd` for CIMD-registered clients.",
		},
		"external_metadata_created_by": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "Who created the external metadata client: `admin` (via Management API) " +
				"or `client` (self-registered).",
		},
		"jwks_uri": {
			Type:     schema.TypeString,
			Computed: true,
			Description: "URL for the JSON Web Key Set (JWKS) containing the public keys " +
				"used for `private_key_jwt` authentication.",
		},
		"third_party_security_mode": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Security mode for third-party clients. `strict` enforces enhanced security controls",
		},
		"description": {
			Type:         schema.TypeString,
			Optional:     true,
			ValidateFunc: validation.StringLenBetween(0, 140),
			Description:  "Description of the purpose of the client.",
		},
		"app_type": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
			ValidateFunc: validation.StringInSlice(
				[]string{"native", "regular_web", "spa"}, false,
			),
			Description: "Type of application the client represents. " +
				"CIMD clients only support `native`, `spa`, and `regular_web`.",
		},
		"allowed_origins": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
			Description: "URLs that represent valid origins for cross-origin resource sharing. " +
				"By default, all your callback URLs will be allowed.",
		},
		"web_origins": {
			Type:        schema.TypeList,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "URLs that represent valid web origins for use with web message response mode.",
		},
		"grant_types": {
			Type:     schema.TypeList,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
				ValidateFunc: validation.StringInSlice(
					[]string{"authorization_code", "refresh_token"}, false,
				),
			},
			Description: "Types of grants that this client is authorized to use. " +
				"CIMD clients support `authorization_code` and `refresh_token`.",
		},
		"oidc_conformant": {
			Type:             schema.TypeBool,
			Optional:         true,
			Computed:         true,
			ValidateDiagFunc: validateBoolEquals(true),
			Description: "Whether this client conforms to strict OIDC specifications. " +
				"Must be `true` for CIMD clients.",
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
		"client_metadata": {
			Type:     schema.TypeMap,
			Optional: true,
			Elem:     schema.TypeString,
			Description: "Metadata associated with the client, in the form of an object with string values " +
				"(max 255 chars). Maximum of 10 metadata properties allowed. Field names (max 255 chars) are " +
				"alphanumeric and may only include the following special characters: " +
				"`:,-+=_*?\"/\\()<>@ [Tab] [Space]`.",
		},
		"require_proof_of_possession": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Makes the use of Proof-of-Possession mandatory for this client.",
		},
		"skip_non_verifiable_callback_uri_confirmation_prompt": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Indicates whether the confirmation prompt appears when using non-verifiable callback URIs. Set to true to skip the prompt, false to show it.",
		},
		"default_organization": {
			Type:        schema.TypeList,
			Optional:    true,
			MaxItems:    1,
			Description: "Configure and associate an organization with the Client",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"flows": {
						Type:        schema.TypeList,
						Required:    true,
						Elem:        &schema.Schema{Type: schema.TypeString},
						Description: "Definition of the flow that needs to be configured. Eg. client_credentials",
					},
					"organization_id": {
						Type:        schema.TypeString,
						Required:    true,
						Description: "The unique identifier of the organization",
					},
				},
			},
		},
		"validation": {
			Type:        schema.TypeList,
			Computed:    true,
			Description: "Validation result of the CIMD metadata document.",
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"valid": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Whether the metadata document passed validation.",
					},
					"violations": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Description: "Array of validation violation messages, if any. " +
							"Violations indicate issues that prevented the metadata document from being fully processed.",
					},
					"warnings": {
						Type:     schema.TypeList,
						Computed: true,
						Elem:     &schema.Schema{Type: schema.TypeString},
						Description: "Array of warning messages, if any. " +
							"Warnings indicate non-critical issues such as unsupported properties being ignored.",
					},
				},
			},
		},
		"token_quota": commons.TokenQuotaSchema(),
		"redirection_policy": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Controls whether Auth0 redirects users to the application's callback URL on authentication errors or in email verification flows.",
			ValidateFunc: validation.StringInSlice(
				[]string{"allow_always", "open_redirect_protection"}, false,
			),
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
					"alg": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						ValidateFunc: validation.StringInSlice(
							[]string{"RS256", "RS512", "PS256"}, false,
						),
						Description: "Algorithm used to sign JWTs. " +
							"CIMD clients support `RS256`, `RS512`, and `PS256` (asymmetric only).",
					},
					"secret_encoded": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Indicates whether the client secret is Base64-encoded.",
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
						Optional: true,
						Computed: true,
						Description: "Refresh token rotation type." +
							"Valid values are `rotating` and `non-rotating`",
						ValidateFunc: validation.StringInSlice([]string{"rotating", "non-rotating"}, false),
					},
					"expiration_type": {
						Type:     schema.TypeString,
						Optional: true,
						Computed: true,
						Description: "Refresh token expiration type. " +
							"Must be `expiring` for CIMD clients.",
						ValidateFunc: validation.StringInSlice(
							[]string{"expiring"}, false,
						),
					},
					"leeway": {
						Type:     schema.TypeInt,
						Optional: true,
						Computed: true,
						Description: "The amount of time in seconds in which a refresh token may be " +
							"reused without triggering reuse detection.",
					},
					"token_lifetime": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The absolute lifetime of a refresh token in seconds.",
					},
					"infinite_token_lifetime": {
						Type:     schema.TypeBool,
						Optional: true,
						Computed: true,
						Description: "Whether refresh tokens should remain valid indefinitely. " +
							"If false, `token_lifetime` should also be set.",
					},
					"idle_token_lifetime": {
						Type:        schema.TypeInt,
						Optional:    true,
						Computed:    true,
						Description: "The time in seconds after which inactive refresh tokens will expire.",
					},
					"infinite_idle_token_lifetime": {
						Type:             schema.TypeBool,
						Optional:         true,
						Computed:         true,
						ValidateDiagFunc: validateBoolEquals(false),
						Description: "Whether inactive refresh tokens should remain valid indefinitely. " +
							"Must be `false` for CIMD clients.",
					},
				},
			},
		},
	}
}

func validateBoolEquals(expected bool) schema.SchemaValidateDiagFunc {
	return func(val interface{}, path cty.Path) diag.Diagnostics {
		v, ok := val.(bool)
		if !ok {
			return diag.Errorf("expected type of %v to be bool", path)
		}
		if v != expected {
			return diag.Errorf("%v must be %t for CIMD clients", path, expected)
		}
		return nil
	}
}

func createCIMDClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	result, err := apiv2.Clients.RegisterCimdClient(ctx, &mgmtv2.RegisterCimdClientRequestContent{
		ExternalClientID: data.Get("external_client_id").(string),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("CIMD registration failed: %w", err))
	}

	clientID := result.GetClientID()
	if clientID == "" {
		return diag.Errorf("CIMD registration response missing client_id")
	}

	data.SetId(clientID)

	return updateCIMDClient(ctx, data, meta)
}

func updateCIMDClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if !data.IsNewResource() && data.HasChange("external_client_id_version") {
		if _, err := apiv2.Clients.RegisterCimdClient(ctx, &mgmtv2.RegisterCimdClientRequestContent{
			ExternalClientID: data.Get("external_client_id").(string),
		}); err != nil {
			return diag.FromErr(fmt.Errorf("CIMD sync failed: %w", err))
		}
	}

	updateReq := expandCIMDClient(data)
	if !data.IsNewResource() {
		applyCIMDNullFields(data, updateReq)
	}

	if emptyreq, err := isEmptyRequest(updateReq); err != nil {
		return diag.FromErr(fmt.Errorf("failed to determine if update request is empty: %w", err))
	} else if !emptyreq {
		if _, err := apiv2.Clients.Update(ctx, data.Id(), updateReq); err != nil {
			return diag.FromErr(err)
		}
	}

	return readCIMDClient(ctx, data, meta)
}

func readCIMDClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	client, err := apiv2.Clients.Get(ctx, data.Id(), &mgmtv2.GetClientRequestParameters{})
	if err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	preview, err := apiv2.Clients.PreviewCimdMetadata(ctx, &mgmtv2.PreviewCimdMetadataRequestContent{
		ExternalClientID: client.GetExternalClientID(),
	})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to fetch CIMD metadata preview: %w", err))
	}

	if err := flattenCIMDClient(data, client, preview.Validation); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func deleteCIMDClient(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	if err := apiv2.Clients.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}

func importCIMDClient(ctx context.Context, data *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	apiv2 := meta.(*config.Config).GetAPIV2()

	client, err := apiv2.Clients.Get(ctx, data.Id(), &mgmtv2.GetClientRequestParameters{})
	if err != nil {
		return nil, err
	}

	if client.GetExternalMetadataType() != mgmtv2.ClientExternalMetadataTypeEnumCimd {
		return nil, fmt.Errorf(
			"client %q is not a CIMD client. "+
				"Use the auth0_client resource to manage regular clients",
			data.Id(),
		)
	}

	data.SetId(client.GetClientID())

	return []*schema.ResourceData{data}, nil
}
