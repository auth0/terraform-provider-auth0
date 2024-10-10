package encryptionkeymanager

import (
	"github.com/auth0/go-auth0/management"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func flattenCustomerProvidedRootKey(rootKey *management.EncryptionKey, wrappingKey *management.WrappingKey, wrappedKey *string, rootKeyResponse customerProvidedRootKeyModel) customerProvidedRootKeyModel {
	if rootKey != nil {
		rootKeyResponse.KeyID = types.StringValue(rootKey.GetKID())
		rootKeyResponse.ParentKeyID = types.StringValue(rootKey.GetParentKID())
		rootKeyResponse.Type = types.StringValue(rootKey.GetType())
		rootKeyResponse.State = types.StringValue(rootKey.GetState())
		rootKeyResponse.CreatedAt = timetypes.NewRFC3339TimeValue(rootKey.GetCreatedAt())
		rootKeyResponse.UpdatedAt = timetypes.NewRFC3339TimeValue(rootKey.GetUpdatedAt())
		if rootKey.GetState() != "pre-activation" {
			rootKeyResponse.PublicWrappingKey = types.StringNull()
			rootKeyResponse.WrappingAlgorithm = types.StringNull()
		}
	}
	if wrappingKey != nil {
		rootKeyResponse.PublicWrappingKey = types.StringValue(wrappingKey.GetPublicKey())
		rootKeyResponse.WrappingAlgorithm = types.StringValue(wrappingKey.GetAlgorithm())
	}
	if wrappedKey != nil {
		rootKeyResponse.WrappedKey = types.StringValue(*wrappedKey)
	}

	return rootKeyResponse
}

func flattenEncryptionKeys(keys []*management.EncryptionKey) []encryptionKeyModel {
	flattenedKeys := make([]encryptionKeyModel, 0, len(keys))
	for _, key := range keys {
		flattenedKeys = append(flattenedKeys, flattenKey(key))
	}

	return flattenedKeys
}

func flattenKey(key *management.EncryptionKey) encryptionKeyModel {
	return encryptionKeyModel{
		KeyID:       types.StringValue(key.GetKID()),
		ParentKeyID: types.StringValue(key.GetParentKID()),
		Type:        types.StringValue(key.GetType()),
		State:       types.StringValue(key.GetState()),
		CreatedAt:   timetypes.NewRFC3339TimeValue(key.GetCreatedAt()),
		UpdatedAt:   timetypes.NewRFC3339TimeValue(key.GetUpdatedAt()),
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
