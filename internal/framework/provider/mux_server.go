package provider

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MuxServer returns a muxed server which combines an SDK provider and a Framework provider.
func MuxServer(sdkProvider *schema.Provider, frameworkProvider provider.Provider) (tfprotov6.ProviderServer, error) {
	muxServer, err := tf6muxserver.NewMuxServer(
		context.Background(),
		// Add SDK provider.
		func() tfprotov6.ProviderServer {
			// Upgrade SDK provider to protocol 6.
			upgradedSdkProvider, err := tf5to6server.UpgradeServer(
				context.Background(),
				sdkProvider.GRPCProvider,
			)
			if err != nil {
				log.Fatal(err)
			}
			return upgradedSdkProvider
		},
		// Add Framework provider.
		providerserver.NewProtocol6(frameworkProvider),
	)
	if err != nil {
		return nil, err
	}
	return muxServer.ProviderServer(), nil
}
