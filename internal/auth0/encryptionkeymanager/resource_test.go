package encryptionkeymanager_test

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"math"
	"os"
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
resource "auth0_encryption_key_manager" "my_key_manager" {
}
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

const testAccEncryptionKeyManagerCreateRootKey = `
resource "auth0_tenant" "my_tenant" {
}

resource "auth0_encryption_key_manager" "my_key_manager" {
	customer_provided_root_key {
	}
}
`

const testAccEncryptionKeyManagerRemoveRootKey = `
resource "auth0_tenant" "my_tenant" {
}

resource "auth0_encryption_key_manager" "my_key_manager" {
}
`

const testAccEncryptionKeyManagerAddWrappedKey = `
resource "auth0_tenant" "my_tenant" {
}

variable "WRAPPED_KEY" {
	type = string
}

resource "auth0_encryption_key_manager" "my_key_manager" {
	customer_provided_root_key {
		wrapped_key = var.WRAPPED_KEY
	}
}
`

func TestAccEncryptionKeyManagerRotation(t *testing.T) {
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
					extractKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", "active", &initialKey),
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
					extractKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", "active", &firstRotationKey),
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
					extractKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", "active", &secondRotationKey),
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
					extractKey("auth0_encryption_key_manager.my_key_manager", "encryption_keys", "tenant-master-key", "active", &unsetRotationKey),
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

func TestAccEncryptionKeyManagerCustomerProvidedRootKey(t *testing.T) {
	initialRootKey := make(map[string]string)
	initialWrappingKey := make(map[string]string)
	secondRootKey := make(map[string]string)
	secondWrappingKey := make(map[string]string)
	thirdRootKey := make(map[string]string)

	var wrappedKey string

	acctest.Test(t, resource.TestCase{
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					_ = os.Setenv("TF_VAR_WRAPPED_KEY", "bad-key-value")
				},
				Config:      testAccEncryptionKeyManagerAddWrappedKey,
				ExpectError: regexp.MustCompile(`Error: The wrapped_key attribute should not be specified`),
			},
			{
				Config: testAccEncryptionKeyManagerCreateRootKey,
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.#", "1"),
					extractRootKey("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key", &initialRootKey),
					extractWrappingKey("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key", &initialWrappingKey),
					func(_ *terraform.State) error {
						keyID, ok := initialRootKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.Equal(t, initialRootKey["type"], "customer-provided-root-key")
						assert.Equal(t, initialRootKey["state"], "pre-activation")
						createdAt, ok := initialRootKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := initialRootKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						publicWrappingKey, ok := initialWrappingKey["public_wrapping_key"]
						assert.True(t, ok && len(publicWrappingKey) > 0, "public_wrapping_key should exist")
						assert.Equal(t, initialWrappingKey["wrapping_algorithm"], "CKM_RSA_AES_KEY_WRAP")

						return nil
					},
				),
			},
			{
				Config: testAccEncryptionKeyManagerRemoveRootKey,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.#", "0"),
				),
			},
			{
				Config: testAccEncryptionKeyManagerCreateRootKey,
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.#", "1"),
					extractRootKey("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key", &secondRootKey),
					extractWrappingKey("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key", &secondWrappingKey),
					func(_ *terraform.State) error {
						keyID, ok := secondRootKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.NotEqual(t, initialRootKey["key_id"], secondRootKey["key_id"])
						assert.Equal(t, secondRootKey["state"], "pre-activation")
						createdAt, ok := secondRootKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := secondRootKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						publicWrappingKey, ok := secondWrappingKey["public_wrapping_key"]
						assert.True(t, ok && len(publicWrappingKey) > 0, "public_wrapping_key should exist")
						assert.Equal(t, secondWrappingKey["wrapping_algorithm"], "CKM_RSA_AES_KEY_WRAP")

						tmpWrappedKey, err := createAWSWrappedCiphertext(publicWrappingKey)
						if err != nil {
							return err
						}

						wrappedKey = tmpWrappedKey

						return nil
					},
				),
			},
			{
				PreConfig: func() {
					_ = os.Setenv("TF_VAR_WRAPPED_KEY", "bad-key-value")
				},
				Config:      testAccEncryptionKeyManagerAddWrappedKey,
				ExpectError: regexp.MustCompile(`Wrapped key material is invalid`),
			},
			{
				PreConfig: func() {
					_ = os.Setenv("TF_VAR_WRAPPED_KEY", wrappedKey)
				},
				Config: testAccEncryptionKeyManagerAddWrappedKey,
			},
			{
				RefreshState: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.#", "1"),
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.0.public_wrapping_key", ""),
					resource.TestCheckResourceAttr("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key.0.wrapping_algorithm", ""),
					extractRootKey("auth0_encryption_key_manager.my_key_manager", "customer_provided_root_key", &thirdRootKey),
					func(_ *terraform.State) error {
						keyID, ok := thirdRootKey["key_id"]
						assert.True(t, ok && len(keyID) > 0, "key_id should exist")
						assert.Equal(t, secondRootKey["key_id"], thirdRootKey["key_id"])
						assert.Equal(t, thirdRootKey["state"], "active")
						createdAt, ok := thirdRootKey["created_at"]
						assert.True(t, ok && len(createdAt) > 0, "created_at should exist")
						updatedAt, ok := thirdRootKey["updated_at"]
						assert.True(t, ok && len(updatedAt) > 0, "updated_at should exist")

						return nil
					},
				),
			},
		},
	})
}

func extractKey(resource, attribute, keyType, keyState string, keyMapPtr *map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		clear(*keyMapPtr)

		tfResource, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("extractKey: failed to find resource with name: %q", resource)
		}
		countValue, ok := tfResource.Primary.Attributes[fmt.Sprintf("%s.#", attribute)]
		if !ok {
			return fmt.Errorf("extractKey: failed to find attribute with name: %q", attribute)
		}
		count, err := strconv.Atoi(countValue)
		if err != nil {
			return err
		}
		for i := range count {
			stateValue, ok := tfResource.Primary.Attributes[keyName(attribute, i, "state")]
			if !ok {
				return fmt.Errorf("extractKey: failed to find state for attribute with name: %q", attribute)
			}
			if stateValue != keyState {
				continue
			}
			typeValue, ok := tfResource.Primary.Attributes[keyName(attribute, i, "type")]
			if !ok {
				return fmt.Errorf("extractKey: failed to find type for attribute with name: %q", attribute)
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
		return fmt.Errorf("extractKey: active key of type %q not found", keyType)
	}
}

func extractRootKey(resource, attribute string, keyMapPtr *map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		clear(*keyMapPtr)

		tfResource, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("extractRootKey: failed to find resource with name: %q", resource)
		}
		for key, value := range tfResource.Primary.Attributes {
			if strings.HasPrefix(key, keyName(attribute, 0, "")) {
				foundKey, _ := strings.CutPrefix(key, keyName(attribute, 0, ""))
				switch foundKey {
				case "key_id", "parent_key_id", "type", "state", "created_at", "updated_at":
					(*keyMapPtr)[foundKey] = value
				}
			}
		}
		return nil
	}
}

func extractWrappingKey(resource, attribute string, keyMapPtr *map[string]string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		clear(*keyMapPtr)

		tfResource, ok := state.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("extractWrappingKey: failed to find resource with name: %q", resource)
		}
		for key, value := range tfResource.Primary.Attributes {
			if strings.HasPrefix(key, keyName(attribute, 0, "")) {
				foundKey, _ := strings.CutPrefix(key, keyName(attribute, 0, ""))
				switch foundKey {
				case "public_wrapping_key", "wrapping_algorithm":
					(*keyMapPtr)[foundKey] = value
				}
			}
		}
		return nil
	}
}

func keyName(attribute string, index int, key string) string {
	return fmt.Sprintf("%s.%d.%s", attribute, index, key)
}

// Utility methods and constants for wrapping keys.

// Constants for wrapping sizes and parameters.
const (
	minWrapSize = 16
	maxWrapSize = 8192
	roundCount  = 6
	ivPrefix    = uint32(0xA65959A6)
)

// kwpImpl is a Key Wrapping with Padding implementation.
type kwpImpl struct {
	block cipher.Block
}

func createAWSWrappedCiphertext(publicKeyPEM string) (string, error) {
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode public key PEM")
	}

	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	publicRSAKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("public key is not of type *rsa.PublicKey")
	}

	ephemeralKey := make([]byte, 32)
	if _, err := rand.Read(ephemeralKey); err != nil {
		return "", fmt.Errorf("failed to generate ephemeral key: %w", err)
	}

	plaintextKey := make([]byte, 32)
	if _, err := rand.Read(plaintextKey); err != nil {
		return "", fmt.Errorf("failed to generate plaintext key: %w", err)
	}

	wrappedEphemeralKey, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicRSAKey, ephemeralKey, nil)
	if err != nil {
		return "", fmt.Errorf("failed to wrap ephemeral key: %w", err)
	}

	kwp, err := newKWP(ephemeralKey)
	if err != nil {
		return "", fmt.Errorf("failed to create KWP instance: %w", err)
	}

	wrappedTargetKey, err := kwp.wrap(plaintextKey)
	if err != nil {
		return "", fmt.Errorf("failed to wrap target key using KWP: %w", err)
	}

	wrappedEphemeralKey = append(wrappedEphemeralKey, wrappedTargetKey...)
	return base64.StdEncoding.EncodeToString(wrappedEphemeralKey), nil
}

func newKWP(wrappingKey []byte) (*kwpImpl, error) {
	switch len(wrappingKey) {
	case 16, 32:
		block, err := aes.NewCipher(wrappingKey)
		if err != nil {
			return nil, fmt.Errorf("kwp: error building AES cipher: %v", err)
		}
		return &kwpImpl{block: block}, nil
	default:
		return nil, fmt.Errorf("kwp: invalid AES key size; want 16 or 32, got %d", len(wrappingKey))
	}
}

func wrappingSize(inputSize int) int {
	paddingSize := 7 - (inputSize+7)%8
	return inputSize + paddingSize + 8
}

func (kwp *kwpImpl) computeW(iv, key []byte) ([]byte, error) {
	if len(key) <= 8 || len(key) > math.MaxInt32-16 || len(iv) != 8 {
		return nil, fmt.Errorf("kwp: computeW called with invalid parameters")
	}

	data := make([]byte, wrappingSize(len(key)))
	copy(data, iv)
	copy(data[8:], key)
	blockCount := len(data)/8 - 1

	buf := make([]byte, 16)
	copy(buf, data[:8])

	for i := 0; i < roundCount; i++ {
		for j := 0; j < blockCount; j++ {
			copy(buf[8:], data[8*(j+1):])
			kwp.block.Encrypt(buf, buf)

			roundConst := uint(i*blockCount + j + 1)
			for b := 0; b < 4; b++ {
				buf[7-b] ^= byte(roundConst & 0xFF)
				roundConst >>= 8
			}

			copy(data[8*(j+1):], buf[8:])
		}
	}
	copy(data[:8], buf)
	return data, nil
}

func (kwp *kwpImpl) wrap(data []byte) ([]byte, error) {
	if len(data) < minWrapSize {
		return nil, fmt.Errorf("kwp: key size to wrap too small")
	}
	if len(data) > maxWrapSize {
		return nil, fmt.Errorf("kwp: key size to wrap too large")
	}

	iv := make([]byte, 8)
	binary.BigEndian.PutUint32(iv, ivPrefix)
	binary.BigEndian.PutUint32(iv[4:], uint32(len(data)))

	return kwp.computeW(iv, data)
}
