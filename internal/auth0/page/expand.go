package page

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

// expandChangePasswordPage expands the change password page config.
func expandChangePasswordPage(config cty.Value) *management.TenantChangePassword {
	var changePassword management.TenantChangePassword

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		changePassword.Enabled = value.Bool(d.GetAttr("enabled"))
		changePassword.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	if changePassword == (management.TenantChangePassword{}) {
		return nil
	}

	return &changePassword
}

// expandGuardianMFAPage expands the guardian mfa page config.
func expandGuardianMFAPage(config cty.Value) *management.TenantGuardianMFAPage {
	var mfa management.TenantGuardianMFAPage

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		mfa.Enabled = value.Bool(d.GetAttr("enabled"))
		mfa.HTML = value.String(d.GetAttr("html"))
		return stop
	})

	if mfa == (management.TenantGuardianMFAPage{}) {
		return nil
	}

	return &mfa
}

// expandErrorPage expands the error page config.
func expandErrorPage(config cty.Value) *management.TenantErrorPage {
	var errorPage management.TenantErrorPage

	config.ForEachElement(func(_ cty.Value, d cty.Value) (stop bool) {
		errorPage.HTML = value.String(d.GetAttr("html"))
		errorPage.ShowLogLink = value.Bool(d.GetAttr("show_log_link"))
		errorPage.URL = value.String(d.GetAttr("url"))
		return stop
	})

	if errorPage == (management.TenantErrorPage{}) {
		return nil
	}

	return &errorPage
}

func expandLoginPage(data *schema.ResourceData) *management.Client {
	if !data.HasChange("login") {
		return nil
	}

	var clientWithLoginPage *management.Client

	data.GetRawConfig().GetAttr("login").ForEachElement(func(_ cty.Value, cfg cty.Value) (stop bool) {
		clientWithLoginPage = &management.Client{
			CustomLoginPageOn: value.Bool(cfg.GetAttr("enabled")),
			CustomLoginPage:   value.String(cfg.GetAttr("html")),
		}

		return stop
	})

	return clientWithLoginPage
}

func expandTenantPages(cfg cty.Value) *management.Tenant {
	tenantPages := &management.Tenant{
		ChangePassword:  expandChangePasswordPage(cfg.GetAttr("change_password")),
		GuardianMFAPage: expandGuardianMFAPage(cfg.GetAttr("guardian_mfa")),
		ErrorPage:       expandErrorPage(cfg.GetAttr("error")),
	}

	if tenantPages.String() == "{}" {
		return nil
	}

	return tenantPages
}
