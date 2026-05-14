// addons.go — typed schema + expand/flatten for the auth0_client `addons`
// attribute. The Auth0 Management API ships ~30 distinct addon types; each
// one has its own JSON shape. To keep this file readable:
//
//   - a single SingleNestedAttribute declaration in `addonsSchemaAttribute`
//     enumerates every addon and its fields, grouped by section banner;
//   - `expandAddons` and `flattenAddons` walk the typed model and the SDK
//     struct field by field — one block per addon, each block trivially
//     small thanks to a few generic helpers at the bottom of the file;
//   - a per-addon attr-type-map function makes nested object construction
//     a one-liner; helpers `singleStringAttrTypes` and `singleBoolAttrTypes`
//     cover the common shapes.
//
// Adding a new Auth0 addon is a 4-step recipe:
//
//  1. extend the model field below with the new types.Object;
//  2. add an entry under the appropriate section banner in
//     `addonsSchemaAttribute`;
//  3. add an `if !plan.<X>.IsNull() { ... }` block in `expandAddons`;
//  4. set `out.<X> = flatten...` in `flattenAddons`.
//
// Hybrid escape hatch (planned): an `addons_extra_json` attribute will
// be kept as a JSON string for any future addon Auth0 ships before this
// file is updated.
package client

import (
	"context"

	mgmt "github.com/auth0/go-auth0/v2/management"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/auth0/terraform-provider-auth0/v2/internal/framework"
)

// =========================================================================
// model — one types.Object per addon.
// =========================================================================

type addonsModel struct {
	AWS                  types.Object `tfsdk:"aws"`
	AzureBlob            types.Object `tfsdk:"azure_blob"`
	AzureSb              types.Object `tfsdk:"azure_sb"`
	Box                  types.Object `tfsdk:"box"`
	Cloudbees            types.Object `tfsdk:"cloudbees"`
	Concur               types.Object `tfsdk:"concur"`
	Dropbox              types.Object `tfsdk:"dropbox"`
	Echosign             types.Object `tfsdk:"echosign"`
	Egnyte               types.Object `tfsdk:"egnyte"`
	Firebase             types.Object `tfsdk:"firebase"`
	Layer                types.Object `tfsdk:"layer"`
	Mscrm                types.Object `tfsdk:"mscrm"`
	Newrelic             types.Object `tfsdk:"newrelic"`
	Office365            types.Object `tfsdk:"office365"`
	Oag                  types.Object `tfsdk:"oag"`
	Rms                  types.Object `tfsdk:"rms"`
	Salesforce           types.Object `tfsdk:"salesforce"`
	SalesforceAPI        types.Object `tfsdk:"salesforce_api"`
	SalesforceSandboxAPI types.Object `tfsdk:"salesforce_sandbox_api"`
	Samlp                types.Object `tfsdk:"samlp"`
	SapAPI               types.Object `tfsdk:"sap_api"`
	Sentry               types.Object `tfsdk:"sentry"`
	Sharepoint           types.Object `tfsdk:"sharepoint"`
	Slack                types.Object `tfsdk:"slack"`
	Springcm             types.Object `tfsdk:"springcm"`
	SSOIntegration       types.Object `tfsdk:"sso_integration"`
	Wams                 types.Object `tfsdk:"wams"`
	Wsfed                types.Object `tfsdk:"wsfed"`
	Zendesk              types.Object `tfsdk:"zendesk"`
	Zoom                 types.Object `tfsdk:"zoom"`
}

// addonsAttrTypes is the attr.Type map that mirrors `addonsModel` — needed
// when constructing or returning a typed Object value.
func addonsAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"aws":                    types.ObjectType{AttrTypes: awsAddonAttrTypes()},
		"azure_blob":             types.ObjectType{AttrTypes: azureBlobAddonAttrTypes()},
		"azure_sb":               types.ObjectType{AttrTypes: azureSbAddonAttrTypes()},
		"box":                    types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"cloudbees":              types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"concur":                 types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"dropbox":                types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"echosign":               types.ObjectType{AttrTypes: singleStringAttrTypes("domain")},
		"egnyte":                 types.ObjectType{AttrTypes: singleStringAttrTypes("domain")},
		"firebase":               types.ObjectType{AttrTypes: firebaseAddonAttrTypes()},
		"layer":                  types.ObjectType{AttrTypes: layerAddonAttrTypes()},
		"mscrm":                  types.ObjectType{AttrTypes: singleStringAttrTypes("url")},
		"newrelic":               types.ObjectType{AttrTypes: singleStringAttrTypes("account")},
		"office365":              types.ObjectType{AttrTypes: office365AddonAttrTypes()},
		"oag":                    types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"rms":                    types.ObjectType{AttrTypes: singleStringAttrTypes("url")},
		"salesforce":             types.ObjectType{AttrTypes: singleStringAttrTypes("entity_id")},
		"salesforce_api":         types.ObjectType{AttrTypes: salesforceAPIAttrTypes()},
		"salesforce_sandbox_api": types.ObjectType{AttrTypes: salesforceAPIAttrTypes()},
		"samlp":                  types.ObjectType{AttrTypes: samlpAddonAttrTypes()},
		"sap_api":                types.ObjectType{AttrTypes: sapAPIAttrTypes()},
		"sentry":                 types.ObjectType{AttrTypes: sentryAddonAttrTypes()},
		"sharepoint":             types.ObjectType{AttrTypes: sharepointAddonAttrTypes()},
		"slack":                  types.ObjectType{AttrTypes: singleStringAttrTypes("team")},
		"springcm":               types.ObjectType{AttrTypes: singleStringAttrTypes("acsurl")},
		"sso_integration":        types.ObjectType{AttrTypes: ssoIntegrationAttrTypes()},
		"wams":                   types.ObjectType{AttrTypes: singleStringAttrTypes("masterkey")},
		"wsfed":                  types.ObjectType{AttrTypes: emptyAddonAttrTypes()},
		"zendesk":                types.ObjectType{AttrTypes: singleStringAttrTypes("account_name")},
		"zoom":                   types.ObjectType{AttrTypes: singleStringAttrTypes("account")},
	}
}

// =========================================================================
// schema — one big SingleNestedAttribute, sectioned by banner comment.
// =========================================================================

func addonsSchemaAttribute() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Application add-ons. Configure any combination of the supported Auth0 add-ons. See https://auth0.com/docs/configure/applications/addons.",
		Attributes: map[string]schema.Attribute{
			// ----- AWS / cloud storage -----------------------------------
			"aws":        awsAddonSchema(),
			"azure_blob": azureBlobAddonSchema(),
			"azure_sb":   azureSbAddonSchema(),

			// ----- Empty (flag-only) addons ------------------------------
			"box":       emptyAddonSchema("Box"),
			"cloudbees": emptyAddonSchema("CloudBees"),
			"concur":    emptyAddonSchema("Concur"),
			"dropbox":   emptyAddonSchema("Dropbox"),
			"oag":       emptyAddonSchema("Oracle Access Gateway"),
			"wsfed":     emptyAddonSchema("WS-Federation"),

			// ----- Single-string addons ----------------------------------
			"echosign":   singleStringAddonSchema("EchoSign", "domain", "Adobe Sign account domain (e.g. `your-account`)."),
			"egnyte":     singleStringAddonSchema("Egnyte", "domain", "Egnyte tenant domain (e.g. `your-account`)."),
			"mscrm":      singleStringAddonSchema("Microsoft Dynamics CRM", "url", "Dynamics CRM URL."),
			"newrelic":   singleStringAddonSchema("New Relic", "account", "New Relic account ID."),
			"rms":        singleStringAddonSchema("Active Directory RMS", "url", "RMS server URL."),
			"salesforce": singleStringAddonSchema("Salesforce", "entity_id", "Salesforce SAML Entity ID."),
			"slack":      singleStringAddonSchema("Slack", "team", "Slack team name (e.g. `your-team`)."),
			"springcm":   singleStringAddonSchema("SpringCM", "acsurl", "SpringCM ACS URL."),
			"wams":       singleStringAddonSchema("Windows Azure Mobile Services", "masterkey", "Master key for the Mobile Services app."),
			"zendesk":    singleStringAddonSchema("Zendesk", "account_name", "Zendesk account name."),
			"zoom":       singleStringAddonSchema("Zoom", "account", "Zoom account name."),

			// ----- Firebase / Google ------------------------------------
			"firebase": firebaseAddonSchema(),

			// ----- SAML-family ------------------------------------------
			"samlp":                  samlpAddonSchema(),
			"office365":              office365AddonSchema(),
			"sharepoint":             sharepointAddonSchema(),
			"salesforce_api":         salesforceAPIAddonSchema("Salesforce API"),
			"salesforce_sandbox_api": salesforceAPIAddonSchema("Salesforce Sandbox API"),
			"sap_api":                sapAPIAddonSchema(),

			// ----- Misc -------------------------------------------------
			"layer":           layerAddonSchema(),
			"sentry":          sentryAddonSchema(),
			"sso_integration": ssoIntegrationAddonSchema(),
		},
	}
}

// =========================================================================
// expand: TF model -> *mgmt.ClientAddons
// =========================================================================

func expandAddons(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientAddons {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	var m addonsModel
	if d := o.As(ctx, &m, framework.ObjectAsOpts()); d.HasError() {
		diags.Append(d...)
		return nil
	}
	out := &mgmt.ClientAddons{}

	// ----- AWS / cloud storage -----------------------------------------
	if v := expandAWSAddon(m.AWS); v != nil {
		out.Aws = v
	}
	if v := expandAzureBlobAddon(m.AzureBlob); v != nil {
		out.AzureBlob = v
	}
	if v := expandAzureSbAddon(m.AzureSb); v != nil {
		out.AzureSb = v
	}

	// ----- Empty (flag-only) addons ------------------------------------
	// The SDK uses `map[string]any` aliases for these — passing a non-nil
	// empty map is what makes Auth0 enable the addon.
	if isAddonEnabled(m.Box) {
		x := mgmt.ClientAddonBox{}
		out.Box = &x
	}
	if isAddonEnabled(m.Cloudbees) {
		x := mgmt.ClientAddonCloudBees{}
		out.Cloudbees = &x
	}
	if isAddonEnabled(m.Concur) {
		x := mgmt.ClientAddonConcur{}
		out.Concur = &x
	}
	if isAddonEnabled(m.Dropbox) {
		x := mgmt.ClientAddonDropbox{}
		out.Dropbox = &x
	}
	if isAddonEnabled(m.Wsfed) {
		x := mgmt.ClientAddonWsFed{}
		out.Wsfed = &x
	}
	if isAddonEnabled(m.Oag) {
		out.Oag = &mgmt.ClientAddonOag{}
	}

	// ----- Single-string addons ----------------------------------------
	if v := singleStringFromTF(m.Echosign, "domain"); v != nil {
		out.Echosign = &mgmt.ClientAddonEchoSign{Domain: v}
	}
	if v := singleStringFromTF(m.Egnyte, "domain"); v != nil {
		out.Egnyte = &mgmt.ClientAddonEgnyte{Domain: v}
	}
	if v := singleStringFromTF(m.Mscrm, "url"); v != nil {
		out.Mscrm = &mgmt.ClientAddonMscrm{URL: *v}
	}
	if v := singleStringFromTF(m.Newrelic, "account"); v != nil {
		out.Newrelic = &mgmt.ClientAddonNewRelic{Account: v}
	}
	if v := singleStringFromTF(m.Rms, "url"); v != nil {
		out.Rms = &mgmt.ClientAddonRms{URL: *v}
	}
	if v := singleStringFromTF(m.Salesforce, "entity_id"); v != nil {
		out.Salesforce = &mgmt.ClientAddonSalesforce{EntityID: v}
	}
	if v := singleStringFromTF(m.Slack, "team"); v != nil {
		out.Slack = &mgmt.ClientAddonSlack{Team: *v}
	}
	if v := singleStringFromTF(m.Springcm, "acsurl"); v != nil {
		out.Springcm = &mgmt.ClientAddonSpringCm{Acsurl: v}
	}
	if v := singleStringFromTF(m.Wams, "masterkey"); v != nil {
		out.Wams = &mgmt.ClientAddonWams{Masterkey: v}
	}
	if v := singleStringFromTF(m.Zendesk, "account_name"); v != nil {
		out.Zendesk = &mgmt.ClientAddonZendesk{AccountName: v}
	}
	if v := singleStringFromTF(m.Zoom, "account"); v != nil {
		out.Zoom = &mgmt.ClientAddonZoom{Account: v}
	}

	// ----- Firebase / Google -------------------------------------------
	if v := expandFirebaseAddon(m.Firebase); v != nil {
		out.Firebase = v
	}

	// ----- SAML-family --------------------------------------------------
	if v := expandSamlpAddon(ctx, m.Samlp, diags); v != nil {
		out.Samlp = v
	}
	if v := expandOffice365Addon(m.Office365); v != nil {
		out.Office365 = v
	}
	if v := expandSharepointAddon(ctx, m.Sharepoint, diags); v != nil {
		out.Sharepoint = v
	}
	if v := expandSalesforceAPIAddon(m.SalesforceAPI); v != nil {
		out.SalesforceAPI = v
	}
	if v := expandSalesforceAPIAddonSandbox(m.SalesforceSandboxAPI); v != nil {
		out.SalesforceSandboxAPI = v
	}
	if v := expandSapAPIAddon(m.SapAPI); v != nil {
		out.SapAPI = v
	}

	// ----- Misc ---------------------------------------------------------
	if v := expandLayerAddon(m.Layer); v != nil {
		out.Layer = v
	}
	if v := expandSentryAddon(m.Sentry); v != nil {
		out.Sentry = v
	}
	if v := expandSSOIntegrationAddon(m.SSOIntegration); v != nil {
		out.SSOIntegration = v
	}

	return out
}

// =========================================================================
// flatten: *mgmt.ClientAddons -> types.Object
// =========================================================================

func flattenAddons(in *mgmt.ClientAddons, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(addonsAttrTypes())
	}
	values := map[string]attr.Value{
		// ----- AWS / cloud storage -----------------------------------
		"aws":        flattenAWSAddon(in.Aws, diags),
		"azure_blob": flattenAzureBlobAddon(in.AzureBlob, diags),
		"azure_sb":   flattenAzureSbAddon(in.AzureSb, diags),

		// ----- Empty (flag-only) addons ------------------------------
		"box":       flattenEmptyAddon(in.Box != nil, diags),
		"cloudbees": flattenEmptyAddon(in.Cloudbees != nil, diags),
		"concur":    flattenEmptyAddon(in.Concur != nil, diags),
		"dropbox":   flattenEmptyAddon(in.Dropbox != nil, diags),
		"oag":       flattenEmptyAddon(in.Oag != nil, diags),
		"wsfed":     flattenEmptyAddon(in.Wsfed != nil, diags),

		// ----- Single-string addons ----------------------------------
		"echosign":   flattenSingleString("domain", echoSignDomain(in.Echosign), diags),
		"egnyte":     flattenSingleString("domain", egnyteDomain(in.Egnyte), diags),
		"mscrm":      flattenSingleStringRequired("url", mscrmURL(in.Mscrm), diags),
		"newrelic":   flattenSingleString("account", newrelicAccount(in.Newrelic), diags),
		"rms":        flattenSingleStringRequired("url", rmsURL(in.Rms), diags),
		"salesforce": flattenSingleString("entity_id", salesforceEntityID(in.Salesforce), diags),
		"slack":      flattenSingleStringRequired("team", slackTeam(in.Slack), diags),
		"springcm":   flattenSingleString("acsurl", springcmAcsurl(in.Springcm), diags),
		"wams":       flattenSingleString("masterkey", wamsMasterkey(in.Wams), diags),
		"zendesk":    flattenSingleString("account_name", zendeskAccountName(in.Zendesk), diags),
		"zoom":       flattenSingleString("account", zoomAccount(in.Zoom), diags),

		// ----- Firebase / Google -------------------------------------
		"firebase": flattenFirebaseAddon(in.Firebase, diags),

		// ----- SAML-family -------------------------------------------
		"samlp":                  flattenSamlpAddon(in.Samlp, diags),
		"office365":              flattenOffice365Addon(in.Office365, diags),
		"sharepoint":             flattenSharepointAddon(in.Sharepoint, diags),
		"salesforce_api":         flattenSalesforceAPIAddon(in.SalesforceAPI, diags),
		"salesforce_sandbox_api": flattenSalesforceAPIAddonSandbox(in.SalesforceSandboxAPI, diags),
		"sap_api":                flattenSapAPIAddon(in.SapAPI, diags),

		// ----- Misc --------------------------------------------------
		"layer":           flattenLayerAddon(in.Layer, diags),
		"sentry":          flattenSentryAddon(in.Sentry, diags),
		"sso_integration": flattenSSOIntegrationAddon(in.SSOIntegration, diags),
	}
	out, d := types.ObjectValue(addonsAttrTypes(), values)
	diags.Append(d...)
	return out
}

// =========================================================================
// Reusable helpers — keep the per-addon code thin.
// =========================================================================

// emptyAddonAttrTypes is the attr-type map for an addon with no fields. The
// presence of a non-null object means "addon enabled".
func emptyAddonAttrTypes() map[string]attr.Type { return map[string]attr.Type{} }

// emptyAddonSchema is the schema fragment for an addon that takes no fields.
func emptyAddonSchema(name string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Enable the " + name + " add-on. The block has no fields — its presence enables the integration.",
		Attributes:  map[string]schema.Attribute{},
	}
}

// isAddonEnabled returns true when the typed object is non-null. Used for
// flag-only addons.
func isAddonEnabled(o types.Object) bool { return !o.IsNull() && !o.IsUnknown() }

// flattenEmptyAddon returns an empty Object when the addon is enabled, else null.
func flattenEmptyAddon(enabled bool, diags *diag.Diagnostics) types.Object {
	if !enabled {
		return types.ObjectNull(emptyAddonAttrTypes())
	}
	v, d := types.ObjectValue(emptyAddonAttrTypes(), map[string]attr.Value{})
	diags.Append(d...)
	return v
}

// singleStringAttrTypes builds the attr-type map for an addon with exactly
// one string field. Used by ~10 addons.
func singleStringAttrTypes(field string) map[string]attr.Type {
	return map[string]attr.Type{field: types.StringType}
}

// singleStringAddonSchema is the schema fragment for an addon with one string
// field.
func singleStringAddonSchema(name, field, fieldDesc string) schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: name + " add-on configuration.",
		Attributes: map[string]schema.Attribute{
			field: schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: fieldDesc,
			},
		},
	}
}

// singleStringFromTF reads the named string field out of a single-field
// addon object. Returns nil when the parent is null/unknown.
func singleStringFromTF(o types.Object, field string) *string {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	if v, ok := o.Attributes()[field].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		return v.ValueStringPointer()
	}
	return nil
}

// flattenSingleString builds a single-field object with a *string value
// (nil-safe — null parent for nil pointer).
func flattenSingleString(field string, v *string, diags *diag.Diagnostics) types.Object {
	if v == nil {
		return types.ObjectNull(singleStringAttrTypes(field))
	}
	out, d := types.ObjectValue(singleStringAttrTypes(field), map[string]attr.Value{
		field: types.StringPointerValue(v),
	})
	diags.Append(d...)
	return out
}

// flattenSingleStringRequired is the variant for required-string addons
// (Slack/Mscrm/Rms — value, not pointer).
func flattenSingleStringRequired(field string, v *string, diags *diag.Diagnostics) types.Object {
	return flattenSingleString(field, v, diags)
}

// Field accessors so the flatten map above stays compact. Each returns a
// *string (nil when the addon was absent).
func echoSignDomain(a *mgmt.ClientAddonEchoSign) *string {
	if a == nil {
		return nil
	}
	return a.Domain
}
func egnyteDomain(a *mgmt.ClientAddonEgnyte) *string {
	if a == nil {
		return nil
	}
	return a.Domain
}
func mscrmURL(a *mgmt.ClientAddonMscrm) *string {
	if a == nil {
		return nil
	}
	v := a.URL
	return &v
}
func newrelicAccount(a *mgmt.ClientAddonNewRelic) *string {
	if a == nil {
		return nil
	}
	return a.Account
}
func rmsURL(a *mgmt.ClientAddonRms) *string {
	if a == nil {
		return nil
	}
	v := a.URL
	return &v
}
func salesforceEntityID(a *mgmt.ClientAddonSalesforce) *string {
	if a == nil {
		return nil
	}
	return a.EntityID
}
func slackTeam(a *mgmt.ClientAddonSlack) *string {
	if a == nil {
		return nil
	}
	v := a.Team
	return &v
}
func springcmAcsurl(a *mgmt.ClientAddonSpringCm) *string {
	if a == nil {
		return nil
	}
	return a.Acsurl
}
func wamsMasterkey(a *mgmt.ClientAddonWams) *string {
	if a == nil {
		return nil
	}
	return a.Masterkey
}
func zendeskAccountName(a *mgmt.ClientAddonZendesk) *string {
	if a == nil {
		return nil
	}
	return a.AccountName
}
func zoomAccount(a *mgmt.ClientAddonZoom) *string {
	if a == nil {
		return nil
	}
	return a.Account
}

// =========================================================================
// AWS
// =========================================================================

func awsAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"principal":           types.StringType,
		"role":                types.StringType,
		"lifetime_in_seconds": types.Int64Type,
	}
}

func awsAddonSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "AWS STS add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"principal":           schema.StringAttribute{Optional: true, Computed: true, Description: "AWS Principal ARN."},
			"role":                schema.StringAttribute{Optional: true, Computed: true, Description: "AWS Role ARN to assume."},
			"lifetime_in_seconds": schema.Int64Attribute{Optional: true, Computed: true, Description: "AssumeRoleWithSAML credentials lifetime."},
		},
	}
}

func expandAWSAddon(o types.Object) *mgmt.ClientAddonAws {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientAddonAws{}
	if v, ok := a["principal"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.Principal = v.ValueStringPointer()
	}
	if v, ok := a["role"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.Role = v.ValueStringPointer()
	}
	if v, ok := a["lifetime_in_seconds"].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		out.LifetimeInSeconds = framework.Int64ToIntPtr(v)
	}
	return out
}

func flattenAWSAddon(in *mgmt.ClientAddonAws, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(awsAddonAttrTypes())
	}
	v, d := types.ObjectValue(awsAddonAttrTypes(), map[string]attr.Value{
		"principal":           types.StringPointerValue(in.Principal),
		"role":                types.StringPointerValue(in.Role),
		"lifetime_in_seconds": framework.IntPtrToInt64(in.LifetimeInSeconds),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Azure Blob
// =========================================================================

func azureBlobAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"account_name":       types.StringType,
		"storage_access_key": types.StringType,
		"container_name":     types.StringType,
		"blob_name":          types.StringType,
		"expiration":         types.Int64Type,
		"signed_identifier":  types.StringType,
		"blob_read":          types.BoolType,
		"blob_write":         types.BoolType,
		"blob_delete":        types.BoolType,
		"container_read":     types.BoolType,
		"container_write":    types.BoolType,
		"container_delete":   types.BoolType,
		"container_list":     types.BoolType,
	}
}

func azureBlobAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	bl := func(d string) schema.Attribute {
		return schema.BoolAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Azure Blob Storage add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"account_name":       str("Storage account name."),
			"storage_access_key": str("Storage access key (sensitive)."),
			"container_name":     str("Blob container name."),
			"blob_name":          str("Blob name."),
			"expiration":         schema.Int64Attribute{Optional: true, Computed: true, Description: "SAS lifetime in minutes."},
			"signed_identifier":  str("Signed identifier."),
			"blob_read":          bl("Permit blob read access."),
			"blob_write":         bl("Permit blob write access."),
			"blob_delete":        bl("Permit blob delete access."),
			"container_read":     bl("Permit container read access."),
			"container_write":    bl("Permit container write access."),
			"container_delete":   bl("Permit container delete access."),
			"container_list":     bl("Permit container list access."),
		},
	}
}

func expandAzureBlobAddon(o types.Object) *mgmt.ClientAddonAzureBlob {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientAddonAzureBlob{}
	out.AccountName = strPtrFromObj(a, "account_name")
	out.StorageAccessKey = strPtrFromObj(a, "storage_access_key")
	out.ContainerName = strPtrFromObj(a, "container_name")
	out.BlobName = strPtrFromObj(a, "blob_name")
	out.Expiration = intPtrFromObj(a, "expiration")
	out.SignedIdentifier = strPtrFromObj(a, "signed_identifier")
	out.BlobRead = boolPtrFromObj(a, "blob_read")
	out.BlobWrite = boolPtrFromObj(a, "blob_write")
	out.BlobDelete = boolPtrFromObj(a, "blob_delete")
	out.ContainerRead = boolPtrFromObj(a, "container_read")
	out.ContainerWrite = boolPtrFromObj(a, "container_write")
	out.ContainerDelete = boolPtrFromObj(a, "container_delete")
	out.ContainerList = boolPtrFromObj(a, "container_list")
	return out
}

func flattenAzureBlobAddon(in *mgmt.ClientAddonAzureBlob, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(azureBlobAddonAttrTypes())
	}
	v, d := types.ObjectValue(azureBlobAddonAttrTypes(), map[string]attr.Value{
		"account_name":       types.StringPointerValue(in.AccountName),
		"storage_access_key": types.StringPointerValue(in.StorageAccessKey),
		"container_name":     types.StringPointerValue(in.ContainerName),
		"blob_name":          types.StringPointerValue(in.BlobName),
		"expiration":         framework.IntPtrToInt64(in.Expiration),
		"signed_identifier":  types.StringPointerValue(in.SignedIdentifier),
		"blob_read":          types.BoolPointerValue(in.BlobRead),
		"blob_write":         types.BoolPointerValue(in.BlobWrite),
		"blob_delete":        types.BoolPointerValue(in.BlobDelete),
		"container_read":     types.BoolPointerValue(in.ContainerRead),
		"container_write":    types.BoolPointerValue(in.ContainerWrite),
		"container_delete":   types.BoolPointerValue(in.ContainerDelete),
		"container_list":     types.BoolPointerValue(in.ContainerList),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Azure Service Bus
// =========================================================================

func azureSbAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"namespace":    types.StringType,
		"sas_key_name": types.StringType,
		"sas_key":      types.StringType,
		"entity_path":  types.StringType,
		"expiration":   types.Int64Type,
	}
}

func azureSbAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Azure Service Bus add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"namespace":    str("Service Bus namespace."),
			"sas_key_name": str("SAS policy name."),
			"sas_key":      str("SAS key (sensitive)."),
			"entity_path":  str("Queue or topic name."),
			"expiration":   schema.Int64Attribute{Optional: true, Computed: true, Description: "SAS lifetime in minutes."},
		},
	}
}

func expandAzureSbAddon(o types.Object) *mgmt.ClientAddonAzureSb {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonAzureSb{
		Namespace:  strPtrFromObj(a, "namespace"),
		SasKeyName: strPtrFromObj(a, "sas_key_name"),
		SasKey:     strPtrFromObj(a, "sas_key"),
		EntityPath: strPtrFromObj(a, "entity_path"),
		Expiration: intPtrFromObj(a, "expiration"),
	}
}

func flattenAzureSbAddon(in *mgmt.ClientAddonAzureSb, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(azureSbAddonAttrTypes())
	}
	v, d := types.ObjectValue(azureSbAddonAttrTypes(), map[string]attr.Value{
		"namespace":    types.StringPointerValue(in.Namespace),
		"sas_key_name": types.StringPointerValue(in.SasKeyName),
		"sas_key":      types.StringPointerValue(in.SasKey),
		"entity_path":  types.StringPointerValue(in.EntityPath),
		"expiration":   framework.IntPtrToInt64(in.Expiration),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Firebase
// =========================================================================

func firebaseAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"secret":              types.StringType,
		"private_key_id":      types.StringType,
		"private_key":         types.StringType,
		"client_email":        types.StringType,
		"lifetime_in_seconds": types.Int64Type,
	}
}

func firebaseAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Firebase add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"secret":              str("Legacy Firebase secret (v2 deprecated)."),
			"private_key_id":      str("Service-account private-key ID (v3)."),
			"private_key":         str("Service-account private key (v3)."),
			"client_email":        str("Service-account client email (v3)."),
			"lifetime_in_seconds": schema.Int64Attribute{Optional: true, Computed: true, Description: "Token lifetime in seconds."},
		},
	}
}

func expandFirebaseAddon(o types.Object) *mgmt.ClientAddonFirebase {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonFirebase{
		Secret:            strPtrFromObj(a, "secret"),
		PrivateKeyID:      strPtrFromObj(a, "private_key_id"),
		PrivateKey:        strPtrFromObj(a, "private_key"),
		ClientEmail:       strPtrFromObj(a, "client_email"),
		LifetimeInSeconds: intPtrFromObj(a, "lifetime_in_seconds"),
	}
}

func flattenFirebaseAddon(in *mgmt.ClientAddonFirebase, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(firebaseAddonAttrTypes())
	}
	v, d := types.ObjectValue(firebaseAddonAttrTypes(), map[string]attr.Value{
		"secret":              types.StringPointerValue(in.Secret),
		"private_key_id":      types.StringPointerValue(in.PrivateKeyID),
		"private_key":         types.StringPointerValue(in.PrivateKey),
		"client_email":        types.StringPointerValue(in.ClientEmail),
		"lifetime_in_seconds": framework.IntPtrToInt64(in.LifetimeInSeconds),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Office 365
// =========================================================================

func office365AddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"domain":     types.StringType,
		"connection": types.StringType,
	}
}

func office365AddonSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Office 365 add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"domain":     schema.StringAttribute{Optional: true, Computed: true, Description: "Office 365 tenant domain."},
			"connection": schema.StringAttribute{Optional: true, Computed: true, Description: "Auth0 connection name to use."},
		},
	}
}

func expandOffice365Addon(o types.Object) *mgmt.ClientAddonOffice365 {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonOffice365{
		Domain:     strPtrFromObj(a, "domain"),
		Connection: strPtrFromObj(a, "connection"),
	}
}

func flattenOffice365Addon(in *mgmt.ClientAddonOffice365, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(office365AddonAttrTypes())
	}
	v, d := types.ObjectValue(office365AddonAttrTypes(), map[string]attr.Value{
		"domain":     types.StringPointerValue(in.Domain),
		"connection": types.StringPointerValue(in.Connection),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Salesforce API & Salesforce Sandbox API (same shape — different SDK type)
// =========================================================================

func salesforceAPIAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"clientid":              types.StringType,
		"principal":             types.StringType,
		"community_name":        types.StringType,
		"community_url_section": types.StringType,
	}
}

func salesforceAPIAddonSchema(name string) schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: name + " add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"clientid":              str("Salesforce-issued client ID."),
			"principal":             str("Principal name."),
			"community_name":        str("Salesforce community name."),
			"community_url_section": str("Salesforce community URL section."),
		},
	}
}

func expandSalesforceAPIAddon(o types.Object) *mgmt.ClientAddonSalesforceAPI {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonSalesforceAPI{
		Clientid:            strPtrFromObj(a, "clientid"),
		Principal:           strPtrFromObj(a, "principal"),
		CommunityName:       strPtrFromObj(a, "community_name"),
		CommunityURLSection: strPtrFromObj(a, "community_url_section"),
	}
}

func expandSalesforceAPIAddonSandbox(o types.Object) *mgmt.ClientAddonSalesforceSandboxAPI {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonSalesforceSandboxAPI{
		Clientid:            strPtrFromObj(a, "clientid"),
		Principal:           strPtrFromObj(a, "principal"),
		CommunityName:       strPtrFromObj(a, "community_name"),
		CommunityURLSection: strPtrFromObj(a, "community_url_section"),
	}
}

func flattenSalesforceAPIAddon(in *mgmt.ClientAddonSalesforceAPI, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(salesforceAPIAttrTypes())
	}
	v, d := types.ObjectValue(salesforceAPIAttrTypes(), map[string]attr.Value{
		"clientid":              types.StringPointerValue(in.Clientid),
		"principal":             types.StringPointerValue(in.Principal),
		"community_name":        types.StringPointerValue(in.CommunityName),
		"community_url_section": types.StringPointerValue(in.CommunityURLSection),
	})
	diags.Append(d...)
	return v
}

func flattenSalesforceAPIAddonSandbox(in *mgmt.ClientAddonSalesforceSandboxAPI, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(salesforceAPIAttrTypes())
	}
	v, d := types.ObjectValue(salesforceAPIAttrTypes(), map[string]attr.Value{
		"clientid":              types.StringPointerValue(in.Clientid),
		"principal":             types.StringPointerValue(in.Principal),
		"community_name":        types.StringPointerValue(in.CommunityName),
		"community_url_section": types.StringPointerValue(in.CommunityURLSection),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// SAP API
// =========================================================================

func sapAPIAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"clientid":               types.StringType,
		"username_attribute":     types.StringType,
		"token_endpoint_url":     types.StringType,
		"scope":                  types.StringType,
		"service_password":       types.StringType,
		"name_identifier_format": types.StringType,
	}
}

func sapAPIAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "SAP API add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"clientid":               str("OAuth2 SAP client ID."),
			"username_attribute":     str("User-object property mapping to SAP username."),
			"token_endpoint_url":     str("SAP OData OAuth2 token endpoint."),
			"scope":                  str("Requested SAP scope."),
			"service_password":       str("Service-account password (sensitive)."),
			"name_identifier_format": str("SAML NameID format."),
		},
	}
}

func expandSapAPIAddon(o types.Object) *mgmt.ClientAddonSapapi {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonSapapi{
		Clientid:             strPtrFromObj(a, "clientid"),
		UsernameAttribute:    strPtrFromObj(a, "username_attribute"),
		TokenEndpointURL:     strPtrFromObj(a, "token_endpoint_url"),
		Scope:                strPtrFromObj(a, "scope"),
		ServicePassword:      strPtrFromObj(a, "service_password"),
		NameIdentifierFormat: strPtrFromObj(a, "name_identifier_format"),
	}
}

func flattenSapAPIAddon(in *mgmt.ClientAddonSapapi, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(sapAPIAttrTypes())
	}
	v, d := types.ObjectValue(sapAPIAttrTypes(), map[string]attr.Value{
		"clientid":               types.StringPointerValue(in.Clientid),
		"username_attribute":     types.StringPointerValue(in.UsernameAttribute),
		"token_endpoint_url":     types.StringPointerValue(in.TokenEndpointURL),
		"scope":                  types.StringPointerValue(in.Scope),
		"service_password":       types.StringPointerValue(in.ServicePassword),
		"name_identifier_format": types.StringPointerValue(in.NameIdentifierFormat),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// SharePoint
// =========================================================================

func sharepointAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"url":           types.StringType,
		"external_urls": types.ListType{ElemType: types.StringType},
	}
}

func sharepointAddonSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "SharePoint add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"url":           schema.StringAttribute{Optional: true, Computed: true, Description: "SharePoint URL."},
			"external_urls": schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "Additional external URLs to advertise."},
		},
	}
}

func expandSharepointAddon(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientAddonSharePoint {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientAddonSharePoint{
		URL: strPtrFromObj(a, "url"),
	}
	if v, ok := a["external_urls"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		urls := framework.StringListToGo(ctx, v, diags)
		if len(urls) > 0 {
			ext := &mgmt.ClientAddonSharePointExternalURL{}
			if len(urls) == 1 {
				// SDK union: single value flows in as the "string" arm.
				ext.String = urls[0]
			} else {
				ext.StringList = urls
			}
			out.ExternalURL = ext
		}
	}
	return out
}

func flattenSharepointAddon(in *mgmt.ClientAddonSharePoint, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(sharepointAddonAttrTypes())
	}
	urls := types.ListNull(types.StringType)
	if in.ExternalURL != nil {
		var combined []string
		if in.ExternalURL.String != "" {
			combined = append(combined, in.ExternalURL.String)
		}
		combined = append(combined, in.ExternalURL.StringList...)
		urls = framework.StringSliceToList(combined)
	}
	v, d := types.ObjectValue(sharepointAddonAttrTypes(), map[string]attr.Value{
		"url":           types.StringPointerValue(in.URL),
		"external_urls": urls,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Layer
// =========================================================================

func layerAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"provider_id": types.StringType,
		"key_id":      types.StringType,
		"private_key": types.StringType,
		"principal":   types.StringType,
		"expiration":  types.Int64Type,
	}
}

func layerAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Layer add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"provider_id": str("Layer provider ID."),
			"key_id":      str("Layer key ID."),
			"private_key": str("Layer private key (sensitive)."),
			"principal":   str("Principal."),
			"expiration":  schema.Int64Attribute{Optional: true, Computed: true, Description: "Token lifetime in seconds."},
		},
	}
}

func expandLayerAddon(o types.Object) *mgmt.ClientAddonLayer {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientAddonLayer{
		Principal:  strPtrFromObj(a, "principal"),
		Expiration: intPtrFromObj(a, "expiration"),
	}
	if v, ok := a["provider_id"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.ProviderID = v.ValueString()
	}
	if v, ok := a["key_id"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.KeyID = v.ValueString()
	}
	if v, ok := a["private_key"].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		out.PrivateKey = v.ValueString()
	}
	return out
}

func flattenLayerAddon(in *mgmt.ClientAddonLayer, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(layerAddonAttrTypes())
	}
	v, d := types.ObjectValue(layerAddonAttrTypes(), map[string]attr.Value{
		"provider_id": types.StringValue(in.ProviderID),
		"key_id":      types.StringValue(in.KeyID),
		"private_key": types.StringValue(in.PrivateKey),
		"principal":   types.StringPointerValue(in.Principal),
		"expiration":  framework.IntPtrToInt64(in.Expiration),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// Sentry
// =========================================================================

func sentryAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"org_slug": types.StringType,
		"base_url": types.StringType,
	}
}

func sentryAddonSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "Sentry add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"org_slug": schema.StringAttribute{Optional: true, Computed: true, Description: "Sentry organisation slug."},
			"base_url": schema.StringAttribute{Optional: true, Computed: true, Description: "Sentry on-prem base URL."},
		},
	}
}

func expandSentryAddon(o types.Object) *mgmt.ClientAddonSentry {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonSentry{
		OrgSlug: strPtrFromObj(a, "org_slug"),
		BaseURL: strPtrFromObj(a, "base_url"),
	}
}

func flattenSentryAddon(in *mgmt.ClientAddonSentry, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(sentryAddonAttrTypes())
	}
	v, d := types.ObjectValue(sentryAddonAttrTypes(), map[string]attr.Value{
		"org_slug": types.StringPointerValue(in.OrgSlug),
		"base_url": types.StringPointerValue(in.BaseURL),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// SSO Integration
// =========================================================================

func ssoIntegrationAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":    types.StringType,
		"version": types.StringType,
	}
}

func ssoIntegrationAddonSchema() schema.Attribute {
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "SSO Integration add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"name":    schema.StringAttribute{Optional: true, Computed: true, Description: "Integration name."},
			"version": schema.StringAttribute{Optional: true, Computed: true, Description: "Integration version."},
		},
	}
}

func expandSSOIntegrationAddon(o types.Object) *mgmt.ClientAddonSSOIntegration {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	return &mgmt.ClientAddonSSOIntegration{
		Name:    strPtrFromObj(a, "name"),
		Version: strPtrFromObj(a, "version"),
	}
}

func flattenSSOIntegrationAddon(in *mgmt.ClientAddonSSOIntegration, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(ssoIntegrationAttrTypes())
	}
	v, d := types.ObjectValue(ssoIntegrationAttrTypes(), map[string]attr.Value{
		"name":    types.StringPointerValue(in.Name),
		"version": types.StringPointerValue(in.Version),
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// SAML (samlp) — by far the largest addon
// =========================================================================

func samlpAddonAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"audience":                           types.StringType,
		"recipient":                          types.StringType,
		"create_upn_claim":                   types.BoolType,
		"map_unknown_claims_as_is":           types.BoolType,
		"passthrough_claims_with_no_mapping": types.BoolType,
		"map_identities":                     types.BoolType,
		"signature_algorithm":                types.StringType,
		"digest_algorithm":                   types.StringType,
		"issuer":                             types.StringType,
		"destination":                        types.StringType,
		"lifetime_in_seconds":                types.Int64Type,
		"sign_response":                      types.BoolType,
		"name_identifier_format":             types.StringType,
		"name_identifier_probes":             types.ListType{ElemType: types.StringType},
		"authn_context_class_ref":            types.StringType,
		// `mappings` is `map[string]any` in the SDK; a JSON-encoded string
		// keeps full fidelity without forcing us to commit to one type.
		"mappings": types.StringType,
	}
}

func samlpAddonSchema() schema.Attribute {
	str := func(d string) schema.Attribute {
		return schema.StringAttribute{Optional: true, Computed: true, Description: d}
	}
	bl := func(d string) schema.Attribute {
		return schema.BoolAttribute{Optional: true, Computed: true, Description: d}
	}
	return schema.SingleNestedAttribute{
		Optional: true, Computed: true,
		Description: "SAML protocol (`samlp`) add-on configuration.",
		Attributes: map[string]schema.Attribute{
			"audience":                           str("SAML audience."),
			"recipient":                          str("SAML recipient (ACS URL)."),
			"create_upn_claim":                   bl("Create the userPrincipalName claim."),
			"map_unknown_claims_as_is":           bl("Map unknown claims as-is."),
			"passthrough_claims_with_no_mapping": bl("Pass through claims with no explicit mapping."),
			"map_identities":                     bl("Include identity-provider data in the assertion."),
			"signature_algorithm":                str("`rsa-sha1` or `rsa-sha256`."),
			"digest_algorithm":                   str("`sha1` or `sha256`."),
			"issuer":                             str("Override the SAML issuer."),
			"destination":                        str("SAML response Destination attribute."),
			"lifetime_in_seconds":                schema.Int64Attribute{Optional: true, Computed: true, Description: "Assertion lifetime in seconds."},
			"sign_response":                      bl("Sign the SAML response."),
			"name_identifier_format":             str("NameID format URI."),
			"name_identifier_probes":             schema.ListAttribute{Optional: true, Computed: true, ElementType: types.StringType, Description: "User-attribute names probed in order to populate NameID."},
			"authn_context_class_ref":            str("AuthnContextClassRef value."),
			"mappings":                           str("Free-form claim mappings as a JSON object string."),
		},
	}
}

func expandSamlpAddon(ctx context.Context, o types.Object, diags *diag.Diagnostics) *mgmt.ClientAddonSAML {
	if o.IsNull() || o.IsUnknown() {
		return nil
	}
	a := o.Attributes()
	out := &mgmt.ClientAddonSAML{
		Audience:                       strPtrFromObj(a, "audience"),
		Recipient:                      strPtrFromObj(a, "recipient"),
		CreateUpnClaim:                 boolPtrFromObj(a, "create_upn_claim"),
		MapUnknownClaimsAsIs:           boolPtrFromObj(a, "map_unknown_claims_as_is"),
		PassthroughClaimsWithNoMapping: boolPtrFromObj(a, "passthrough_claims_with_no_mapping"),
		MapIdentities:                  boolPtrFromObj(a, "map_identities"),
		SignatureAlgorithm:             strPtrFromObj(a, "signature_algorithm"),
		DigestAlgorithm:                strPtrFromObj(a, "digest_algorithm"),
		Issuer:                         strPtrFromObj(a, "issuer"),
		Destination:                    strPtrFromObj(a, "destination"),
		LifetimeInSeconds:              intPtrFromObj(a, "lifetime_in_seconds"),
		SignResponse:                   boolPtrFromObj(a, "sign_response"),
		NameIdentifierFormat:           strPtrFromObj(a, "name_identifier_format"),
		AuthnContextClassRef:           strPtrFromObj(a, "authn_context_class_ref"),
	}
	if v, ok := a["name_identifier_probes"].(types.List); ok && !v.IsNull() && !v.IsUnknown() {
		out.NameIdentifierProbes = framework.StringListToGo(ctx, v, diags)
	}
	if v, ok := a["mappings"].(types.String); ok {
		if parsed, present := framework.ParseJSONString(v, "addons.samlp.mappings", diags); present {
			m := mgmt.ClientAddonSAMLMapping{}
			if mp, ok := parsed.(map[string]any); ok {
				m = mp
			}
			out.Mappings = &m
		}
	}
	return out
}

func flattenSamlpAddon(in *mgmt.ClientAddonSAML, diags *diag.Diagnostics) types.Object {
	if in == nil {
		return types.ObjectNull(samlpAddonAttrTypes())
	}
	mappingsStr := types.StringNull()
	if in.Mappings != nil {
		mappingsStr = framework.FlattenJSONToString(map[string]any(*in.Mappings), diags)
	}
	v, d := types.ObjectValue(samlpAddonAttrTypes(), map[string]attr.Value{
		"audience":                           types.StringPointerValue(in.Audience),
		"recipient":                          types.StringPointerValue(in.Recipient),
		"create_upn_claim":                   types.BoolPointerValue(in.CreateUpnClaim),
		"map_unknown_claims_as_is":           types.BoolPointerValue(in.MapUnknownClaimsAsIs),
		"passthrough_claims_with_no_mapping": types.BoolPointerValue(in.PassthroughClaimsWithNoMapping),
		"map_identities":                     types.BoolPointerValue(in.MapIdentities),
		"signature_algorithm":                types.StringPointerValue(in.SignatureAlgorithm),
		"digest_algorithm":                   types.StringPointerValue(in.DigestAlgorithm),
		"issuer":                             types.StringPointerValue(in.Issuer),
		"destination":                        types.StringPointerValue(in.Destination),
		"lifetime_in_seconds":                framework.IntPtrToInt64(in.LifetimeInSeconds),
		"sign_response":                      types.BoolPointerValue(in.SignResponse),
		"name_identifier_format":             types.StringPointerValue(in.NameIdentifierFormat),
		"name_identifier_probes":             framework.StringSliceToList(in.NameIdentifierProbes),
		"authn_context_class_ref":            types.StringPointerValue(in.AuthnContextClassRef),
		"mappings":                           mappingsStr,
	})
	diags.Append(d...)
	return v
}

// =========================================================================
// tiny obj-attribute readers (used by every addon expand)
// =========================================================================

func strPtrFromObj(a map[string]attr.Value, name string) *string {
	if v, ok := a[name].(types.String); ok && !v.IsNull() && !v.IsUnknown() {
		return v.ValueStringPointer()
	}
	return nil
}

func boolPtrFromObj(a map[string]attr.Value, name string) *bool {
	if v, ok := a[name].(types.Bool); ok && !v.IsNull() && !v.IsUnknown() {
		return v.ValueBoolPointer()
	}
	return nil
}

func intPtrFromObj(a map[string]attr.Value, name string) *int {
	if v, ok := a[name].(types.Int64); ok && !v.IsNull() && !v.IsUnknown() {
		return framework.Int64ToIntPtr(v)
	}
	return nil
}
