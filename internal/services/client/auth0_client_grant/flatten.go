package auth0clientgrant

import (
	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// commonClientGrantFields is the set of fields shared across the
// Create / Get / Update response shapes (which are all functionally
// identical for client grants).
type commonClientGrantFields struct {
	ID                        *string
	ClientID                  *string
	Audience                  *string
	Scope                     []string
	OrganizationUsage         *mgmt.ClientGrantOrganizationUsageEnum
	AllowAnyOrganization      *bool
	DefaultFor                *mgmt.ClientGrantDefaultForEnum
	IsSystem                  *bool
	SubjectType               *mgmt.ClientGrantSubjectTypeEnum
	AuthorizationDetailsTypes []string
	AllowAllScopes            *bool
}

func flattenInto(m *clientGrantResourceModel, c commonClientGrantFields) {
	m.ID = types.StringPointerValue(c.ID)
	m.ClientID = types.StringPointerValue(c.ClientID)
	m.Audience = types.StringPointerValue(c.Audience)
	m.Scopes = framework.StringSliceToList(c.Scope)
	m.OrganizationUsage = framework.EnumPtrToString(c.OrganizationUsage)
	m.AllowAnyOrganization = types.BoolPointerValue(c.AllowAnyOrganization)
	m.DefaultFor = framework.EnumPtrToString(c.DefaultFor)
	m.IsSystem = types.BoolPointerValue(c.IsSystem)
	m.SubjectType = framework.EnumPtrToString(c.SubjectType)
	m.AuthorizationDetailsTypes = framework.StringSliceToList(c.AuthorizationDetailsTypes)
	m.AllowAllScopes = types.BoolPointerValue(c.AllowAllScopes)
}

func flattenCreate(m *clientGrantResourceModel, c *mgmt.CreateClientGrantResponseContent) {
	flattenInto(m, commonClientGrantFields{
		ID:                        c.ID,
		ClientID:                  c.ClientID,
		Audience:                  c.Audience,
		Scope:                     c.Scope,
		OrganizationUsage:         c.OrganizationUsage,
		AllowAnyOrganization:      c.AllowAnyOrganization,
		DefaultFor:                c.DefaultFor,
		IsSystem:                  c.IsSystem,
		SubjectType:               c.SubjectType,
		AuthorizationDetailsTypes: c.AuthorizationDetailsTypes,
		AllowAllScopes:            c.AllowAllScopes,
	})
}

func flattenGet(m *clientGrantResourceModel, c *mgmt.GetClientGrantResponseContent) {
	flattenInto(m, commonClientGrantFields{
		ID:                        c.ID,
		ClientID:                  c.ClientID,
		Audience:                  c.Audience,
		Scope:                     c.Scope,
		OrganizationUsage:         c.OrganizationUsage,
		AllowAnyOrganization:      c.AllowAnyOrganization,
		DefaultFor:                c.DefaultFor,
		IsSystem:                  c.IsSystem,
		SubjectType:               c.SubjectType,
		AuthorizationDetailsTypes: c.AuthorizationDetailsTypes,
		AllowAllScopes:            c.AllowAllScopes,
	})
}

func flattenUpdate(m *clientGrantResourceModel, c *mgmt.UpdateClientGrantResponseContent) {
	flattenInto(m, commonClientGrantFields{
		ID:                        c.ID,
		ClientID:                  c.ClientID,
		Audience:                  c.Audience,
		Scope:                     c.Scope,
		OrganizationUsage:         c.OrganizationUsage,
		AllowAnyOrganization:      c.AllowAnyOrganization,
		DefaultFor:                c.DefaultFor,
		IsSystem:                  c.IsSystem,
		SubjectType:               c.SubjectType,
		AuthorizationDetailsTypes: c.AuthorizationDetailsTypes,
		AllowAllScopes:            c.AllowAllScopes,
	})
}
