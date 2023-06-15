package connection_new //nolint:all

import (
	"github.com/auth0/go-auth0/management" // TypeSpecificExpandConnectionFunction is a generic function signature for connection expand functions.
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TypeSpecificExpandConnectionFunction is a generic function signature for connection expansion.
type TypeSpecificExpandConnectionFunction[T interface{}] func(
	d *schema.ResourceData,
	config cty.Value,
	api *management.Management,
) (*T, diag.Diagnostics)

// TypeSpecificFlattenConnectionFunction is a generic function signature for connection flatten.
type TypeSpecificFlattenConnectionFunction[T interface{}] func(
	d *schema.ResourceData,
	options *T,
) (map[string]interface{}, diag.Diagnostics)
