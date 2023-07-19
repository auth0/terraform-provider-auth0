package sweep

import (
	"context"
	"fmt"
	"os"

	"github.com/auth0/go-auth0/management"
)

// auth0API returns an instance of the Management
// API Client used within test sweepers.
func auth0API() (*management.Management, error) {
	ctx := context.Background()

	domain := os.Getenv("AUTH0_DOMAIN")
	if domain == "" {
		return nil, fmt.Errorf("failed to instantiate api client: AUTH0_DOMAIN is empty")
	}

	apiToken := os.Getenv("AUTH0_API_TOKEN")
	authenticationOption := management.WithStaticToken(apiToken)
	if apiToken == "" {
		authenticationOption = management.WithClientCredentials(
			ctx,
			os.Getenv("AUTH0_CLIENT_ID"),
			os.Getenv("AUTH0_CLIENT_SECRET"),
		)
	}

	return management.New(domain, authenticationOption)
}
