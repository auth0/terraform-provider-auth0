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
		DomainMetadata:       expandCustomDomainMetadata(data),
	}

	if data.IsNewResource() {
		customDomain.Domain = value.String(cfg.GetAttr("domain"))
		customDomain.Type = value.String(cfg.GetAttr("type"))
	}

	return customDomain
}

func expandCustomDomainMetadata(data *schema.ResourceData) *map[string]interface{} {
	if !data.HasChange("domain_metadata") {
		return nil
	}

	oldMetadata, newMetadata := data.GetChange("domain_metadata")
	oldMetadataMap := oldMetadata.(map[string]interface{})
	newMetadataMap := newMetadata.(map[string]interface{})

	for key := range oldMetadataMap {
		if _, ok := newMetadataMap[key]; !ok {
			newMetadataMap[key] = nil
		}
	}

	return &newMetadataMap
}
