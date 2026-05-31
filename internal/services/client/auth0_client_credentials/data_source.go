// data_source.go — implements the auth0_client_credentials data source.
//
// Looks up the active authentication method of a client (by `client_id`) and,
// for credential-based methods, fetches all attached credentials. Credentials
// are exposed as a flat list — there is no `pem` because the Auth0 API never
// returns it (write-only).
package auth0clientcredentials

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

var (
	_ datasource.DataSource              = (*credentialsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*credentialsDataSource)(nil)
)

// NewDataSource returns a fresh auth0_client_credentials data source implementation.
func NewDataSource() datasource.DataSource { return &credentialsDataSource{} }

type credentialsDataSource struct {
	mgmt *mgmtclient.Management
}

// dsModel is the data-source view: a flat list of credentials regardless of
// auth method (vs. the resource which uses typed blocks per method).
type dsModel struct {
	ID                   types.String `tfsdk:"id"`
	ClientID             types.String `tfsdk:"client_id"`
	AuthenticationMethod types.String `tfsdk:"authentication_method"`
	Credentials          types.List   `tfsdk:"credentials"`
}

// dsCredentialAttrTypes returns the attribute types for one entry in the
// `credentials` list of the data source. It excludes `pem` and
// `parse_expiry_from_cert` which are write-only on the API.
func dsCredentialAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                types.StringType,
		"name":              types.StringType,
		"credential_type":   types.StringType,
		"algorithm":         types.StringType,
		"kid":               types.StringType,
		"subject_dn":        types.StringType,
		"thumbprint_sha256": types.StringType,
		"expires_at":        types.StringType,
		"created_at":        types.StringType,
		"updated_at":        types.StringType,
	}
}

func (d *credentialsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_client_credentials"
}

func (d *credentialsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if m, ok := framework.ManagementFromDataSource(req, resp); ok {
		d.mgmt = m
	}
}

func (d *credentialsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dsschema.Schema{
		Description: "Read the current authentication method and credentials of an Auth0 client. " +
			"For secret-based methods, `credentials` is empty.",
		Attributes: map[string]dsschema.Attribute{
			"id": dsschema.StringAttribute{
				Computed:    true,
				Description: "Same as `client_id`.",
			},
			"client_id": dsschema.StringAttribute{
				Required:    true,
				Description: "ID of the client to look up.",
			},
			"authentication_method": dsschema.StringAttribute{
				Computed: true,
				Description: "Currently-active authentication method on the client. One of: " +
					"`none`, `client_secret_post`, `client_secret_basic`, " +
					"`private_key_jwt`, `tls_client_auth`, `self_signed_tls_client_auth`.",
			},
			"credentials": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "Credentials currently attached to the client for the active credential-based authentication method. Empty for secret-based methods.",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"id":                dsschema.StringAttribute{Computed: true, Description: "Credential ID."},
						"name":              dsschema.StringAttribute{Computed: true, Description: "Friendly name."},
						"credential_type":   dsschema.StringAttribute{Computed: true, Description: "Credential type."},
						"algorithm":         dsschema.StringAttribute{Computed: true, Description: "Signing algorithm."},
						"kid":               dsschema.StringAttribute{Computed: true, Description: "Key identifier."},
						"subject_dn":        dsschema.StringAttribute{Computed: true, Description: "X.509 subject DN."},
						"thumbprint_sha256": dsschema.StringAttribute{Computed: true, Description: "SHA-256 thumbprint."},
						"expires_at":        dsschema.StringAttribute{Computed: true, Description: "ISO 8601 expiry timestamp."},
						"created_at":        dsschema.StringAttribute{Computed: true, Description: "ISO 8601 creation timestamp."},
						"updated_at":        dsschema.StringAttribute{Computed: true, Description: "ISO 8601 last-update timestamp."},
					},
				},
			},
		},
	}
}

func (d *credentialsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg dsModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clientID := cfg.ClientID.ValueString()
	if clientID == "" {
		resp.Diagnostics.AddError("Missing client_id", "`client_id` is required.")
		return
	}

	got, err := d.mgmt.Clients.Get(ctx, clientID, &mgmt.GetClientRequestParameters{})
	if err != nil {
		framework.AddAPIError(&resp.Diagnostics, "Failed to read client", err)
		return
	}

	state := dsModel{
		ID:       types.StringValue(clientID),
		ClientID: cfg.ClientID,
	}

	// Determine the auth method and any referenced credential IDs.
	cam := got.ClientAuthenticationMethods
	if cam != nil && (cam.PrivateKeyJwt != nil || cam.TLSClientAuth != nil || cam.SelfSignedTLSClientAuth != nil) {
		var method string
		var refIDs []string
		switch {
		case cam.PrivateKeyJwt != nil:
			method = "private_key_jwt"
			for _, r := range cam.PrivateKeyJwt.Credentials {
				if r != nil {
					refIDs = append(refIDs, r.ID)
				}
			}
		case cam.TLSClientAuth != nil:
			method = "tls_client_auth"
			for _, r := range cam.TLSClientAuth.Credentials {
				if r != nil {
					refIDs = append(refIDs, r.ID)
				}
			}
		case cam.SelfSignedTLSClientAuth != nil:
			method = "self_signed_tls_client_auth"
			for _, r := range cam.SelfSignedTLSClientAuth.Credentials {
				if r != nil {
					refIDs = append(refIDs, r.ID)
				}
			}
		}
		state.AuthenticationMethod = types.StringValue(method)

		credObjs := make([]attr.Value, 0, len(refIDs))
		for _, id := range refIDs {
			detail, err := d.mgmt.Clients.Credentials.Get(ctx, clientID, id)
			if err != nil {
				if framework.IsNotFound(err) {
					continue
				}
				framework.AddAPIError(&resp.Diagnostics, "Failed to read credential "+id, err)
				return
			}
			obj, d := dsCredentialToObject(detail)
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			credObjs = append(credObjs, obj)
		}
		listVal, d := types.ListValue(types.ObjectType{AttrTypes: dsCredentialAttrTypes()}, credObjs)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Credentials = listVal
	} else {
		method := "client_secret_post"
		if got.TokenEndpointAuthMethod != nil {
			method = string(*got.TokenEndpointAuthMethod)
		}
		state.AuthenticationMethod = types.StringValue(method)
		state.Credentials = types.ListValueMust(types.ObjectType{AttrTypes: dsCredentialAttrTypes()}, []attr.Value{})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// dsCredentialToObject converts a GetClientCredentialResponseContent into the
// data-source object shape (without write-only fields).
func dsCredentialToObject(d *mgmt.GetClientCredentialResponseContent) (types.Object, diag.Diagnostics) {
	ce := credentialEntryFromGetResponse(d)
	strOrNull := func(s string) attr.Value {
		if s == "" {
			return types.StringNull()
		}
		return types.StringValue(s)
	}
	vals := map[string]attr.Value{
		"id":                strOrNull(ce.ID),
		"name":              strOrNull(ce.Name),
		"credential_type":   strOrNull(ce.CredentialType),
		"algorithm":         strOrNull(ce.Algorithm),
		"kid":               strOrNull(ce.Kid),
		"subject_dn":        strOrNull(ce.SubjectDn),
		"thumbprint_sha256": strOrNull(ce.ThumbprintSha256),
		"expires_at":        strOrNull(ce.ExpiresAt),
		"created_at":        strOrNull(ce.CreatedAt),
		"updated_at":        strOrNull(ce.UpdatedAt),
	}
	return types.ObjectValue(dsCredentialAttrTypes(), vals)
}
