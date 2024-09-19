package encryptionkey

import (
	"github.com/auth0/go-auth0/management"
)

func flattenEncryptionKeys(keys []*management.EncryptionKey) []interface{} {
	var result []interface{}
	const timeRFC3339WithMilliseconds = "2006-01-02T15:04:05.000Z07:00"

	for _, key := range keys {
		result = append(result, map[string]interface{}{
			"key_id":        key.GetKID(),
			"parent_key_id": key.GetParentKID(),
			"type":          key.GetType(),
			"state":         key.GetState(),
			"created_at":    key.GetCreatedAt().Format(timeRFC3339WithMilliseconds),
			"updated_at":    key.GetUpdatedAt().Format(timeRFC3339WithMilliseconds),
		})
	}
	return result
}
