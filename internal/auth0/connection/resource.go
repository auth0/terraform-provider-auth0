package connection

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/config"
	internalError "github.com/auth0/terraform-provider-auth0/internal/error"
)

// NewResource will return a new auth0_connection resource.
func NewResource() *schema.Resource {
	return &schema.Resource{
		CreateContext: createConnection,
		ReadContext:   readConnection,
		UpdateContext: updateConnection,
		DeleteContext: deleteConnection,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: validateConnection,
		Description: "With Auth0, you can define sources of users, otherwise known as connections, " +
			"which may include identity providers (such as Google or LinkedIn), databases, or " +
			"passwordless authentication methods. This resource allows you to configure " +
			"and manage connections to be used with your clients and users.",
		Schema:        resourceSchema,
		SchemaVersion: 3,
	}
}

// validateConnection rejects a planned change to options.attributes.email.unique.
// The Management API only accepts that property when the connection is created and
// rejects any PATCH that changes it, so such a diff can never be applied and would
// otherwise persist forever. Failing here surfaces it at plan time instead.
func validateConnection(_ context.Context, diff *schema.ResourceDiff, _ interface{}) error {
	// On create there is nothing to conflict with; the API accepts any value.
	if diff.Id() == "" {
		return nil
	}

	// Without a stored email attribute there is no stored unique to conflict with,
	// so the connection is free to adopt whatever the configuration asks for. This
	// has to be checked before reading the value itself: GetChange cannot express
	// an absent value and reports the zero value, making "no email attribute" look
	// identical to "stored as false".
	storedEmailCount, _ := diff.GetChange("options.0.attributes.0.email.#")
	if count, ok := storedEmailCount.(int); !ok || count == 0 {
		return nil
	}

	// A value the configuration leaves unset is filled in from state by the
	// Computed flag, so only an explicit change shows up here.
	oldValue, newValue := diff.GetChange("options.0.attributes.0.email.0.unique")
	if oldValue == newValue {
		return nil
	}

	return fmt.Errorf(
		"options.attributes.email.unique cannot be changed after the connection is created: "+
			"it is %v and the configuration requests %v. The Auth0 Management API only accepts "+
			"this property when the connection is created. Restore it to %v, or destroy and "+
			"recreate the connection to change it",
		oldValue, newValue, oldValue,
	)
}

func createConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, data, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	if err := api.Connection.Create(ctx, connection); err != nil {
		return diag.FromErr(err)
	}

	data.SetId(connection.GetID())

	return readConnection(ctx, data, meta)
}

func readConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, err := api.Connection.Read(ctx, data.Id())
	if err != nil {
		return internalError.HandleReadAPIError("auth0_connection", data, err)
	}

	return flattenConnection(data, connection)
}

func updateConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	connection, diagnostics := expandConnection(ctx, data, api)
	if diagnostics.HasError() {
		return diagnostics
	}

	// The typed Update drops the nil field (omitempty), so null it out via a raw PATCH.
	if isConnectionCrossAppAccessResourceAppNull(data) {
		nullFields := map[string]interface{}{"cross_app_access_resource_app": nil}
		if err := api.Request(ctx, http.MethodPatch, api.URI("connections", data.Id()), nullFields); err != nil {
			return diag.FromErr(err)
		}
	}

	if err := api.Connection.Update(ctx, data.Id(), connection); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return readConnection(ctx, data, meta)
}

func deleteConnection(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*config.Config).GetAPI()

	if err := api.Connection.Delete(ctx, data.Id()); err != nil {
		return diag.FromErr(internalError.HandleAPIError(data, err))
	}

	return nil
}
