package page

import (
	"github.com/auth0/go-auth0/management"
)

func flattenLoginPage(clientWithLoginPage *management.Client) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"enabled": clientWithLoginPage.GetCustomLoginPageOn(),
			"html":    clientWithLoginPage.GetCustomLoginPage(),
		},
	}
}
