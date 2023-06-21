package connection

import (
	"context"
	"errors"
	"net/http"

	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalSchema "github.com/auth0/terraform-provider-auth0/internal/schema"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// TypeSpecificExpandConnectionFunction is a generic function signature for connection expansion.
type TypeSpecificExpandConnectionFunction[T interface{}] func(
	conn *management.Connection,
	data *schema.ResourceData,
	api *management.Management,
) (*management.Connection, diag.Diagnostics)

// TypeSpecificFlattenConnectionFunction is a generic function signature for connection flatten.
type TypeSpecificFlattenConnectionFunction[T interface{}] func(
	data *schema.ResourceData,
	options *T,
) (map[string]interface{}, diag.Diagnostics)

// NewBaseConnectionResource will return a new auth0_connection resource.
func NewBaseConnectionResource[T interface{}](description string, optionsSchema map[string]*schema.Schema, typeSpecificExpand TypeSpecificExpandConnectionFunction[T], typeSpecificFlatten TypeSpecificFlattenConnectionFunction[T]) *schema.Resource {
	resourceSchema := internalSchema.Clone(baseConnectionSchema)

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
		Description:   description,
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

		result := multierror.Append(
			d.Set("name", connection.GetName()),
			d.Set("display_name", connection.GetDisplayName()),
			d.Set("is_domain_connection", connection.GetIsDomainConnection()),
			d.Set("strategy", connection.GetStrategy()),
			d.Set("realms", connection.GetRealms()),
			d.Set("metadata", connection.GetMetadata()),
			d.Set("enabled_clients", connection.GetEnabledClients()),
		)

		connectionOptions, diags := flattenConnectionOptions(d, opts, typeSpecificFlatten)
		if diags.HasError() {
			return diags
		}

		for _, opts := range connectionOptions {
			for key, v := range opts {
				if key == "configuration" {
					continue // Preventing update of sensitive field.
				}
				result = multierror.Append(d.Set(key, v), err)
			}
			break
		}

		switch connection.GetStrategy() {
		//TODO: eventually remove this and integrate with connection-specific flatten functions.
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
	}

	if d.HasChange("realms") {
		connection.Realms = value.Strings(config.GetAttr("realms"))
	}

	if d.HasChange("enabled_clients") {
		connection.EnabledClients = value.Strings(config.GetAttr("enabled_clients"))
	}

	return expandSpecificConnectionType(connection, d, api)
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
			return nil
		}
		return diag.FromErr(err)
	}

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
