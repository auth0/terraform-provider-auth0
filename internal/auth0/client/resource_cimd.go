package client

import (
	"context"
	"fmt"

	managementv2 "github.com/auth0/go-auth0/v2/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// schemaFields maps field names to their editable definitions.
// Nil entry = editable leaf. Non-nil entry = editable with overrides/children.
// Unlisted fields are made read-only.
type schemaFields map[string]*schemaFieldDefinition

// schemaFieldDefinition defines overrides and nested editable children for a field.
type schemaFieldDefinition struct {
	ValidateDiagFunc schema.SchemaValidateDiagFunc
	Description      string
	editableChildren schemaFields
}

// cimdEditableSchemaDefinitions defines PATCHable fields for CIMD clients.
var cimdEditableSchemaDefinitions = schemaFields{
	"allowed_origins": nil,
	"description":     nil,
	"oidc_conformant": {
		Description: "Indicates whether this client will conform to strict OIDC specifications." +
			"Must be `true` for CIMD clients.",
	},
	"organization_discovery_methods": nil,
	"web_origins":                    nil,
	"grant_types": {
		Description: "Types of grants that this client is authorized to use." +
			"Only `authorization_code` and `refresh_token` are supported for CIMD clients.",
	},
	"client_metadata":             nil,
	"require_proof_of_possession": nil,
	"app_type": {
		Description:      "Type of application the client represents. CIMD clients only support `native`, `spa`, and `regular_web`.",
		ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"native", "regular_web", "spa"}, false)),
	},
	"default_organization": nil,
	"token_quota":          nil,
	"jwt_configuration": {
		editableChildren: schemaFields{
			"lifetime_in_seconds": nil,
			"alg": {
				Description:      "Algorithm used to sign JWTs. CIMD clients support `RS256`, `RS512`, and `PS256` (asymmetric only).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"RS256", "RS512", "PS256"}, false)),
			},
		},
	},
	"refresh_token": {
		editableChildren: schemaFields{
			"rotation_type":           nil,
			"leeway":                  nil,
			"token_lifetime":          nil,
			"infinite_token_lifetime": nil,
			"idle_token_lifetime":     nil,
			"expiration_type": {
				Description:      "Must be `expiring` for CIMD clients. Required in PATCH body when refresh_token is present.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"expiring"}, false)),
			},
			"infinite_idle_token_lifetime": {
				Description: "Whether inactive refresh tokens should remain valid indefinitely." +
					"Must be `false` for CIMD clients.",
			},
		},
	},
	"skip_non_verifiable_callback_uri_confirmation_prompt": nil,
}

// NewCIMDResource returns a new auth0_client_cimd resource.
func NewCIMDResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createCIMDClient,
		ReadContext:   readClient,
		UpdateContext: updateClient,
		DeleteContext: deleteClient,
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

// cimdClientSchema derives the CIMD schema from auth0_client by applying
// cimdEditableSchemaDefinitions and adding the external_client_id field.
func cimdClientSchema() map[string]*schema.Schema {
	baseSchema := NewResource().Schema

	updateSchemaProperties(baseSchema, cimdEditableSchemaDefinitions)

	// Computed in auth0_client → Required+ForceNew for CIMD registration URL.
	baseSchema["external_client_id"] = &schema.Schema{
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
		Description: "The HTTPS URL of the Client ID Metadata Document. " +
			"Must include a path component (e.g. `https://app.example.com/client.json`). " +
			"This value is immutable after creation.",
	}

	return baseSchema
}

// updateSchemaProperties applies the editable allowlist to a schema map.
// Listed fields become Optional+Computed with any overrides applied.
// Unlisted fields (and their children) become read-only.
func updateSchemaProperties(schemas map[string]*schema.Schema, editable schemaFields) {
	for key, field := range schemas {
		entry, isEditable := editable[key]
		if !isEditable {
			makeSchemaReadOnly(field)
			continue
		}

		makeSchemaEditable(field)

		if entry == nil {
			continue
		}

		if entry.Description != "" {
			field.Description = entry.Description
		}
		if entry.ValidateDiagFunc != nil {
			field.ValidateFunc = nil
			field.ValidateDiagFunc = entry.ValidateDiagFunc
		}
		if entry.editableChildren != nil {
			if nestedResource, ok := field.Elem.(*schema.Resource); ok {
				updateSchemaProperties(nestedResource.Schema, entry.editableChildren)
			}
		}
	}
}

// makeSchemaEditable marks a field as Optional+Computed.
func makeSchemaEditable(s *schema.Schema) {
	s.Required = false
	s.Optional = true
	s.Computed = true
	s.ForceNew = false
}

// makeSchemaReadOnly recursively marks a field and its children as Computed-only.
func makeSchemaReadOnly(s *schema.Schema) {
	s.Required = false
	s.Optional = false
	s.Computed = true
	s.ForceNew = false
	s.Default = nil
	s.ValidateFunc = nil
	s.ValidateDiagFunc = nil
	s.DiffSuppressFunc = nil
	s.AtLeastOneOf = nil
	s.RequiredWith = nil
	s.ConflictsWith = nil
	s.ExactlyOneOf = nil
	s.MaxItems = 0
	s.MinItems = 0

	if r, ok := s.Elem.(*schema.Resource); ok {
		for _, sub := range r.Schema {
			makeSchemaReadOnly(sub)
		}
	}
}

func createCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) diag.Diagnostics {
	apiv2 := meta.(*config.Config).GetAPIV2()

	externalClientID := data.Get("external_client_id").(string)

	req := &managementv2.RegisterCimdClientRequestContent{}
	req.SetExternalClientID(externalClientID)

	result, err := apiv2.Clients.RegisterCimdClient(ctx, req)
	if err != nil {
		return diag.FromErr(fmt.Errorf("CIMD registration failed: %w", err))
	}

	clientID := result.GetClientID()
	if clientID == "" {
		return diag.Errorf("CIMD registration response missing client_id")
	}

	data.SetId(clientID)

	// PATCH editable fields via v1 SDK.
	api := meta.(*config.Config).GetAPI()

	client, err := expandClient(data)
	if err != nil {
		return diag.FromErr(err)
	}

	// Clear CIMD-blocked field set by expandClient for new resources.
	client.TokenEndpointAuthMethod = nil

	// Skip PATCH if no editable fields were set.
	if clientHasChange(client) {
		if err := api.Client.Update(ctx, data.Id(), client); err != nil {
			return diag.FromErr(internalError.HandleAPIError(data, err))
		}
	}

	return readClient(ctx, data, meta)
}

func importCIMDClient(ctx context.Context, data *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	api := meta.(*config.Config).GetAPI()

	client, err := api.Client.Read(ctx, data.Id())
	if err != nil {
		return nil, err
	}

	if client.GetExternalMetadataType() != "cimd" {
		return nil, fmt.Errorf(
			"client %q is not a CIMD client (external_metadata_type=%q). "+
				"Use the auth0_client resource to manage regular clients",
			data.Id(),
			client.GetExternalMetadataType(),
		)
	}

	data.SetId(client.GetClientID())

	return []*schema.ResourceData{data}, nil
}
