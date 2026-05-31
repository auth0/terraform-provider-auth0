package auth0clientgrant

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// expandCreate converts the Terraform plan into a CreateClientGrantRequestContent.
func expandCreate(ctx context.Context, plan *clientGrantResourceModel) (*mgmt.CreateClientGrantRequestContent, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &mgmt.CreateClientGrantRequestContent{
		Audience: plan.Audience.ValueString(),
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() {
		v := plan.ClientID.ValueString()
		body.ClientID = &v
	}

	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		body.Scope = framework.StringListToGo(ctx, plan.Scopes, &diags)
	}
	if !plan.AuthorizationDetailsTypes.IsNull() && !plan.AuthorizationDetailsTypes.IsUnknown() {
		body.AuthorizationDetailsTypes = framework.StringListToGo(ctx, plan.AuthorizationDetailsTypes, &diags)
	}

	if !plan.AllowAnyOrganization.IsNull() && !plan.AllowAnyOrganization.IsUnknown() {
		v := plan.AllowAnyOrganization.ValueBool()
		body.AllowAnyOrganization = &v
	}
	if !plan.AllowAllScopes.IsNull() && !plan.AllowAllScopes.IsUnknown() {
		v := plan.AllowAllScopes.ValueBool()
		body.AllowAllScopes = &v
	}

	if !plan.OrganizationUsage.IsNull() && !plan.OrganizationUsage.IsUnknown() {
		enum, err := mgmt.NewClientGrantOrganizationUsageEnumFromString(plan.OrganizationUsage.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_usage"), "Invalid organization_usage", err.Error())
			return nil, diags
		}
		body.OrganizationUsage = &enum
	}

	if !plan.SubjectType.IsNull() && !plan.SubjectType.IsUnknown() {
		enum, err := mgmt.NewClientGrantSubjectTypeEnumFromString(plan.SubjectType.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("subject_type"), "Invalid subject_type", err.Error())
			return nil, diags
		}
		body.SubjectType = &enum
	}

	if !plan.DefaultFor.IsNull() && !plan.DefaultFor.IsUnknown() {
		enum, err := mgmt.NewClientGrantDefaultForEnumFromString(plan.DefaultFor.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("default_for"), "Invalid default_for", err.Error())
			return nil, diags
		}
		body.DefaultFor = &enum
	}

	return body, diags
}

// expandUpdate converts the Terraform plan into an UpdateClientGrantRequestContent.
// Only mutable fields are included.
func expandUpdate(ctx context.Context, plan *clientGrantResourceModel) (*mgmt.UpdateClientGrantRequestContent, diag.Diagnostics) {
	var diags diag.Diagnostics

	body := &mgmt.UpdateClientGrantRequestContent{}

	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		body.Scope = framework.StringListToGo(ctx, plan.Scopes, &diags)
	}
	if !plan.AuthorizationDetailsTypes.IsNull() && !plan.AuthorizationDetailsTypes.IsUnknown() {
		body.AuthorizationDetailsTypes = framework.StringListToGo(ctx, plan.AuthorizationDetailsTypes, &diags)
	}
	if !plan.AllowAnyOrganization.IsNull() && !plan.AllowAnyOrganization.IsUnknown() {
		v := plan.AllowAnyOrganization.ValueBool()
		body.AllowAnyOrganization = &v
	}
	if !plan.AllowAllScopes.IsNull() && !plan.AllowAllScopes.IsUnknown() {
		v := plan.AllowAllScopes.ValueBool()
		body.AllowAllScopes = &v
	}

	if !plan.OrganizationUsage.IsNull() && !plan.OrganizationUsage.IsUnknown() {
		enum, err := mgmt.NewClientGrantOrganizationNullableUsageEnumFromString(plan.OrganizationUsage.ValueString())
		if err != nil {
			diags.AddAttributeError(path.Root("organization_usage"), "Invalid organization_usage", err.Error())
			return nil, diags
		}
		body.OrganizationUsage = &enum
	}

	return body, diags
}
