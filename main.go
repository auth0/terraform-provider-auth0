package main

import (
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	frameworkProvider "github.com/auth0/terraform-provider-auth0/internal/framework/provider"
	provider "github.com/auth0/terraform-provider-auth0/internal/provider"
)

// Ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	// Set descriptions to support Markdown syntax for SDK resources,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown

	err := tf6server.Serve(
		"registry.terraform.io/auth0/auth0",
		func() tfprotov6.ProviderServer {
			providerServer, err := frameworkProvider.MuxServer(provider.New(), frameworkProvider.New())
			if err != nil {
				log.Fatal(err)
			}
			return providerServer
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}
