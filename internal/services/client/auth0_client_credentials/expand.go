package auth0clientcredentials

import (
	"context"
	"fmt"
	"time"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// credentialEntry is the plain-Go projection of one credential block element.
// It mirrors what we accept from HCL and what we send to the API.
type credentialEntry struct {
	ID                  string
	Name                string
	CredentialType      string
	Pem                 string
	Algorithm           string
	ParseExpiryFromCert bool
	ExpiresAt           string // RFC3339
	Kid                 string
	SubjectDn           string
	ThumbprintSha256    string
	CreatedAt           string
	UpdatedAt           string
}

// credentialEntriesFromList unpacks the single-element wrapper list and its
// inner credentials list into a slice of credentialEntry.
func credentialEntriesFromList(ctx context.Context, l types.List) ([]credentialEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	if l.IsNull() || l.IsUnknown() {
		return nil, diags
	}
	elems := l.Elements()
	if len(elems) == 0 {
		return nil, diags
	}
	// The outer list holds one wrapper object with a `credentials` list inside.
	wrapper, ok := elems[0].(types.Object)
	if !ok {
		diags.AddError("Internal type error", fmt.Sprintf("expected types.Object, got %T", elems[0]))
		return nil, diags
	}
	attrs := wrapper.Attributes()
	credsAttr, ok := attrs["credentials"].(types.List)
	if !ok || credsAttr.IsNull() || credsAttr.IsUnknown() {
		return nil, diags
	}
	var out []credentialEntry
	for i, e := range credsAttr.Elements() {
		obj, ok := e.(types.Object)
		if !ok {
			diags.AddError("Internal type error", fmt.Sprintf("credentials[%d] is not an object (%T)", i, e))
			return nil, diags
		}
		ce, d := credentialEntryFromObject(obj)
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		out = append(out, ce)
	}
	return out, diags
}

func credentialEntryFromObject(o types.Object) (credentialEntry, diag.Diagnostics) {
	var diags diag.Diagnostics
	a := o.Attributes()
	ce := credentialEntry{
		ID:               stringFromAttr(a, "id"),
		Name:             stringFromAttr(a, "name"),
		CredentialType:   stringFromAttr(a, "credential_type"),
		Pem:              stringFromAttr(a, "pem"),
		Algorithm:        stringFromAttr(a, "algorithm"),
		ExpiresAt:        stringFromAttr(a, "expires_at"),
		Kid:              stringFromAttr(a, "kid"),
		SubjectDn:        stringFromAttr(a, "subject_dn"),
		ThumbprintSha256: stringFromAttr(a, "thumbprint_sha256"),
		CreatedAt:        stringFromAttr(a, "created_at"),
		UpdatedAt:        stringFromAttr(a, "updated_at"),
	}
	if b, ok := a["parse_expiry_from_cert"].(types.Bool); ok && !b.IsNull() && !b.IsUnknown() {
		ce.ParseExpiryFromCert = b.ValueBool()
	}
	return ce, diags
}

func stringFromAttr(a map[string]attr.Value, key string) string {
	v, ok := a[key].(types.String)
	if !ok || v.IsNull() || v.IsUnknown() {
		return ""
	}
	return v.ValueString()
}

// buildPostCredentialBody converts a credentialEntry into the SDK request body
// for `POST /clients/{id}/credentials`.
func buildPostCredentialBody(ce credentialEntry) (*mgmt.PostClientCredentialRequestContent, diag.Diagnostics) {
	var diags diag.Diagnostics
	enum, err := mgmt.NewClientCredentialTypeEnumFromString(ce.CredentialType)
	if err != nil {
		diags.AddError("Invalid credential_type", err.Error())
		return nil, diags
	}
	body := &mgmt.PostClientCredentialRequestContent{
		CredentialType: enum,
	}
	if ce.Name != "" {
		body.Name = &ce.Name
	}
	if ce.SubjectDn != "" {
		body.SubjectDn = &ce.SubjectDn
	}
	if ce.Pem != "" {
		body.Pem = &ce.Pem
	}
	if ce.Algorithm != "" {
		alg, err := mgmt.NewPublicKeyCredentialAlgorithmEnumFromString(ce.Algorithm)
		if err != nil {
			diags.AddError("Invalid algorithm", err.Error())
			return nil, diags
		}
		body.Alg = &alg
	}
	if ce.ParseExpiryFromCert {
		v := ce.ParseExpiryFromCert
		body.ParseExpiryFromCert = &v
	}
	if ce.ExpiresAt != "" {
		ts, err := parseTimeRFC3339(ce.ExpiresAt)
		if err != nil {
			diags.AddError("Invalid expires_at", err.Error())
			return nil, diags
		}
		body.ExpiresAt = &ts
	}
	if ce.Kid != "" {
		body.Kid = &ce.Kid
	}
	return body, diags
}

// buildClientAuthMethodsForRefs constructs the `client_authentication_methods`
// PATCH body for the given method and credential IDs.
func buildClientAuthMethodsForRefs(method string, ids []string) *mgmt.ClientAuthenticationMethod {
	credRefs := make([]*mgmt.CredentialID, 0, len(ids))
	for _, id := range ids {
		credRefs = append(credRefs, &mgmt.CredentialID{ID: id})
	}
	cam := &mgmt.ClientAuthenticationMethod{}
	switch method {
	case "private_key_jwt":
		cam.PrivateKeyJwt = &mgmt.ClientAuthenticationMethodPrivateKeyJwt{
			Credentials: credRefs,
		}
	case "tls_client_auth":
		cam.TLSClientAuth = &mgmt.ClientAuthenticationMethodTLSClientAuth{
			Credentials: credRefs,
		}
	case "self_signed_tls_client_auth":
		cam.SelfSignedTLSClientAuth = &mgmt.ClientAuthenticationMethodSelfSignedTLSClientAuth{
			Credentials: credRefs,
		}
	}
	return cam
}

// credentialsListForMethod returns the model field that holds the credentials
// for the given auth method.
func credentialsListForMethod(method string, m model) types.List {
	switch method {
	case "private_key_jwt":
		return m.PrivateKeyJWT
	case "tls_client_auth":
		return m.TLSClientAuth
	case "self_signed_tls_client_auth":
		return m.SelfSignedTLSClientAuth
	}
	return types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
}

func setCredentialsListForMethod(method string, m *model, l types.List) {
	switch method {
	case "private_key_jwt":
		m.PrivateKeyJWT = l
	case "tls_client_auth":
		m.TLSClientAuth = l
	case "self_signed_tls_client_auth":
		m.SelfSignedTLSClientAuth = l
	}
}

// nullifyOtherBlocks marks the credential blocks for the *other* two methods
// as null in state, so Terraform sees a clean transition.
func nullifyOtherBlocks(method string, m *model) {
	nullList := types.ListNull(types.ObjectType{AttrTypes: credentialBlockAttrTypes()})
	if method != "private_key_jwt" {
		m.PrivateKeyJWT = nullList
	}
	if method != "tls_client_auth" {
		m.TLSClientAuth = nullList
	}
	if method != "self_signed_tls_client_auth" {
		m.SelfSignedTLSClientAuth = nullList
	}
}

// validateMethodAndBlocks enforces that the right credential block is
// populated for the chosen authentication method.
func validateMethodAndBlocks(method string, plan model) error {
	if !isSecretBased(method) && !isCredentialBased(method) {
		return fmt.Errorf("unsupported authentication_method %q", method)
	}
	hasBlock := func(l types.List) bool {
		return !l.IsNull() && !l.IsUnknown() && len(l.Elements()) > 0
	}
	if isSecretBased(method) {
		if hasBlock(plan.PrivateKeyJWT) || hasBlock(plan.TLSClientAuth) || hasBlock(plan.SelfSignedTLSClientAuth) {
			return fmt.Errorf("authentication_method %q must not declare any credential block", method)
		}
		return nil
	}
	// Credential-based: the matching block must be set, the others must not.
	pjwt := hasBlock(plan.PrivateKeyJWT)
	mtls := hasBlock(plan.TLSClientAuth)
	smtls := hasBlock(plan.SelfSignedTLSClientAuth)
	switch method {
	case "private_key_jwt":
		if !pjwt || mtls || smtls {
			return fmt.Errorf("authentication_method = %q requires the `private_key_jwt` block to be set, and the other credential blocks to be unset", method)
		}
	case "tls_client_auth":
		if !mtls || pjwt || smtls {
			return fmt.Errorf("authentication_method = %q requires the `tls_client_auth` block to be set, and the other credential blocks to be unset", method)
		}
	case "self_signed_tls_client_auth":
		if !smtls || pjwt || mtls {
			return fmt.Errorf("authentication_method = %q requires the `self_signed_tls_client_auth` block to be set, and the other credential blocks to be unset", method)
		}
	}
	return nil
}

// credentialImmutableEqual returns true when two credentialEntry values share
// the same immutable identity (the API treats these as not updatable; only
// `expires_at` can be patched).
func credentialImmutableEqual(a, b credentialEntry) bool {
	return a.CredentialType == b.CredentialType &&
		a.Pem == b.Pem &&
		a.Algorithm == b.Algorithm &&
		a.SubjectDn == b.SubjectDn &&
		a.Kid == b.Kid &&
		a.ParseExpiryFromCert == b.ParseExpiryFromCert &&
		a.Name == b.Name
}

func stringListToGo(l types.List) []string {
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	out := make([]string, 0, len(l.Elements()))
	for _, e := range l.Elements() {
		if s, ok := e.(types.String); ok && !s.IsNull() && !s.IsUnknown() {
			out = append(out, s.ValueString())
		}
	}
	return out
}

func stringSet(in []string) map[string]struct{} {
	s := make(map[string]struct{}, len(in))
	for _, v := range in {
		s[v] = struct{}{}
	}
	return s
}

func parseTimeRFC3339(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}
