package resourceserver

import (
	"context"
	"fmt"

	"github.com/auth0/go-auth0/management"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
)

type resourceScopesType struct {
	cfg *config.Config
}

type resourceScopesModel struct {
	ResourceServerIdentifier types.String `tfsdk:"resource_server_identifier"`
	Scopes                   types.Set    `tfsdk:"scopes"`
}

type scopesElementModel struct {
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
}

var scopesElementTypeMap = map[string]attr.Type{
	"name":        types.StringType,
	"description": types.StringType,
}

var scopesElementType = types.ObjectType{
	AttrTypes: scopesElementTypeMap,
}

// NewScopesResource will return a new auth0_resource_server_scopes resource.
func NewScopesResource() resource.Resource {
	return &resourceScopesType{}
}

// Configure will be called by the framework to configure the auth0_resource_server_scopes resource.
func (r *resourceScopesType) Configure(_ context.Context, request resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if request.ProviderData != nil {
		r.cfg = request.ProviderData.(*config.Config)
	}
}

// Metadata will be called by the framework to get the type name for the auth0_resource_server_scopes resource.
func (r *resourceScopesType) Metadata(_ context.Context, _ resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = "auth0_resource_server_scopes"
}

// Schema will be called by the framework to get the schema for the auth0_resource_server_scopes resource.
func (r *resourceScopesType) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	if response != nil {
		response.Schema = schema.Schema{
			Description: "With this resource, you can manage scopes (permissions) associated with a resource server (API).",
			Attributes: map[string]schema.Attribute{
				"resource_server_identifier": schema.StringAttribute{
					Required: true,
					PlanModifiers: []planmodifier.String{
						stringplanmodifier.RequiresReplace(),
					},
					Description: "Identifier of the resource server that the scopes (permission) are associated with.",
				},
			},
			Blocks: map[string]schema.Block{
				"scopes": schema.SetNestedBlock{
					Description: "List of the scopes associated with a resource server.",
					NestedObject: schema.NestedBlockObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Required: true,
								Description: "Name of the scope (permission). Examples include " +
									"read:appointments or delete:appointments.",
								MarkdownDescription: "Name of the scope (permission). Examples include " +
									"`read:appointments` or `delete:appointments`.",
							},
							"description": schema.StringAttribute{
								Optional:    true,
								Description: "User-friendly description of the scope (permission).",
							},
						},
					},
				},
			},
		}
	}
}

// ImportState will be called by the framework to import an existing auth0_resource_server_scopes resource.
func (r *resourceScopesType) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("resource_server_identifier"), request, response)
}

// Create will be called by the framework to initialise a new auth0_resource_server_scopes resource.
func (r *resourceScopesType) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()

	var model resourceScopesModel
	response.Diagnostics.Append(request.Config.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	existingResourceServer, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		return
	}

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandResourceServerScopes(ctx, model.Scopes),
	}

	response.Diagnostics.Append(guardAgainstErasingUnwantedScopes(
		existingResourceServer.GetIdentifier(),
		existingResourceServer.GetScopes(),
		updatedResourceServer.GetScopes(),
	)...)

	if response.Diagnostics.HasError() {
		response.State.RemoveResource(ctx)
		return
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, updatedResourceServer); err != nil {
		response.Diagnostics.Append(internalError.Diagnostics(err)...)
		response.State.RemoveResource(ctx)
		return
	}

	response.Diagnostics.Append(response.State.SetAttribute(ctx, path.Root("resource_server_identifier"), resourceServerIdentifier)...)
	response.Diagnostics.Append(readScopesResource(ctx, api, resourceServerIdentifier, &response.State)...)
}

// Update will be called by the framework to initialise a new auth0_resource_server_scopes resource.
func (r *resourceScopesType) Update(ctx context.Context, request resource.UpdateRequest, response *resource.UpdateResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()

	var model resourceScopesModel
	response.Diagnostics.Append(request.Config.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	updatedResourceServer := &management.ResourceServer{
		Scopes: expandResourceServerScopes(ctx, model.Scopes),
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, updatedResourceServer); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
		return
	}

	response.Diagnostics.Append(readScopesResource(ctx, api, resourceServerIdentifier, &response.State)...)
}

// Delete will be called by the framework to delete an auth0_resource_server_scopes resource.
func (r *resourceScopesType) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	api := r.cfg.GetAPI()
	mutex := r.cfg.GetMutex()

	var model resourceScopesModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()

	mutex.Lock(resourceServerIdentifier) // Prevents colliding API requests between other `auth0_resource_server_scope` resource.
	defer mutex.Unlock(resourceServerIdentifier)

	resourceServer := &management.ResourceServer{
		Scopes: &[]management.ResourceServerScope{},
	}

	if err := api.ResourceServer.Update(ctx, resourceServerIdentifier, resourceServer); err != nil {
		response.Diagnostics.Append(internalError.HandleAPIError(ctx, &response.State, err)...)
	}
}

// Read will be called by the framework to read an auth0_resource_server_scopes resource.
func (r *resourceScopesType) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	api := r.cfg.GetAPI()

	var model resourceScopesModel
	response.Diagnostics.Append(request.State.Get(ctx, &model)...)
	if response.Diagnostics.HasError() {
		return
	}

	resourceServerIdentifier := model.ResourceServerIdentifier.ValueString()

	response.Diagnostics.Append(readScopesResource(ctx, api, resourceServerIdentifier, &response.State)...)
}

func readScopesResource(ctx context.Context, api *management.Management, resourceServerIdentifier string, responseState *tfsdk.State) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	resourceServer, err := api.ResourceServer.Read(ctx, resourceServerIdentifier)
	if err != nil {
		diagnostics.Append(internalError.HandleAPIError(ctx, responseState, err)...)
		return diagnostics
	}

	scopes, diagnostics := flattenResourceServerScopesSet(resourceServer.Scopes)
	if !diagnostics.HasError() {
		diagnostics.Append(responseState.SetAttribute(ctx, path.Root("scopes"), scopes)...)
	}

	return diagnostics
}

func guardAgainstErasingUnwantedScopes(
	apiIdentifier string,
	apiScopes []management.ResourceServerScope,
	configScopes []management.ResourceServerScope,
) diag.Diagnostics {
	var diagnostics diag.Diagnostics

	if len(apiScopes) > 0 && !cmp.Equal(configScopes, apiScopes) {
		diagnostics.AddError(
			"Resource Server with non empty scopes",
			cmp.Diff(configScopes, apiScopes)+
				fmt.Sprintf("\nThe resource server already has scopes attached to it. "+
					"Import the resource instead in order to proceed with the changes. "+
					"Run: 'terraform import auth0_resource_server_scopes.<given-name> %s'.", apiIdentifier),
		)
	}

	return diagnostics
}
