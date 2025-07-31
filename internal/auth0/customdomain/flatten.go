package customdomain

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenCustomDomain(data *schema.ResourceData, customDomain *management.CustomDomain) error {
	result := multierror.Append(
		data.Set("domain", customDomain.GetDomain()),
		data.Set("type", customDomain.GetType()),
		data.Set("primary", customDomain.GetPrimary()),
		data.Set("status", customDomain.GetStatus()),
		data.Set("origin_domain_name", customDomain.GetOriginDomainName()),
		data.Set("custom_client_ip_header", customDomain.GetCustomClientIPHeader()),
		data.Set("tls_policy", customDomain.GetTLSPolicy()),
		data.Set("verification", flattenCustomDomainVerificationMethods(customDomain.GetVerification())),
		data.Set("domain_metadata", customDomain.GetDomainMetadata()),
		data.Set("certificate", flattenCustomDomainCertificates(customDomain.GetCertificate())),
	)
	return result.ErrorOrNil()
}

func flattenCustomDomainCertificates(certificate *management.CustomDomainCertificate) []map[string]interface{} {
	if certificate == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"status":                certificate.GetStatus(),
			"error_msg":             certificate.GetErrorMsg(),
			"certificate_authority": certificate.GetCertificateAuthority(),
			"renews_before":         certificate.GetRenewsBefore(),
		},
	}
}

func flattenCustomDomainVerificationMethods(verification *management.CustomDomainVerification) []map[string]interface{} {
	if verification == nil {
		return nil
	}

	return []map[string]interface{}{
		{
			"methods":          verification.Methods,
			"status":           verification.GetStatus(),
			"error_msg":        verification.GetErrorMsg(),
			"last_verified_at": verification.GetLastVerifiedAt(),
		},
	}
}

func flattenCustomDomainVerification(data *schema.ResourceData, customDomain *management.CustomDomain) error {
	result := multierror.Append(
		data.Set("custom_domain_id", customDomain.GetID()),
		data.Set("origin_domain_name", customDomain.GetOriginDomainName()),
	)

	return result.ErrorOrNil()
}
