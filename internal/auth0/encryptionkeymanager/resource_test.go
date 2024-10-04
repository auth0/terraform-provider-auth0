package encryptionkeymanager_test

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

const testAccEncryptionKeyManagerCreate = `
resource "auth0_encryption_key_manager" "my_key_manager" { }
`

const testAccEncryptionKeyManagerFirstRotation = `
resource "auth0_encryption_key_manager" "my_key_manager" {
	key_rotation_id = "initial_value"
}
`

const testAccEncryptionKeyManagerSecondRotation = `
resource "auth0_encryption_key_manager" "my_key_manager" {
	key_rotation_id = "changed_value"
}
`

const testAccEncryptionKeyManagerUnsetRotation = `
resource "auth0_encryption_key_manager" "my_key_manager" {
}
`

func TestAccEncryptionKeyManager(t *testing.T) {
	initialKey := make(map[string]string)
	firstRotationKey := make(map[string]string)
	secondRotationKey := make(map[string]string)
	unsetRotationKey := make(map[string]string)

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				Config: testAccEncryptionKeyManagerCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key_manager.my_key_manager", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", &initialKey),
					func(_ *terraform.State) error {
						keyID, ok := initialKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						parentKeyID, ok := initialKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, initialKey["type"], "tenant-master-key")
						assert.Equal(t, initialKey["state"], "active")
						createdAt, ok := initialKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := initialKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeyManagerFirstRotation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key_manager.my_key_manager", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", &firstRotationKey),
					func(_ *terraform.State) error {
						keyID, ok := firstRotationKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.NotEqual(t, firstRotationKey["key_id"], initialKey["key_id"])
						parentKeyID, ok := firstRotationKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, firstRotationKey["type"], "tenant-master-key")
						assert.Equal(t, firstRotationKey["state"], "active")
						createdAt, ok := firstRotationKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := firstRotationKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeyManagerSecondRotation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key_manager.my_key_manager", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", &secondRotationKey),
					func(_ *terraform.State) error {
						keyID, ok := secondRotationKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.NotEqual(t, secondRotationKey["key_id"], firstRotationKey["key_id"])
						parentKeyID, ok := secondRotationKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, secondRotationKey["type"], "tenant-master-key")
						assert.Equal(t, secondRotationKey["state"], "active")
						createdAt, ok := secondRotationKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := secondRotationKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeyManagerUnsetRotation,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("auth0_encryption_key_manager.my_key_manager", "encryption_keys.#", regexp.MustCompile("^[1-9][0-9]*")),
					extractActiveKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", &unsetRotationKey),
					func(_ *terraform.State) error {
						keyID, ok := unsetRotationKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.Equal(t, unsetRotationKey["key_id"], secondRotationKey["key_id"])
						parentKeyID, ok := unsetRotationKey["parent_key_id"]
						assert.True(t, ok && len(parentKeyID) > 0, "parent_key_id should exist")
						assert.Equal(t, unsetRotationKey["type"], "tenant-master-key")
						assert.Equal(t, unsetRotationKey["state"], "active")
						createdAt, ok := unsetRotationKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := unsetRotationKey["updated_at"]
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
