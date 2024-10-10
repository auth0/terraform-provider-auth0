package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/provider"
)

// Ensure the documentation is formatted properly.
//go:generate terraform fmt -recursive ./examples/

// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {
	// Set descriptions to support Markdown syntax for SDK resources,
	// this will be used in document generation.
	schema.DescriptionKind = schema.StringMarkdown

	muxServer, err := tf6muxserver.NewMuxServer(
		context.Background(),
		// Add SDK provider.
		func() tfprotov6.ProviderServer {
			// Upgrade SDK provider to protocol 6.
			upgradedSdkProvider, err := tf5to6server.UpgradeServer(
				context.Background(),
				provider.New().GRPCProvider,
			)
			if err != nil {
				log.Fatal(err)
			}
			return upgradedSdkProvider
		},
		// Add Framework provider.
		providerserver.NewProtocol6(provider.NewAuth0Provider()),
	)
	if err != nil {
		log.Fatal(err)
	}

	err = tf6server.Serve(
		"registry.terraform.io/auth0/auth0",
		muxServer.ProviderServer,
	)
	if err != nil {
		log.Fatal(err)
	}
}
