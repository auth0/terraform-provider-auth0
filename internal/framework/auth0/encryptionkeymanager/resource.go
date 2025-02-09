package encryptionkeymanager

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/auth0/go-auth0"
	"github.com/auth0/go-auth0/management"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
	frameworkError "github.com/auth0/terraform-provider-auth0/internal/framework/error"
	"github.com/auth0/terraform-provider-auth0/internal/wait"
)

type managerResource struct {
	cfg *config.Config
}

type encryptionKeyModel struct {
	KeyID       types.String      `tfsdk:"key_id"`
	Type        types.String      `tfsdk:"type"`
	State       types.String      `tfsdk:"state"`
	ParentKeyID types.String      `tfsdk:"parent_key_id"`
	CreatedAt   timetypes.RFC3339 `tfsdk:"created_at"`
	UpdatedAt   timetypes.RFC3339 `tfsdk:"updated_at"`
}

type customerProvidedRootKeyModel struct {
	encryptionKeyModel
	WrappedKey        types.String `tfsdk:"wrapped_key"`
	PublicWrappingKey types.String `tfsdk:"public_wrapping_key"`
	WrappingAlgorithm types.String `tfsdk:"wrapping_algorithm"`
}

// NewResource will return a new auth0_encryption_key_manager resource.
func NewResource() resource.Resource {
	return &managerResource{}
}

// Configure will be called by the framework to configure the auth0_encryption_key_manager resource.
func (r *managerResource) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_encryption_key_manager resource.
func (r *managerResource) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "auth0_encryption_key_manager"
}

// Schema will be called by the framework to get the schema for the auth0_encryption_key_manager resource.
func (r *managerResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description:         "A resource for managing the tenant root key.",
			MarkdownDescription: "A resource for managing the tenant root key.",
			Attributes: map[string]schema.Attribute{
				"key_rotation_id": schema.StringAttribute{
					Optional: true,
					Description: "If this value is changed, the encryption keys will be rotated. " +
						"A UUID is recommended for the key_rotation_id.",
					MarkdownDescription: "If this value is changed, the encryption keys will be rotated. " +
						"A UUID is recommended for the `key_rotation_id`.",
				},
				"encryption_keys": schema.ListNestedAttribute{
					Computed:            true,
					Description:         "All encryption keys.",
					MarkdownDescription: "All encryption keys.",
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"key_id": schema.StringAttribute{
								Computed:            true,
								Description:         "The key ID of the encryption key.",
								MarkdownDescription: "The key ID of the customer provided root key.",
							},
							"type": schema.StringAttribute{
								Computed: true,
								Description: "The type of the encryption key. One of " +
									"customer-provided-root-key, environment-root-key, " +
									"or tenant-master-key.",
								MarkdownDescription: "The type of the encryption key. One of " +
									"`customer-provided-root-key`, `environment-root-key`, " +
									"or `tenant-master-key`.",
							},
							"state": schema.StringAttribute{
								Computed: true,
								Description: "The state of the encryption key. One of " +
									"pre-activation, active, deactivated, or destroyed.",
								MarkdownDescription: "The state of the encryption key. One of " +
									"`pre-activation`, `active`, `deactivated`, or `destroyed`.",
							},
							"parent_key_id": schema.StringAttribute{
								Computed:            true,
								Description:         "The key ID of the parent wrapping key.",
								MarkdownDescription: "The key ID of the parent wrapping key.",
							},
							"created_at": schema.StringAttribute{
								Computed: true,
								Description: "The ISO 8601 formatted date the customer provided " +
									"root key was created.",
								MarkdownDescription: "The ISO 8601 formatted date the customer provided " +
									"root key was created.",
							},
							"updated_at": schema.StringAttribute{
								Computed: true,
								Description: "The ISO 8601 formatted date the customer provided " +
									"root key was updated.",
								MarkdownDescription: "The ISO 8601 formatted date the customer provided " +
									"root key was updated.",
							},
						},
					},
				},
			},
			Blocks: map[string]schema.Block{
				"customer_provided_root_key": schema.SingleNestedBlock{
					Description: "This attribute is used for provisioning the customer provided " +
						"root key. To initiate the provisioning process, create a new empty " +
						"customer_provided_root_key block. After applying this, the " +
						"public_wrapping_key can be retreived from the resource, and the new root " +
						"key should be generated by the customer and wrapped with the wrapping key, " +
						"then base64-encoded and added as the wrapped_key attribute.",
					MarkdownDescription: "This attribute is used for provisioning the customer provided " +
						"root key. To initiate the provisioning process, create a new empty " +
						"`customer_provided_root_key` block. After applying this, the " +
						"`public_wrapping_key` can be retreived from the resource, and the new root " +
						"key should be generated by the customer and wrapped with the wrapping key, " +
						"then base64-encoded and added as the `wrapped_key` attribute.",
					Attributes: map[string]schema.Attribute{
						"wrapped_key": schema.StringAttribute{
							Optional: true,
							Description: "The base64-encoded customer provided root key, " +
								"wrapped using the public_wrapping_key. This can be removed " +
								"after the wrapped key has been applied.",
							MarkdownDescription: "The base64-encoded customer provided root key, " +
								"wrapped using the `public_wrapping_key`. This can be removed " +
								"after the wrapped key has been applied.",
						},
						"public_wrapping_key": schema.StringAttribute{
							Computed:            true,
							Description:         "The public wrapping key in PEM format.",
							MarkdownDescription: "The public wrapping key in PEM format.",
						},
						"wrapping_algorithm": schema.StringAttribute{
							Computed: true,
							Description: "The algorithm that should be used to wrap the " +
								"customer provided root key. Should be CKM_RSA_AES_KEY_WRAP.",
							MarkdownDescription: "The algorithm that should be used to wrap the " +
								"customer provided root key. Should be `CKM_RSA_AES_KEY_WRAP`.",
						},
						"key_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The key ID of the customer provided root key.",
							MarkdownDescription: "The key ID of the customer provided root key.",
						},
						"type": schema.StringAttribute{
							Computed: true,
							Description: "The type of the customer provided root key. " +
								"Should be customer-provided-root-key.",
							MarkdownDescription: "The type of the customer provided root key. " +
								"Should be `customer-provided-root-key`.",
						},
						"state": schema.StringAttribute{
							Computed: true,
							Description: "The state of the encryption key. One of " +
								"pre-activation, active, deactivated, or destroyed.",
							MarkdownDescription: "The state of the encryption key. One of " +
								"`pre-activation`, `active`, `deactivated`, or `destroyed`.",
						},
						"parent_key_id": schema.StringAttribute{
							Computed:            true,
							Description:         "The key ID of the parent wrapping key.",
							MarkdownDescription: "The key ID of the parent wrapping key.",
						},
						"created_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
							Description: "The ISO 8601 formatted date the customer provided " +
								"root key was created.",
							MarkdownDescription: "The ISO 8601 formatted date the customer provided " +
								"root key was created.",
						},
						"updated_at": schema.StringAttribute{
							CustomType: timetypes.RFC3339Type{},
							Computed:   true,
							Description: "The ISO 8601 formatted date the customer provided " +
								"root key was updated.",
							MarkdownDescription: "The ISO 8601 formatted date the customer provided " +
								"root key was updated.",
						},
					},
				},
			},
		}
	}
}

// Create will be called by the framework to initialise a new auth0_encryption_key_manager resource.
func (r *managerResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	response.Diagnostics.Append(updateResource(ctx, r.cfg.GetAPI(), request.Plan, response.State, &response.State)...)
}

// Update will be called by the framework to update an auth0_encryption_key_manager resource.
func (r *managerResource) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	response.Diagnostics.Append(updateResource(ctx, r.cfg.GetAPI(), request.Plan, request.State, &response.State)...)
}

func updateResource(ctx context.Context, api *management.Management, requestPlan tfsdk.Plan, requestState tfsdk.State, responseState *tfsdk.State) diag.Diagnostics {
	diagnostics := updateKeyRotationID(ctx, api, requestPlan, requestState, responseState)
	if diagnostics.HasError() {
		return diagnostics
	}

	diagnostics = updateCustomerProvidedRootKey(ctx, api, requestPlan, requestState, responseState)
	if diagnostics.HasError() {
		return diagnostics
	}

	diagnostics.Append(readResource(ctx, api, responseState)...)
	return diagnostics
}

func updateKeyRotationID(ctx context.Context, api *management.Management, requestPlan tfsdk.Plan, requestState tfsdk.State, responseState *tfsdk.State) diag.Diagnostics {
	var keyRotationIDPlan, keyRotationIDState types.String
	var diagnostics diag.Diagnostics
	diagnostics.Append(requestPlan.GetAttribute(ctx, path.Root("key_rotation_id"), &keyRotationIDPlan)...)
	if diagnostics.HasError() {
		return diagnostics
	}

	diagnostics.Append(requestState.GetAttribute(ctx, path.Root("key_rotation_id"), &keyRotationIDState)...)
	if diagnostics.HasError() {
		return diagnostics
	}
	if !keyRotationIDPlan.Equal(keyRotationIDState) && !keyRotationIDPlan.IsNull() {
		diagnostics.Append(frameworkError.Diagnostics(api.EncryptionKey.Rekey(ctx))...)
		if diagnostics.HasError() {
			return diagnostics
		}
	}
	if !keyRotationIDPlan.Equal(keyRotationIDState) {
		diagnostics.Append(responseState.SetAttribute(ctx, path.Root("key_rotation_id"), keyRotationIDPlan)...)
	}
	return diagnostics
}

func updateCustomerProvidedRootKey(ctx context.Context, api *management.Management, requestPlan tfsdk.Plan, requestState tfsdk.State, responseState *tfsdk.State) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	var rootKeyPlanObject, rootKeyStateObject, rootKeyResponseObject types.Object
	var rootKeyPlan, rootKeyState, rootKeyResponse customerProvidedRootKeyModel
	diagnostics.Append(requestPlan.GetAttribute(ctx, path.Root("customer_provided_root_key"), &rootKeyPlanObject)...)
	diagnostics.Append(responseState.GetAttribute(ctx, path.Root("customer_provided_root_key"), &rootKeyResponseObject)...)
	diagnostics.Append(requestState.GetAttribute(ctx, path.Root("customer_provided_root_key"), &rootKeyStateObject)...)
	if diagnostics.HasError() {
		return diagnostics
	}
	if !rootKeyPlanObject.IsNull() {
		// We have a customer_provided_root_key block in the new plan.
		diagnostics.Append(rootKeyPlanObject.As(ctx, &rootKeyPlan, basetypes.ObjectAsOptions{})...)
		if diagnostics.HasError() {
			return diagnostics
		}
	}
	if !rootKeyResponseObject.IsNull() {
		// We have a customer_provided_root_key block in the response.
		diagnostics.Append(rootKeyResponseObject.As(ctx, &rootKeyResponse, basetypes.ObjectAsOptions{})...)
		if diagnostics.HasError() {
			return diagnostics
		}
	}
	if !rootKeyStateObject.IsNull() {
		// We had a customer_provided_root_key block previously.
		diagnostics.Append(rootKeyStateObject.As(ctx, &rootKeyState, basetypes.ObjectAsOptions{})...)
		if diagnostics.HasError() {
			return diagnostics
		}

		rootKeyID := rootKeyState.KeyID.ValueString()
		if rootKeyPlanObject.IsNull() {
			// The customer_provided_root_key block is not present, check if there was a key.
			if len(rootKeyID) > 0 {
				// If it had a value, but the block was removed, remove the key.
				diagnostics.Append(frameworkError.Diagnostics(removeKey(ctx, api, rootKeyID))...)
				if diagnostics.HasError() {
					return diagnostics
				}
			}
		} else {
			// We still have a customer_provided_root_key block, check if we have a new wrapped_key.
			if rootKeyState.WrappedKey.IsNull() && !rootKeyPlan.WrappedKey.IsNull() {
				if len(rootKeyState.PublicWrappingKey.ValueString()) == 0 {
					diagnostics.AddError("Error", "The wrapped_key attribute should not be specified in the "+
						"customer_provided_root_key block until after the public_wrapping_key has been generated")
					return diagnostics
				}
				wrappedKey := rootKeyPlan.WrappedKey.ValueString()
				diagnostics.Append(frameworkError.Diagnostics(importWrappedKey(ctx, api, auth0.String(rootKeyID), auth0.String(wrappedKey)))...)
				if diagnostics.HasError() {
					return diagnostics
				}
				rootKeyResponse = flattenCustomerProvidedRootKey(nil, nil, &wrappedKey, rootKeyResponse)
			}
		}
	} else if !rootKeyPlanObject.IsNull() {
		// We did not have a customer_provided_root_key block previously, but now we do.
		if !rootKeyPlan.WrappedKey.IsNull() && !rootKeyPlan.WrappedKey.IsUnknown() {
			diagnostics.AddError("Error", "The wrapped_key attribute should not be specified in the "+
				"customer_provided_root_key block until after the public_wrapping_key has been generated")
			return diagnostics
		}
		rootKey, wrappingKey, err := createRootKey(ctx, api)
		diagnostics.Append(frameworkError.Diagnostics(err)...)
		if diagnostics.HasError() {
			return diagnostics
		}
		rootKeyResponse = flattenCustomerProvidedRootKey(rootKey, wrappingKey, nil, rootKeyResponse)
		if diagnostics.HasError() {
			return diagnostics
		}
	}

	if !rootKeyPlanObject.IsNull() {
		diagnostics.Append(responseState.SetAttribute(ctx, path.Root("customer_provided_root_key"), rootKeyResponse)...)
	} else {
		diagnostics.Append(responseState.SetAttribute(ctx, path.Root("customer_provided_root_key"), types.ObjectNull(rootKeyPlanObject.AttributeTypes(ctx)))...)
	}

	return diagnostics
}

// Read will be called by the framework to read an auth0_encryption_key_manager resource.
func (r *managerResource) Read(ctx context.Context, _ resource.ReadRequest, response *resource.ReadResponse) {
	response.Diagnostics.Append(readResource(ctx, r.cfg.GetAPI(), &response.State)...)
}

func readResource(ctx context.Context, api *management.Management, responseState *tfsdk.State) diag.Diagnostics {
	var diagnostics diag.Diagnostics
	encryptionKeys := make([]*management.EncryptionKey, 0)
	page := 0

	for {
		encryptionKeyList, err := api.EncryptionKey.List(ctx, management.Page(page), management.PerPage(100))
		if err != nil {
			diagnostics.Append(frameworkError.Diagnostics(err)...)
			return diagnostics
		}
		encryptionKeys = append(encryptionKeys, encryptionKeyList.Keys...)
		if !encryptionKeyList.HasNext() {
			break
		}
		page++
	}

	var rootKeyResponseObject types.Object
	diagnostics.Append(responseState.GetAttribute(ctx, path.Root("customer_provided_root_key"), &rootKeyResponseObject)...)
	if diagnostics.HasError() {
		return diagnostics
	}
	if !rootKeyResponseObject.IsNull() {
		var rootKeyResponse customerProvidedRootKeyModel
		diagnostics.Append(rootKeyResponseObject.As(ctx, &rootKeyResponse, basetypes.ObjectAsOptions{})...)
		if diagnostics.HasError() {
			return diagnostics
		}
		// First try to find a key that is going through the activation process.
		rootKey := getKeyByTypeAndState("customer-provided-root-key", "pre-activation", encryptionKeys)

		if rootKey == nil {
			// If we didn't find one, try to find a key that is already active.
			rootKey = getKeyByTypeAndState("customer-provided-root-key", "active", encryptionKeys)
		}

		if rootKey != nil {
			var rootKeyResponse customerProvidedRootKeyModel
			if !rootKeyResponseObject.IsNull() && !rootKeyResponseObject.IsUnknown() {
				diagnostics.Append(rootKeyResponseObject.As(ctx, &rootKeyResponse, basetypes.ObjectAsOptions{})...)
				if diagnostics.HasError() {
					return diagnostics
				}
			}
			diagnostics.Append(responseState.SetAttribute(ctx, path.Root("customer_provided_root_key"), flattenCustomerProvidedRootKey(rootKey, nil, nil, rootKeyResponse))...)
			if diagnostics.HasError() {
				return diagnostics
			}
		}
	}

	diagnostics.Append(responseState.SetAttribute(ctx, path.Root("encryption_keys"), flattenEncryptionKeys(encryptionKeys))...)

	return diagnostics
}

func (r *managerResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	api := r.cfg.GetAPI()
	var rootKeyRequestObject types.Object

	response.Diagnostics.Append(request.State.GetAttribute(ctx, path.Root("customer_provided_root_key"), &rootKeyRequestObject)...)
	if !response.Diagnostics.HasError() && !rootKeyRequestObject.IsNull() {
		var rootKeyRequest customerProvidedRootKeyModel
		response.Diagnostics.Append(rootKeyRequestObject.As(ctx, &rootKeyRequest, basetypes.ObjectAsOptions{})...)
		if !response.Diagnostics.HasError() && len(rootKeyRequest.KeyID.ValueString()) > 0 {
			response.Diagnostics.Append(frameworkError.Diagnostics(removeKey(ctx, api, rootKeyRequest.KeyID.ValueString()))...)
		}
	}
}

func removeKey(ctx context.Context, api *management.Management, keyID string) error {
	if err := api.EncryptionKey.Delete(ctx, keyID); err != nil {
		return err
	}

	// Wait until the key is actually destroyed.
	return wait.Until(200, 10, func() (bool, error) {
		key, err := api.EncryptionKey.Read(ctx, keyID)
		if err != nil {
			return false, err
		}
		return key.GetState() == "destroyed", nil
	})
}

func importWrappedKey(ctx context.Context, api *management.Management, keyID, wrappedKey *string) error {
	encryptionKey := management.EncryptionKey{
		KID:        keyID,
		WrappedKey: wrappedKey,
	}
	if err := api.EncryptionKey.ImportWrappedKey(ctx, &encryptionKey); err != nil {
		return err
	}
	// Wait until the key is actually activated.
	return wait.Until(200, 10, func() (bool, error) {
		key, err := api.EncryptionKey.Read(ctx, *keyID)
		if err != nil {
			return false, err
		}
		return key.GetState() == "active", nil
	})
}

func createRootKey(ctx context.Context, api *management.Management) (*management.EncryptionKey, *management.WrappingKey, error) {
	key := management.EncryptionKey{
		Type: auth0.String("customer-provided-root-key"),
	}
	if err := api.EncryptionKey.Create(ctx, &key); err != nil {
		return nil, nil, err
	}

	// Wait until the key is actually available.
	err := wait.Until(100, 20, func() (bool, error) {
		if _, err := api.EncryptionKey.Read(ctx, key.GetKID()); err != nil {
			if internalError.IsStatusNotFound(err) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
	if err != nil {
		return nil, nil, err
	}

	wrappingKey, err := api.EncryptionKey.CreatePublicWrappingKey(ctx, key.GetKID())
	if err != nil {
		return nil, nil, err
	}

	return &key, wrappingKey, nil
}
