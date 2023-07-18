package customdomain

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/auth0/terraform-provider-auth0/internal/value"
)

func expandCustomDomain(data *schema.ResourceData) *management.CustomDomain {
	cfg := data.GetRawConfig()

	customDomain := &management.CustomDomain{
		TLSPolicy:            value.String(cfg.GetAttr("tls_policy")),
		CustomClientIPHeader: value.String(cfg.GetAttr("custom_client_ip_header")),
	}

	if data.IsNewResource() {
		customDomain.Domain = value.String(cfg.GetAttr("domain"))
		customDomain.Type = value.String(cfg.GetAttr("type"))
	}

	return customDomain
}
