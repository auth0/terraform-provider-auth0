package signingkey

import (
	"github.com/auth0/go-auth0/management"
)

func flattenSigningKeys(keys []*management.SigningKey) []interface{} {
	var result []interface{}
	for _, key := range keys {
		result = append(result, map[string]interface{}{
			"kid":         key.GetKID(),
			"cert":        key.GetCert(),
			"pkcs7":       key.GetPKCS7(),
			"current":     key.GetCurrent(),
			"next":        key.GetNext(),
			"previous":    key.GetPrevious(),
			"revoked":     key.GetRevoked(),
			"fingerprint": key.GetFingerprint(),
			"thumbprint":  key.GetThumbprint(),
		})
	}
	return result
}
