package main

import (
	"os"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
)

// Ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	// Set descriptions to support Markdown syntax,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown

	debug := false

	if debugEnv, err := strconv.ParseBool(os.Getenv("TF_PROVIDER_AUTH0_DEBUG")); err == nil {
		debug = debugEnv
	}

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: provider.New,
		Debug:        debug,
	})
}
