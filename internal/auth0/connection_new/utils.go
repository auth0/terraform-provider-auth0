package connection_new //nolint:all

import (
	"context"
	"errors"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// NewBaseConnectionResource will return a new auth0_connection resource.
func NewBaseConnectionResource[T interface{}](optionsSchema map[string]*schema.Schema, typeSpecificExpand TypeSpecificExpandConnectionFunction[T], typeSpecificFlatten TypeSpecificFlattenConnectionFunction[T]) *schema.Resource {
	resourceSchema := baseSchema

	for key, value := range optionsSchema {
		resourceSchema[key] = value
	}

	return &schema.Resource{
		CreateContext: createConnection(typeSpecificExpand, typeSpecificFlatten),
		ReadContext:   readConnection(typeSpecificFlatten),
		UpdateContext: updateConnection(typeSpecificExpand, typeSpecificFlatten),
		DeleteContext: deleteConnection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Description: "With Auth0, you can define sources of users, otherwise known as connections, " +
			"which may include identity providers (such as Google or LinkedIn), databases, or " +
			"passwordless authentication methods. This resource allows you to configure " +
			"and manage connections to be used with your clients and users.",
		Schema:        resourceSchema,
		SchemaVersion: 0,
	}
}

func createConnection[T interface{}](typeSpecificExpand TypeSpecificExpandConnectionFunction[T], typeSpecificFlatten TypeSpecificFlattenConnectionFunction[T]) func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		api := m.(*config.Config).GetAPI()

		connection, diagnostics := expandBaseConnection(d, api, typeSpecificExpand)
		if diagnostics.HasError() {
			return diagnostics
		}

		if err := api.Connection.Create(connection); err != nil {
			diagnostics = append(diagnostics, diag.FromErr(err)...)
			return diagnostics
		}

		d.SetId(connection.GetID())

		diagnostics = append(diagnostics, readConnection(typeSpecificFlatten)(ctx, d, m)...)
		return diagnostics
	}
}

func readConnection[T interface{}](typeSpecificFlatten TypeSpecificFlattenConnectionFunction[T]) func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		api := m.(*config.Config).GetAPI()

		connection, err := api.Connection.Read(d.Id())
		if err != nil {
			if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}

		opts, ok := connection.Options.(*T)
		if !ok {
			return diag.FromErr(errors.New("could not convert connection options to specified type"))
		}

		connectionOptions, diags := flattenConnectionOptions(d, opts, typeSpecificFlatten)
		if diags.HasError() {
			return diags
		}

		result := multierror.Append(
			d.Set("name", connection.GetName()),
			d.Set("display_name", connection.GetDisplayName()),
			d.Set("is_domain_connection", connection.GetIsDomainConnection()),
			d.Set("strategy", connection.GetStrategy()),
			d.Set("realms", connection.GetRealms()),
			d.Set("metadata", connection.GetMetadata()),
			d.Set("enabled_clients", connection.GetEnabledClients()),
		)

		for _, value := range connectionOptions {
			for key, v := range value {
				if key == "configuration" {
					continue // For ignoring sensitive fields.
				}
				result = multierror.Append(d.Set(key, v), err)
			}
			break
		}

		switch connection.GetStrategy() {
		case management.ConnectionStrategyGoogleApps,
			management.ConnectionStrategyOIDC,
			management.ConnectionStrategyAD,
			management.ConnectionStrategyAzureAD,
			management.ConnectionStrategySAML,
			management.ConnectionStrategyADFS:
			result = multierror.Append(result, d.Set("show_as_button", connection.GetShowAsButton()))
		}

		return diag.FromErr(result.ErrorOrNil())
	}
}

func expandBaseConnection[T interface{}](d *schema.ResourceData, api *management.Management, expandSpecificConnectionType TypeSpecificExpandConnectionFunction[T]) (*management.Connection, diag.Diagnostics) {
	config := d.GetRawConfig()

	connection := &management.Connection{
		DisplayName:        value.String(config.GetAttr("display_name")),
		IsDomainConnection: value.Bool(config.GetAttr("is_domain_connection")),
		Metadata:           value.MapOfStrings(config.GetAttr("metadata")),
	}

	if d.IsNewResource() {
		connection.Name = value.String(config.GetAttr("name"))
		connection.Strategy = value.String(config.GetAttr("strategy"))
	}

	if d.HasChange("realms") {
		connection.Realms = value.Strings(config.GetAttr("realms"))
	}

	if d.HasChange("enabled_clients") {
		connection.EnabledClients = value.Strings(config.GetAttr("enabled_clients"))
	}

	var diagnostics diag.Diagnostics

	connection.Options, diagnostics = expandSpecificConnectionType(d, d.GetRawConfig(), api)

	return connection, diagnostics
}

func updateConnection[T interface{}](typeSpecificExpand TypeSpecificExpandConnectionFunction[T], typeSpecificFlatten TypeSpecificFlattenConnectionFunction[T]) func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return func(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
		api := m.(*config.Config).GetAPI()

		connection, diagnostics := expandBaseConnection(d, api, typeSpecificExpand)
		if diagnostics.HasError() {
			return diagnostics
		}

		if err := api.Connection.Update(d.Id(), connection); err != nil {
			diagnostics = append(diagnostics, diag.FromErr(err)...)
			return diagnostics
		}

		diagnostics = append(diagnostics, readConnection(typeSpecificFlatten)(ctx, d, m)...)
		return diagnostics
	}
}

func deleteConnection(_ context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	api := m.(*config.Config).GetAPI()

	if err := api.Connection.Delete(d.Id()); err != nil {
		if mErr, ok := err.(management.Error); ok && mErr.Status() == http.StatusNotFound {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

func flattenConnectionOptions[T interface{}](d *schema.ResourceData, connectionOptions *T, flatten TypeSpecificFlattenConnectionFunction[T]) ([]map[string]interface{}, diag.Diagnostics) {
	if connectionOptions == nil {
		return nil, nil
	}

	var m map[string]interface{}
	var diags diag.Diagnostics

	m, diags = flatten(d, connectionOptions)

	return []map[string]interface{}{m}, diags
}
