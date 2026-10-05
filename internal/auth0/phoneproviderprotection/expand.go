package phoneproviderprotection

import (
	"github.com/auth0/go-auth0/v3/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func expandPhoneProviderProtection(d *schema.ResourceData) *management.PatchPhoneProviderProtectionRequestContent {
	cfg := &management.PatchPhoneProviderProtectionRequestContent{}
	cfg.SetType(management.PhoneProviderProtectionBackoffStrategyEnum(d.Get("type").(string)))
	return cfg
}

func expandPhoneProviderProtectionForDelete() *management.PatchPhoneProviderProtectionRequestContent {
	cfg := &management.PatchPhoneProviderProtectionRequestContent{}
	cfg.SetType(management.PhoneProviderProtectionBackoffStrategyEnumDefault)
	return cfg
}
