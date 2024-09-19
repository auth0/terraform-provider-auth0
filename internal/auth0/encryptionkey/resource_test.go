package encryptionkey_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/stretchr/testify/assert"

	"github.com/auth0/terraform-provider-auth0/internal/acctest"
)

const testAccEncryptionKeysCreate = `
resource "auth0_encryption_key" "my_keys" { }
`

const testAccEncryptionKeysRekey = `
resource "auth0_encryption_key" "my_keys" {
	rekey = true
}
`

func TestAccEncryptionKeys(t *testing.T) {
	oldKey := make(map[string]string)
	newKey := make(map[string]string)
	newerKey := make(map[string]string)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptionKeysCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key.my_keys", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key.my_keys", "encryption_keys", "tenant-master-key", &oldKey),
					func(_ *terraform.State) error {
						keyID, ok := oldKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						parentKeyID, ok := oldKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, oldKey["type"], "tenant-master-key")
						assert.Equal(t, oldKey["state"], "active")
						createdAt, ok := oldKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := oldKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeysRekey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key.my_keys", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key.my_keys", "encryption_keys", "tenant-master-key", &newKey),
					func(_ *terraform.State) error {
						keyID, ok := newKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.NotEqual(t, newKey["key_id"], oldKey["key_id"])
						parentKeyID, ok := newKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, newKey["type"], "tenant-master-key")
						assert.Equal(t, newKey["state"], "active")
						createdAt, ok := newKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := newKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeysCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key.my_keys", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key.my_keys", "encryption_keys", "tenant-master-key", &newerKey),
					func(_ *terraform.State) error {
						keyID, ok := newerKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.Equal(t, newerKey["key_id"], newKey["key_id"])
						parentKeyID, ok := newerKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, newerKey["type"], "tenant-master-key")
						assert.Equal(t, newerKey["state"], "active")
						createdAt, ok := newerKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := newerKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
		},
	})
}

func extractActiveKey(resource, attribute, keyType string, keyMapPtr *map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		clear(*keyMapPtr)

		tfResource, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("extractActiveKey: failed to find resource with name: %q", resource)
		}
		countValue, ok := tfResource.Primary.Attributes[fmt.Sprintf("%s.#", attribute)]
		if !ok {
			return fmt.Errorf("extractActiveKey: failed to find attribute with name: %q", attribute)
		}
		count, err := strconv.Atoi(countValue)
		if err != nil {
			return err
		}
		fmt.Printf("DEBUG: CRAIG: extract count: %d\n", count)
		for i := range count {
			stateValue, ok := tfResource.Primary.Attributes[keyName(attribute, i, "state")]
			if !ok {
				return fmt.Errorf("extractActiveKey: failed to find state for attribute with name: %q", attribute)
			}
			if stateValue != "active" {
				continue
			}
			typeValue, ok := tfResource.Primary.Attributes[keyName(attribute, i, "type")]
			if !ok {
				return fmt.Errorf("extractActiveKey: failed to find type for attribute with name: %q", attribute)
			}
			if typeValue != keyType {
				continue
			}
			for key, value := range tfResource.Primary.Attributes {
				if strings.HasPrefix(key, keyName(attribute, i, "")) {
					foundKey, _ := strings.CutPrefix(key, keyName(attribute, i, ""))
					(*keyMapPtr)[foundKey] = value
				}
			}
			return nil
		}
		return fmt.Errorf("extractActiveKey: active key of type %q not found", keyType)
	}
}

func keyName(attribute string, index int, key string) string {
	return fmt.Sprintf("%s.%d.%s", attribute, index, key)
}
