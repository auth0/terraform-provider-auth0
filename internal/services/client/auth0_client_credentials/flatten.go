package auth0clientcredentials

import (
	"context"
	"fmt"
	"time"

	mgmt "github.com/auth0/go-auth0/v2/management"
	mgmtclient "github.com/auth0/go-auth0/v2/management/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// refreshStateFromClient walks the response of `GET /clients/{id}` and
// populates `state` with the currently-active authentication method and any
// credentials it references.
//
// Notes about read-only behavior:
//   - `pem`, `parse_expiry_from_cert` are write-only on the API. We preserve
//     them from the prior state.
//   - When the client is configured with a secret-based method, all credential
//     blocks are nullified in state.
func refreshStateFromClient(ctx context.Context, m *mgmtclient.Management, clientID string, c *mgmt.GetClientResponseContent, state *model) error {
	// Cache the prior plain entries by ID so we can preserve write-only fields.
	priorEntriesByMethod := map[string]map[string]credentialEntry{}
	for _, method := range []string{"private_key_jwt", "tls_client_auth", "self_signed_tls_client_auth"} {
		list := credentialsListForMethod(method, *state)
		entries, _ := credentialEntriesFromList(ctx, list)
		byID := map[string]credentialEntry{}
		for _, e := range entries {
			if e.ID != "" {
				byID[e.ID] = e
			}
		}
		priorEntriesByMethod[method] = byID
	}

	// Determine the active method.
	cam := c.ClientAuthenticationMethods
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

		// Fetch each credential and merge with prior write-only fields.
		var entries []credentialEntry
		for _, id := range refIDs {
			detail, err := m.Clients.Credentials.Get(ctx, clientID, id)
			if err != nil {
				if framework.IsNotFound(err) {
					// Credential disappeared out-of-band; skip it.
					continue
				}
				return fmt.Errorf("failed to read credential %s: %w", id, err)
			}
			ce := credentialEntryFromGetResponse(detail)
			// Preserve write-only fields from prior state if available.
			if prior, ok := priorEntriesByMethod[method][id]; ok {
				if ce.Pem == "" {
					ce.Pem = prior.Pem
				}
				if !ce.ParseExpiryFromCert && prior.ParseExpiryFromCert {
					ce.ParseExpiryFromCert = prior.ParseExpiryFromCert
				}
			}
			entries = append(entries, ce)
		}

		list, d := credentialEntriesToList(entries)
		if d.HasError() {
			return fmt.Errorf("failed to flatten credentials: %s", d.Errors()[0].Detail())
		}
		setCredentialsListForMethod(method, state, list)
		nullifyOtherBlocks(method, state)
		state.ManagedCredentialIDs = framework.StringSliceToList(refIDs)
		return nil
	}

	// No client_authentication_methods set — fall back to token_endpoint_auth_method.
	method := "client_secret_post" // sensible default if the API didn't return one
	if c.TokenEndpointAuthMethod != nil {
		method = string(*c.TokenEndpointAuthMethod)
	}
	state.AuthenticationMethod = types.StringValue(method)
	state.PrivateKeyJWT = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	state.TLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	state.SelfSignedTLSClientAuth = types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	state.ManagedCredentialIDs = framework.StringSliceToList(nil)
	return nil
}

// credentialEntryFromGetResponse converts the SDK GetClientCredentialResponseContent
// into a credentialEntry.
func credentialEntryFromGetResponse(d *mgmt.GetClientCredentialResponseContent) credentialEntry {
	if d == nil {
		return credentialEntry{}
	}
	ce := credentialEntry{}
	if d.ID != nil {
		ce.ID = *d.ID
	}
	if d.Name != nil {
		ce.Name = *d.Name
	}
	if d.Kid != nil {
		ce.Kid = *d.Kid
	}
	if d.Alg != nil {
		ce.Algorithm = string(*d.Alg)
	}
	if d.CredentialType != nil {
		ce.CredentialType = string(*d.CredentialType)
	}
	if d.SubjectDn != nil {
		ce.SubjectDn = *d.SubjectDn
	}
	if d.ThumbprintSha256 != nil {
		ce.ThumbprintSha256 = *d.ThumbprintSha256
	}
	if d.CreatedAt != nil {
		ce.CreatedAt = d.CreatedAt.UTC().Format(time.RFC3339)
	}
	if d.UpdatedAt != nil {
		ce.UpdatedAt = d.UpdatedAt.UTC().Format(time.RFC3339)
	}
	if d.ExpiresAt != nil {
		ce.ExpiresAt = d.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return ce
}

// credentialEntryFromPostResponse converts a POST response into credentialEntry.
func credentialEntryFromPostResponse(d *mgmt.PostClientCredentialResponseContent) credentialEntry {
	if d == nil {
		return credentialEntry{}
	}
	ce := credentialEntry{}
	if d.ID != nil {
		ce.ID = *d.ID
	}
	if d.Name != nil {
		ce.Name = *d.Name
	}
	if d.Kid != nil {
		ce.Kid = *d.Kid
	}
	if d.Alg != nil {
		ce.Algorithm = string(*d.Alg)
	}
	if d.CredentialType != nil {
		ce.CredentialType = string(*d.CredentialType)
	}
	if d.SubjectDn != nil {
		ce.SubjectDn = *d.SubjectDn
	}
	if d.ThumbprintSha256 != nil {
		ce.ThumbprintSha256 = *d.ThumbprintSha256
	}
	if d.CreatedAt != nil {
		ce.CreatedAt = d.CreatedAt.UTC().Format(time.RFC3339)
	}
	if d.UpdatedAt != nil {
		ce.UpdatedAt = d.UpdatedAt.UTC().Format(time.RFC3339)
	}
	if d.ExpiresAt != nil {
		ce.ExpiresAt = d.ExpiresAt.UTC().Format(time.RFC3339)
	}
	return ce
}

// credentialEntryToObject converts a credentialEntry into a types.Object using
// the canonical attribute schema. Empty strings are encoded as null.
func credentialEntryToObject(ce credentialEntry) (types.Object, diag.Diagnostics) {
	strOrNull := func(s string) attr.Value {
		if s == "" {
			return types.StringNull()
		}
		return types.StringValue(s)
	}
	vals := map[string]attr.Value{
		"id":                     strOrNull(ce.ID),
		"name":                   strOrNull(ce.Name),
		"credential_type":        strOrNull(ce.CredentialType),
		"pem":                    strOrNull(ce.Pem),
		"algorithm":              strOrNull(ce.Algorithm),
		"parse_expiry_from_cert": types.BoolValue(ce.ParseExpiryFromCert),
		"expires_at":             strOrNull(ce.ExpiresAt),
		"kid":                    strOrNull(ce.Kid),
		"subject_dn":             strOrNull(ce.SubjectDn),
		"thumbprint_sha256":      strOrNull(ce.ThumbprintSha256),
		"created_at":             strOrNull(ce.CreatedAt),
		"updated_at":             strOrNull(ce.UpdatedAt),
	}
	return types.ObjectValue(credentialAttrTypes(), vals)
}

// credentialEntriesToList wraps a slice of credentialEntry as the outer
// single-element list of wrapper objects (the schema shape).
func credentialEntriesToList(entries []credentialEntry) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	innerObjs := make([]attr.Value, 0, len(entries))
	for _, ce := range entries {
		obj, d := credentialEntryToObject(ce)
		diags.Append(d...)
		if diags.HasError() {
			return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
		}
		innerObjs = append(innerObjs, obj)
	}
	innerList, d := types.ListValue(types.ObjectType{AttrTypes: credentialAttrTypes()}, innerObjs)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
	}
	wrapper, d := types.ObjectValue(credentialBlockAttrTypes(), map[string]attr.Value{
		"credentials": innerList,
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
	}
	outer, d := types.ListValue(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}, []attr.Value{wrapper})
	diags.Append(d...)
	return outer, diags
}

// flattenCreatedCredentialsIntoPlan re-shapes the plan's credential block
// after a Create call, merging API-returned fields (id, kid, created_at, …)
// into each credential, while preserving user-supplied write-only fields
// (`pem`, `parse_expiry_from_cert`).
func flattenCreatedCredentialsIntoPlan(ctx context.Context, planList types.List, created []*mgmt.PostClientCredentialResponseContent) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	planEntries, d := credentialEntriesFromList(ctx, planList)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
	}
	if len(planEntries) != len(created) {
		diags.AddError(
			"Internal flatten error",
			fmt.Sprintf("plan declared %d credentials but %d were created", len(planEntries), len(created)),
		)
		return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
	}
	for i := range planEntries {
		merged := mergeCredentialEntry(planEntries[i], credentialEntryFromPostResponse(created[i]))
		planEntries[i] = merged
	}
	return credentialEntriesToList(planEntries)
}

// flattenUpdatedCredentialsIntoPlan handles the update path. Indexes present
// in `newCredsByIndex` were newly created in this Update; other entries are
// preserved from the plan (which carries forward the prior state's ID/kid via
// UseStateForUnknown).
func flattenUpdatedCredentialsIntoPlan(ctx context.Context, planList types.List, newCredsByIndex map[int]*mgmt.PostClientCredentialResponseContent) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	planEntries, d := credentialEntriesFromList(ctx, planList)
	diags.Append(d...)
	if diags.HasError() {
		return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()}), diags
	}
	for i := range planEntries {
		if c, ok := newCredsByIndex[i]; ok {
			planEntries[i] = mergeCredentialEntry(planEntries[i], credentialEntryFromPostResponse(c))
		}
	}
	return credentialEntriesToList(planEntries)
}

// mergeCredentialEntry overlays API-returned values onto a planned entry,
// keeping user-supplied write-only fields (pem, parse_expiry_from_cert) intact.
func mergeCredentialEntry(plan, api credentialEntry) credentialEntry {
	if api.ID != "" {
		plan.ID = api.ID
	}
	if api.Name != "" {
		plan.Name = api.Name
	}
	if api.Kid != "" {
		plan.Kid = api.Kid
	}
	if api.Algorithm != "" {
		plan.Algorithm = api.Algorithm
	}
	if api.CredentialType != "" {
		plan.CredentialType = api.CredentialType
	}
	if api.SubjectDn != "" {
		plan.SubjectDn = api.SubjectDn
	}
	if api.ThumbprintSha256 != "" {
		plan.ThumbprintSha256 = api.ThumbprintSha256
	}
	if api.CreatedAt != "" {
		plan.CreatedAt = api.CreatedAt
	}
	if api.UpdatedAt != "" {
		plan.UpdatedAt = api.UpdatedAt
	}
	if api.ExpiresAt != "" {
		plan.ExpiresAt = api.ExpiresAt
	}
	// pem and parse_expiry_from_cert are intentionally NOT overwritten.
	return plan
}
