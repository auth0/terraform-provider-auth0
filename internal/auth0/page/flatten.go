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

// flattenChangePasswordPage flattens the change password page data.
func flattenChangePasswordPage(changePassword *management.TenantChangePassword) []interface{} {
	if changePassword == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["enabled"] = changePassword.Enabled
	m["html"] = changePassword.HTML

	return []interface{}{m}
}

// flattenGuardianMFAPage flattens the guardian mfa page data.
func flattenGuardianMFAPage(mfa *management.TenantGuardianMFAPage) []interface{} {
	if mfa == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["enabled"] = mfa.Enabled
	m["html"] = mfa.HTML

	return []interface{}{m}
}

// flattenErrorPage flattens the error page data.
func flattenErrorPage(errorPage *management.TenantErrorPage) []interface{} {
	if errorPage == nil {
		return nil
	}

	m := make(map[string]interface{})
	m["html"] = errorPage.HTML
	m["show_log_link"] = errorPage.ShowLogLink
	m["url"] = errorPage.URL

	return []interface{}{m}
}
