package resourceserver

import (
	"context"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/framework/error"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/framework/schema"
)

type resourceScopeType struct {
	cfg *config.Config
}

type scopeModel struct {
	Scope                    types.String `tfsdk:"scope"`
	ResourceServerIdentifier types.String `tfsdk:"resource_server_identifier"`
	Description              types.String `tfsdk:"description"`
}

// NewScopeResource will return a new auth0_resource_server_scope resource.
func NewScopeResource() resource.Resource {
	return &resourceScopeType{}
}

// Configure will be called by the framework to configure the auth0_resource_server_scope resource.
func (r *resourceScopeType) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_resource_server_scope resource.
func (r *resourceScopeType) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "auth0_resource_server_scope"
}

// Schema will be called by the framework to get the schema for the auth0_resource_server_scope resource.
func (r *resourceScopeType) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description: "With this resource, you can manage scopes (permissions) associated with a resource server (API).",
			Attributes: map[string]schema.Attribute{
				"scope": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "Name of the scope (permission).",
				},
				"resource_server_identifier": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "Identifier of the resource server that the scope (permission) is associated with.",
				},
				"description": schema.StringAttribute{
					Optional:    true,
					Description: "Description of the scope (permission).",
				},
			},
		}
	}
}

// ImportState will be called by the framework to import an existing auth0_resource_server_scope resource.
func (r *resourceScopeType) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	internalSchema.ImportStateCompositeID(ctx, request, response, path.Root("resource_server_identifier"), path.Root("scope"))
}

// Create will be called by the framework to initialise a new auth0_resource_server_scope resource.
func (r *resourceScopeType) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()

	var model scopeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()
	scope := model.Scope.ValueString()
	descriptionPtr := model.Description.ValueStringPointer()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("resource_server_identifier"), resourceServerIdentifier)...)
	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("scope"), scope)...)
	if response.Diagnostics.HasError() {
		return
	}

	for _, apiScope := range existingAPI.GetScopes() {
		if apiScope.GetValue() == scope {
			response.Diagnostics.Append(readScopeResource(ctx, api, resourceServerIdentifier, scope, &response.State)...)
			return
		}
	}

	scopes := append(existingAPI.GetScopes(), management.ResourceServerScope{
		Value:       &scope,
		Description: descriptionPtr,
	})

	resourceServer := management.ResourceServer{
		Scopes: &scopes,
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, &resourceServer); err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	response.Diagnostics.Append(readScopeResource(ctx, api, resourceServerIdentifier, scope, &response.State)...)
}

// Update will be called by the framework to update an auth0_resource_server_scope resource.
func (r *resourceScopeType) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()

	var model scopeModel
	response.Diagnostics.Append(request.Config.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()
	scope := model.Scope.ValueString()
	newDescription := model.Description.ValueString()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	updatedScopes := make([]management.ResourceServerScope, 0)

	found := false
	for _, existingScope := range existingAPI.GetScopes() {
		updated := existingScope
		if existingScope.GetValue() == scope {
			found = true
			updated.Description = &newDescription
		}
		updatedScopes = append(updatedScopes, updated)
	}

	if !found {
		response.State.RemoveResource(ctx)
		return
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, &management.ResourceServer{
		Scopes: &updatedScopes,
	}); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	response.Diagnostics.Append(readScopeResource(ctx, api, resourceServerIdentifier, scope, &response.State)...)
}

// Read will be called by the framework to read an auth0_resource_server_scope resource.
func (r *resourceScopeType) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	api := r.cfg.GetAPI()

	var model scopeModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()
	scope := model.Scope.ValueString()

	response.Diagnostics.Append(readScopeResource(ctx, api, resourceServerIdentifier, scope, &response.State)...)
}

func readScopeResource(ctx context.Context, api *management.Management, resourceServerIdentifier, scope string, responseState *tfsdk.State) diag.Diagnostics {
	existingAPI, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		return internalError.HandleAPIError(ctx, responseState, err)
	}

	for _, existingScope := range existingAPI.GetScopes() {
		if existingScope.GetValue() == scope {
			return responseState.SetAttribute(ctx, path.Root("description"), existingScope.Description)
		}
	}

	// If we make it this far, we didn't find it on the server.
	responseState.RemoveResource(ctx)
	var diagnostics diag.Diagnostics
	diagnostics.AddWarning("Resource missing", "The resource_server_scope resource is missing")

	return diagnostics
}

// Delete will be called by the framework to delete an auth0_resource_server_scope resource.
func (r *resourceScopeType) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()
	var model scopeModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()
	scope := model.Scope.ValueString()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	existingAPI, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	updateScopes := make([]management.ResourceServerScope, 0)
	for _, existingScope := range existingAPI.GetScopes() {
		if existingScope.GetValue() != scope {
			updateScopes = append(updateScopes, existingScope)
		}
	}

	if err := api.ResourceServer.Update(
		ctx,
		resourceServerIdentifier,
		&management.ResourceServer{
			Scopes: &updateScopes,
		},
	); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
	}
}
