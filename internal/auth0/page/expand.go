package page

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/auth0/tenant"
	"github.com/auth0/terraform-provider-auth0/internal/value"
)

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
		ChangePassword:  tenant.ExpandTenantChangePassword(cfg.GetAttr("change_password")),
		GuardianMFAPage: tenant.ExpandTenantGuardianMFAPage(cfg.GetAttr("guardian_mfa")),
		ErrorPage:       tenant.ExpandTenantErrorPage(cfg.GetAttr("error")),
	}

	if tenantPages.String() == "{}" {
		return nil
	}

	return tenantPages
}
