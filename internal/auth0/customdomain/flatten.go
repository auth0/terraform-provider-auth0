package customdomain

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenCustomDomain(data *schema.ResourceData, customDomain *management.CustomDomain) error {
	if customDomain == nil {
		return nil
	}

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
		data.Set("relying_party_identifier", customDomain.GetRelyingPartyIdentifier()),
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

func flattenCustomDomainList(data *schema.ResourceData, customDomains []*management.CustomDomain) error {
	if customDomains == nil {
		return data.Set("custom_domains", make([]map[string]interface{}, 0))
	}

	list := make([]map[string]interface{}, 0, len(customDomains))
	for _, domain := range customDomains {
		if domain == nil {
			continue
		}

		entry := map[string]interface{}{
			"domain":                   domain.GetDomain(),
			"type":                     domain.GetType(),
			"primary":                  domain.GetPrimary(),
			"status":                   domain.GetStatus(),
			"origin_domain_name":       domain.GetOriginDomainName(),
			"custom_client_ip_header":  domain.GetCustomClientIPHeader(),
			"tls_policy":               domain.GetTLSPolicy(),
			"verification":             flattenCustomDomainVerificationMethods(domain.GetVerification()),
			"domain_metadata":          domain.GetDomainMetadata(),
			"certificate":              flattenCustomDomainCertificates(domain.GetCertificate()),
			"relying_party_identifier": domain.GetRelyingPartyIdentifier(),
		}

		list = append(list, entry)
	}

	return data.Set("custom_domains", list)
}
