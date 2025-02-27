package encryptionkeymanager

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func flattenCustomerProvidedRootKey(data *schema.ResourceData, rootKey *management.EncryptionKey, wrappingKey *management.WrappingKey) []interface{} {
	const timeRFC3339WithMilliseconds = "2006-01-02T15:04:05.000Z07:00"

	originalValue := data.Get("customer_provided_root_key").([]interface{})
	result := make(map[string]interface{})
	if len(originalValue) > 0 && originalValue[0] != nil {
		result = originalValue[0].(map[string]interface{})
	}
	if rootKey != nil {
		result["key_id"] = rootKey.GetKID()
		result["parent_key_id"] = rootKey.GetParentKID()
		result["type"] = rootKey.GetType()
		result["state"] = rootKey.GetState()
		result["created_at"] = rootKey.GetCreatedAt().Format(timeRFC3339WithMilliseconds)
		result["updated_at"] = rootKey.GetUpdatedAt().Format(timeRFC3339WithMilliseconds)
		if rootKey.GetState() != "pre-activation" {
			result["public_wrapping_key"] = nil
			result["wrapping_algorithm"] = nil
		}
	}
	if wrappingKey != nil {
		result["public_wrapping_key"] = wrappingKey.GetPublicKey()
		result["wrapping_algorithm"] = wrappingKey.GetAlgorithm()
	}

	return []interface{}{result}
}

func flattenEncryptionKeys(keys []*management.EncryptionKey) []interface{} {
	var flattenedKeys []interface{}
	for _, key := range keys {
		flattenedKeys = append(flattenedKeys, flattenKey(key))
	}

	return flattenedKeys
}

func flattenKey(key *management.EncryptionKey) interface{} {
	const timeRFC3339WithMilliseconds = "2006-01-02T15:04:05.000Z07:00"

	return map[string]interface{}{
		"key_id":        key.GetKID(),
		"parent_key_id": key.GetParentKID(),
		"type":          key.GetType(),
		"state":         key.GetState(),
		"created_at":    key.GetCreatedAt().Format(timeRFC3339WithMilliseconds),
		"updated_at":    key.GetUpdatedAt().Format(timeRFC3339WithMilliseconds),
	}
}

func getKeyByTypeAndState(keyType, keyState string, keys []*management.EncryptionKey) *management.EncryptionKey {
	for _, key := range keys {
		if key.GetType() == keyType && key.GetState() == keyState {
			return key
		}
	}
	return nil
}
